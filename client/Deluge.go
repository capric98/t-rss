package client

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"unicode"

	"github.com/gdm85/go-rencode"
)

const (
	none     = 0
	wTimeout = 10
	timeoutA = 1000 //(ms)
	rpcResp  = 1
	rpcError = 2
	rpcEvent = 3
)

// DeType :)
type DeType struct {
	settings    map[string]interface{}
	host        string
	name, label string
	user, pass  string
	version     int
	protoVer    int
	rttx4       time.Duration
}

type reqIDType struct {
	count int
	mu    sync.Mutex
}

var (
	reqID    = &reqIDType{}
	paraList = []string{"add_paused", "auto_managed",
		"download_location", "max_connections", "max_download_speed",
		"max_upload_speed", "move_completed", "move_completed_path",
		"pre_allocated_storage", "prioritize_first_last_pieces",
		"remove_at_ratio", "seed_mode", "sequential_download",
		"shared", "stop_at_ratio", "stop_ratio", "super_seeding"}
	//https://github.com/deluge-torrent/deluge/blob/4b29436cd5eabf9af271f3fa6250cd7c91cdbc9d/deluge/core/torrent.py#L133

	// ErrExpectDHeader :)
	ErrExpectDHeader = errors.New("Expected D as first byte in reply")
	// ErrExpectPVHeader :)
	ErrExpectPVHeader = errors.New("Expected protocal version as first byte in reply")
	// ErrRespIncomplete :)
	ErrRespIncomplete = errors.New("Expected a longer response than actually got")
	// ErrUnknownResp :)
	ErrUnknownResp = errors.New("Unknown RPC response")
	// ErrRPCEvent :)
	ErrRPCEvent = errors.New("Unexpected RPC Event message")
	// ErrAddFail :)
	ErrAddFail = errors.New("Failed to add torrent file after 3 tries")
)

// Add :)
func (c *DeType) Add(data []byte, name string) (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()

	b64 := base64.StdEncoding.EncodeToString(data)

	for try := 0; try < 3; try++ {
		if nil == func() (e error) {
			conn, err := c.newConn()
			if err == nil {
				defer conn.Close()
				if e = c.login(conn); e != nil {
					return
				}
				if e = c.call("core.add_torrent_file", makeList(name, b64, makeDict(c.settings)), makeDict(nil), conn); e != nil {
					// c.call("core.add_torrent_file", makeList(name, b64), makeDict(c.settings), conn)
					// └-> THIS WOULD NOT WORK!!!!!!!
					//     Thank you Deluge!
					return
				}
				return c.recvResp(conn)
			}
			return err
		}() {
			return nil
		}

	}

	return ErrAddFail
}

// NewDeClient :)
func NewDeClient(key string, m map[string]interface{}) *DeType {
	defer func() {
		if p := recover(); p != nil {
			log.Fatal("new deluge client:", p)
		}
	}()

	var nc = &DeType{
		name:     "Deluge",
		label:    key,
		settings: make(map[string]interface{}),
		user:     m["username"].(string),
		pass:     m["password"].(string),
	}

	for _, para := range paraList {
		if m[para] != nil {
			nc.settings[para] = m[para]
		}
	}
	if nc.settings["max_download_speed"] != nil {
		nc.settings["max_download_speed"] = parseSpeed(nc.settings["max_download_speed"])
	}
	if nc.settings["max_upload_speed"] != nil {
		nc.settings["max_upload_speed"] = parseSpeed(nc.settings["max_upload_speed"])
	}

	if m["host"] == nil {
		log.Panicln("Deluge: miss host.")
	}
	nc.host = m["host"].(string)

	var failcount int
	var err error
	var conn *tls.Conn
	for {
		if conn, err = nc.init(); err == nil {
			_ = conn.Close()
			break
		}
		failcount++
		if failcount == 3 {
			log.Fatal("Init deluge client:", err)
		}
	}
	return nc
}

func checkConn(c *tls.Conn, e error) *tls.Conn {
	if e != nil {
		log.Panic(e)
	}
	return c
}

func (c *DeType) init() (conn *tls.Conn, e error) {
	defer func() {
		if p := recover(); p != nil {
			e = fmt.Errorf("%v", p)
		}
	}()

	conn = checkConn(c.newConn())
	conn = checkConn(c.detectVersion(conn))
	//log.Println("Deluge client init with error", e)
	//log.Println("Deluge version:", c.version)
	//log.Println("Protocal version:", c.protoVer)
	return conn, c.login(conn)
}

func (c *DeType) login(conn *tls.Conn) (e error) {
	m := make(map[string]interface{})
	m["client_version"] = "deluge-client"
	dict := makeDict(m)
	list := makeList(c.user, c.pass)

	switch c.version {
	case 1:
		e = c.call("daemon.login", list, makeDict(nil), conn)
	case 2:
		e = c.call("daemon.login", list, dict, conn)
	}

	if e != nil {
		return
	}

	return c.recvResp(conn)
}

func (c *DeType) call(method string, args rencode.List, kargs rencode.Dictionary, conn *tls.Conn) (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()
	return c.sendCall(c.version, c.protoVer, method, args, kargs, conn)
}

