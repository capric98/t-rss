package client

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"unicode"

	"github.com/gdm85/go-rencode"
)

const (
	None     = 0
	WTimeout = 10
	rpcResp  = 1
	rpcError = 2
	rpcEvent = 3
)

type DeType struct {
	client      *tls.Conn
	settings    map[string]interface{}
	host        string
	name, label string
	user, pass  string
	version     int
	protoVer    int
	rttx4       time.Duration
	mu          sync.Mutex
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
	ErrExpectDHeader  = errors.New("Expected D as first byte in reply.")
	ErrExpectPVHeader = errors.New("Expected protocal version as first byte in reply.")
	ErrRespIncomplete = errors.New("Expected a longer response than actually got.")
	ErrUnknownResp    = errors.New("Unknown RPC response.")
	ErrRPCEvent       = errors.New("Unexpected RPC Event message.")
)

func (c *DeType) Add(data []byte, name string) (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()

	var try int
	b64 := base64.StdEncoding.EncodeToString(data)

	for try < 3 {
		e = c.call("core.add_torrent_file", makeList(name, b64), makeDict(c.settings))
		if e != nil {
			_ = c.init()
			try++
			continue
		}

		e = c.recvResp()
		if e == nil {
			return
		}
	}
	return
}

func (c *DeType) Name() string {
	return c.name
}

func (c *DeType) Label() string {
	return c.label
}

func NewDeClient(key string, m map[string]interface{}) *DeType {
	var nc = &DeType{
		client:   nil,
		name:     "Deluge",
		label:    key,
		settings: make(map[string]interface{}),
	}

	for _, para := range paraList {
		if m[para] != nil {
			nc.settings[para] = m[para]
		}
	}
	nc.settings["max_download_speed"] = parseSpeed(nc.settings["max_download_speed"])
	nc.settings["max_upload_speed"] = parseSpeed(nc.settings["max_upload_speed"])

	if m["host"] == nil {
		log.Panicln("Deluge: miss host.")
	}
	nc.host = m["host"].(string)

	var failcount int
	var err error
	for {
		if err = nc.init(); err == nil {
			break
		}
		failcount++
		if failcount == 3 {
			log.Fatal(err)
		}
	}
	return nc
}

func (c *DeType) init() (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()

	if c.client == nil {
		e = c.newConn()
	} else {
		e = c.reconnect()
	}

	e = c.detectVersion()
	log.Println("Deluge client init:", e)
	log.Println("Deluge version:", c.version)
	log.Println("Protocal version:", c.protoVer)

	m := make(map[string]interface{})
	m["client_version"] = "deluge-client"
	dict := makeDict(m)
	list := makeList(c.user, c.pass)

	switch c.version {
	case 1:
		e = c.call("daemon.login", list, makeDict(nil))
	case 2:
		e = c.call("daemon.login", list, dict)
	}

	return c.recvResp()
}

func (c *DeType) call(method string, args rencode.List, kargs rencode.Dictionary) (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()
	return c.sendCall(c.version, c.protoVer, method, args, kargs)
}

func (c *DeType) sendCall(version int, protoVer int, method string, args rencode.List, kargs rencode.Dictionary) error {
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
		case None:
			req.WriteRune('D')
			_ = binary.Write(&req, binary.BigEndian, int32(z.Len()))
		case 1:
			_ = binary.Write(&req, binary.BigEndian, uint8(protoVer))
			_ = binary.Write(&req, binary.BigEndian, uint32(z.Len()))
		}
	}

	_, _ = req.Write(z.Bytes())

	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.client.SetWriteDeadline(time.Now().Add(WTimeout * time.Second))

	if _, err := c.client.Write(req.Bytes()); err != nil {
		return err
	}

	return nil
}

func (c *DeType) detectVersion() error {
	sign := make([]byte, 1)

	c.mu.Lock()
	now := time.Now()
	_ = c.sendCall(1, None, "daemon.info", makeList(), makeDict(nil))
	_ = c.sendCall(2, None, "daemon.info", makeList(), makeDict(nil))
	_ = c.sendCall(2, 1, "daemon.info", makeList(), makeDict(nil))

	_ = c.client.SetDeadline(time.Now().Add(1 * time.Second))
	_, err := c.client.Read(sign)

	c.rttx4 = time.Since(now)
	c.mu.Unlock()

	if err != nil {
		return err
	}

	defer func() {
		garbage := make([]byte, 1)
		_ = c.client.SetDeadline(time.Now().Add(c.rttx4))
		_, e := c.client.Read(garbage)
		for e == nil {
			_ = c.client.SetDeadline(time.Now().Add(c.rttx4))
			_, e = c.client.Read(garbage)
		}
	}() // Clean TCP buf.

	if sign[0] == byte('D') {
		c.version = 2
		c.protoVer = None
	} else if sign[0] == 1 {
		c.version = 2
		c.protoVer = 1
	} else {
		c.version = 1
		c.protoVer = None
		//Deluge 1 doesn't recover well from the bad request. Re-connect!
		if err := c.reconnect(); err != nil {
			return err
		}
	}

	return nil
}

func (c *DeType) recvResp() (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()

	var buf bytes.Buffer

	c.mu.Lock()
	for {
		_ = c.client.SetDeadline(time.Now().Add(c.rttx4))
		if n, _ := io.Copy(&buf, c.client); n == 0 {
			break
		}
	}
	c.mu.Unlock()

	resp := buf.Bytes()
	var zr io.Reader

	switch c.version {
	case 1:
		zr, e = zlib.NewReader(bytes.NewReader(resp))
		if e != nil {
			return
		}
	case 2:
		zr, e = zlib.NewReader(bytes.NewReader(resp[5:]))
		if e != nil {
			return
		}

		// Check validaty.
		var expectLen int
		switch c.protoVer {
		case None:
			if resp[0] != byte('D') {
				return ErrExpectDHeader
			}
			if err := binary.Read(bytes.NewReader(resp[1:5]), binary.BigEndian, &expectLen); err != nil {
				return err
			}
		case 1:
			if resp[0] != 1 {
				return ErrExpectPVHeader
			}
			expectLen = int(binary.BigEndian.Uint32(resp[1:5]))
		}
		if len(resp) < expectLen {
			return ErrRespIncomplete
		}
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
		e = errors.New("rpcError - Type: " + rValue[2].(string) + " & Message: " + rValue[3].(string))
	case rpcEvent:
		e = ErrRPCEvent
	default:
		e = ErrUnknownResp
	}
	return
}

func (c *DeType) newConn() (e error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	d := net.Dialer{Timeout: 10 * time.Second}
	c.client, e = tls.DialWithDialer(&d, "tcp", c.host, &tls.Config{
		InsecureSkipVerify: true,
	})
	return
}

func (c *DeType) reconnect() error {
	if err := c.client.Close(); err != nil {
		return err
	}
	return c.newConn()
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
	switch i.(type) {
	case int8:
		return int(i.(int8))
	case int16:
		return int(i.(int16))
	case int32:
		return int(i.(int32))
	case int64:
		return int(i.(int64))
	case int:
		return i.(int)
	default:
		return -1
	}
}

func parseSpeed(v interface{}) float32 {
	if v == nil {
		return -1
	}
	switch v.(type) {
	case int:
		return float32(v.(int))
	case string:
		var spNum float32
		u := make([]rune, 0)
		for _, c := range v.(string) {
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
