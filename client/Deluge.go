package client

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gdm85/go-rencode"
)

type DeType struct {
	client   *tls.Conn
	settings map[string]interface{}
	name     string
	label    string
	version  int
	protoVer int
	mu       sync.Mutex
}

type reqIDType struct {
	count int
	mu    sync.Mutex
}

var (
	reqID    = &reqIDType{}
	paraList = []string{}
)

func (c *DeType) Add(data []byte, name string) (e error) {
	var try int
	b64 := base64.StdEncoding.EncodeToString(data)

	for try < 3 {
		e = c.call("core.add_torrent_file", makeList(name, b64), makeDict(c.settings))
		if e != nil {
			c.reconnect()
			continue
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
		// Speed!
		nc.settings[para] = m[para]
	}

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

	e = c.detectVersion()

	return
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
	//var reql rencode.List

	e := rencode.NewEncoder(&b)
	//reql = rencode.NewList(rID, method, args, kargs)
	if err := e.Encode(makeList(makeList(rID, method, args, kargs))); err != nil {
		return err
	}

	wzlib := zlib.NewWriter(&z)
	_, _ = wzlib.Write(b.Bytes())
	wzlib.Close()

	if version == 2 {
		// need to send a header to client
		switch protoVer {
		case -1:
			req.WriteRune('D')
			_ = binary.Write(&req, binary.BigEndian, int32(z.Len()))
		case 1:
			_ = binary.Write(&req, binary.BigEndian, uint8(protoVer))
			_ = binary.Write(&req, binary.BigEndian, uint32(z.Len()))
		}
	}

	_, _ = req.Write(z.Bytes())
	// fmt.Printf("%+q\n", z.String())

	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.client.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.client.Write(req.Bytes()); err != nil {
		return err
	}

	return nil
}

func (c *DeType) detectVersion() error {
	_ = c.sendCall(1, -1, "daemon.info", makeList(), makeDict(nil))
	_ = c.sendCall(2, -1, "daemon.info", makeList(), makeDict(nil))
	_ = c.sendCall(2, 1, "daemon.info", makeList(), makeDict(nil))
	var buf bytes.Buffer

	time.Sleep(100 * time.Millisecond)

	//rlock.Lock()
	//_, err := io.CopyN(&buf, bufreader, 1)
	//rlock.Unlock()

	resp := buf.Bytes()
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	if resp[0] == byte('D') {
		c.version = 2
		c.protoVer = -1
		// receive daemon_version!!!
	} else if resp[0] == 1 {
		c.version = 2
		c.protoVer = 1
		// receive daemon_version!!!
	} else {
		c.version = 1
		//Deluge 1 doesn't recover well from the bad request. Re-connect!
		if err := c.reconnect(); err != nil {
			return err
		}
	}
	fmt.Println("Deluge version:", c.version)
	fmt.Println("Protocal version:", c.protoVer)

	return nil
}

// func (c *DeType) recvResp() (*rencode.Decoder, error) {
// 	var buf bytes.Buffer
// 	var body bytes.Buffer
// 	var r io.ReadCloser
// 	w := bufio.NewWriter(&buf)
// 	bw := bufio.NewWriter(&body)

// 	time.Sleep(time.Duration(wait) * time.Millisecond)
// 	fmt.Println("rbuf remains", rbuf.Len())

// 	if c.version == 2 {
// 		header, _ := bufreader.ReadByte()
// 		switch c.protocolVersion {
// 		case -1:
// 			if header != 'D' {
// 				return nil, fmt.Errorf("Expected D as first byte in reply.")
// 			}
// 			var n int32
// 			if err := binary.Read(bufreader, binary.BigEndian, &n); err != nil {
// 				return nil, err
// 			}
// 			if _, err := io.CopyN(w, bufreader, int64(n)); err != nil {
// 				return nil, err
// 			}

// 		case 1:
// 			if header != uint8(1) {
// 				return nil, fmt.Errorf("Expected protocol version %d as first byte in reply", 1)
// 			}
// 			var n uint32
// 			if err := binary.Read(bufreader, binary.BigEndian, &n); err != nil {
// 				return nil, err
// 			}
// 			if _, err := io.CopyN(w, bufreader, int64(n)); err != nil {
// 				return nil, err
// 			}
// 		}
// 		tr, err := zlib.NewReader(&buf)
// 		if err != nil {
// 			return nil, err
// 		}
// 		r = tr
// 	} else {
// 		tmp := rbuf.Bytes()
// 		rtmp := bytes.NewReader(tmp)
// 		tr, err := zlib.NewReader(rtmp)
// 		if err != nil {
// 			return nil, err
// 		}
// 		r = tr
// 	}

// 	_, _ = io.Copy(bw, r)
// 	r.Close()

// 	rbuf.Truncate(0)
// 	return rencode.NewDecoder(&body), nil
// }

func (c *DeType) reconnect() error {
	if err := c.closeSocket(); err != nil {
		return err
	}

	nconn, err := tls.Dial("tcp", c.settings["host"].(string), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	c.client = nconn

	return nil
}

func (c *DeType) closeSocket() error {
	if err := c.client.Close(); err != nil {
		return err
	}
	return nil
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