func (c *DeType) sendCall(version int, protoVer int, method string, args rencode.List, kargs rencode.Dictionary, conn *tls.Conn) error {
	rID := reqID.next()
	var b, z, req bytes.Buffer

	e := rencode.NewEncoder(&b)
	if err := e.Encode(makeList(makeList(rID, method, args, kargs))); err != nil {
		return err
	}

	wzlib := zlib.NewWriter(&z)
	_, _ = wzlib.Write(b.Bytes())
	wzlib.Close()

	if version == 2 {
		// need to send a header to client
		switch protoVer {
		case none:
			req.WriteRune('D')
			_ = binary.Write(&req, binary.BigEndian, int32(z.Len()))
		case 1:
			_ = binary.Write(&req, binary.BigEndian, uint8(protoVer))
			_ = binary.Write(&req, binary.BigEndian, uint32(z.Len()))
		}
	}

	_, _ = req.Write(z.Bytes())

	_ = conn.SetDeadline(time.Now().Add(wTimeout * time.Second))

	if _, err := conn.Write(req.Bytes()); err != nil {
		return err
	}

	return nil
}

func (c *DeType) detectVersion(conn *tls.Conn) (*tls.Conn, error) {
	sign := make([]byte, 1)

	now := time.Now()
	_ = c.sendCall(1, none, "daemon.info", makeList(), makeDict(nil), conn)
	_ = c.sendCall(2, none, "daemon.info", makeList(), makeDict(nil), conn)
	_ = c.sendCall(2, 1, "daemon.info", makeList(), makeDict(nil), conn)

	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))
	_, err := conn.Read(sign)

	c.rttx4 = time.Since(now) + (timeoutA * time.Millisecond)

	if err != nil {
		return nil, err
	}

	if sign[0] == byte('D') {
		c.version = 2
		c.protoVer = none
	} else if sign[0] == 1 {
		c.version = 2
		c.protoVer = 1
	} else {
		c.version = 1
		c.protoVer = none
		//Deluge 1 doesn't recover well from the bad request. Re-connect!
		conn.Close()
		return c.newConn()
	}

	return conn, nil
}

func (c *DeType) recvResp(conn *tls.Conn) (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()

	var buf bytes.Buffer
	var zr io.Reader

	switch c.version {
	case 1:
		for {
			_ = conn.SetDeadline(time.Now().Add(c.rttx4))
			if n, _ := io.Copy(&buf, conn); n == 0 {
				break
			}
		}
	case 2:
		var sign bytes.Buffer
		var expectLen int

		_ = conn.SetDeadline(time.Now().Add(c.rttx4))
		if _, err := io.CopyN(&sign, conn, 5); err != nil {
			return err
		}

		switch c.protoVer {
		case none:
			if (sign.Bytes())[0] != byte('D') {
				return ErrExpectDHeader
			}
			if err := binary.Read(bytes.NewReader((sign.Bytes())[1:5]), binary.BigEndian, &expectLen); err != nil {
				return err
			}
		case 1:
			if (sign.Bytes())[0] != 1 {
				return ErrExpectPVHeader
			}
			expectLen = int(binary.BigEndian.Uint32((sign.Bytes())[1:5]))
		}
		_ = conn.SetDeadline(time.Now().Add(2 * c.rttx4))
		if _, err := io.CopyN(&buf, conn, int64(expectLen)); err != nil {
			return ErrRespIncomplete
		}
	}

	resp := buf.Bytes()
	zr, e = zlib.NewReader(bytes.NewReader(resp))
	if e != nil {
		return
	}

	r := rencode.NewDecoder(zr)
	rli, err := r.DecodeNext()
	if err != nil {
		return err
	}
	rlist := rli.(rencode.List)
	rValue := rlist.Values()
	msgType := convertInt(rValue[0])
	switch msgType {
	case rpcResp:
	case rpcError:
		errorlist := rValue[2].(rencode.List)
		errs := errorlist.Values()
		msg := string(errs[0].([]uint8)) + "\n" + string(errs[1].([]uint8)) + "\n" + string(errs[2].([]uint8))
		e = errors.New("rpcError with message:\n" + msg)
	case rpcEvent:
		e = ErrRPCEvent
	default:
		e = ErrUnknownResp
	}
	return
}

func (c *DeType) newConn() (conn *tls.Conn, e error) {
	d := net.Dialer{Timeout: 10 * time.Second}
	conn, e = tls.DialWithDialer(&d, "tcp", c.host, &tls.Config{
		InsecureSkipVerify: true,
	})
	return
}

func (r *reqIDType) next() (rid int) {
	r.mu.Lock()
	r.count++
	rid = r.count
	r.mu.Unlock()
	return
}

func makeList(args ...interface{}) rencode.List {
	list := rencode.NewList()
	for _, v := range args {
		list.Add(v)
	}
	return list
}

func makeDict(args map[string]interface{}) rencode.Dictionary {
	var dict rencode.Dictionary
	for k, v := range args {
		dict.Add(k, v)
	}
	return dict
}

func convertInt(i interface{}) int {
	switch i := i.(type) {
	case int8:
		return int(i)
	case int16:
		return int(i)
	case int32:
		return int(i)
	case int64:
		return int(i)
	case int:
		return i
	default:
		return -1
	}
}

func parseSpeed(v interface{}) float32 {
	if v == nil {
		return -1
	}
	switch v := v.(type) {
	case int:
		return float32(v)
	case string:
		var spNum float32
		u := make([]rune, 0)
		for _, c := range v {
			if !unicode.IsDigit(c) {
				u = append(u, c)
			} else {
				spNum = spNum*10 + float32(c-'0')
			}
		}
		unit := string(u)
		switch {
		case unit == "K" || unit == "k" || unit == "KB" || unit == "kB" || unit == "KiB" || unit == "kiB":
			spNum = spNum * 1024
		case unit == "M" || unit == "m" || unit == "MB" || unit == "mB" || unit == "MiB" || unit == "miB":
			spNum = spNum * 1024 * 1024
		case unit == "G" || unit == "g" || unit == "GB" || unit == "gB" || unit == "GiB" || unit == "giB":
			spNum = spNum * 1024 * 1024 * 1024
		}
		return spNum
	default:
		return -1
	}
}
