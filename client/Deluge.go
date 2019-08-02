package client

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gdm85/go-rencode"
)

var (
	request_id = 0
	glock      sync.Mutex
	rlock      sync.Mutex
	wlock      sync.Mutex
	rbuf       bytes.Buffer
)

type DeType struct {
	Name            string
	settings        map[string]string
	client          *tls.Conn
	version         int
	protocolVersion int // -1 for None
}

func (c DeType) newReqID() int {
	glock.Lock()
	request_id++
	defer glock.Unlock()
	return request_id
}

func NewDeClient(m map[interface{}]interface{}) ClientType {
	var nc ClientType
	nc.Name = "Deluge"
	nc.Client = DeType{
		client:   nil,
		settings: make(map[string]string),
	}
	for k, v := range m {
		nc.Client.(DeType).settings[k.(string)] = v.(string)
	} // Copy settings.

	// nc.Client.(DeType).settings["dlLimit"] = speedParse(nc.Client.(DeType).settings["dlLimit"])
	// nc.Client.(DeType).settings["upLimit"] = speedParse(nc.Client.(DeType).settings["upLimit"])

	fcount := 1
	fmt.Println("Fuck u man!")
	err := nc.Client.(DeType).Init()
	for err != nil {
		fmt.Println(err)
		fcount++
		if fcount == 3 {
			log.Fatal(err)
		}
		err = nc.Client.(DeType).Init()
	}
	return nc
}

func (c DeType) Init() error {
	conn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
		InsecureSkipVerify: true,
	}) // Deluge use self-signed cert and seems must to use it...
	fmt.Println("Dial success!")
	if err != nil {
		return err
	}
	c.client = conn
	c.newReqID()
	if err := c.detectVersion(); err != nil {
		return err
	}
	fmt.Println("Detect version success!")
	if c.version == 2 {
		if err := c.Call("daemon.login", c.settings["username"], c.settings["password"], "deluge-client"); err != nil {
			return err
		}
		d, err := c.recvResp()
		if err != nil {
			return err
		}
		for {
			i, e := d.DecodeNext()
			if e == io.EOF {
				break
			}
			fmt.Println(i)
		}
	} else {
		if err := c.Call("daemon.login", c.settings["username"], c.settings["password"]); err != nil {
			return err
		}
		d, err := c.recvResp()
		if err != nil {
			return err
		}
		for {
			i, e := d.DecodeNext()
			if e == io.EOF {
				break
			}
			fmt.Println(i)
		}
	}
	return nil
}

func (c DeType) Reconnect() error {
	c.client.Close()
	conn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	c.client = conn
	return nil
}

func (c DeType) Call(method string, args ...interface{}) error {
	return c.sendCall(c.version, c.protocolVersion, method, args...)
}

func (c DeType) sendCall(deVer int, protoVer int, method string, args ...interface{}) error {
	reqID := c.newReqID()
	var b, z bytes.Buffer
	e := rencode.NewEncoder(&b)
	reql := rencode.NewList(reqID, method)
	for _, v := range args {
		reql.Add(v)
	}
	if err := e.Encode(rencode.NewList(reql)); err != nil {
		return err
	}
	tmp := zlib.NewWriter(&z)
	_, _ = tmp.Write(b.Bytes())
	tmp.Close()

	wlock.Lock()
	_ = c.client.SetDeadline(time.Now().Add(10 * time.Second))
	if deVer == 2 {
		var req bytes.Buffer
		switch protoVer {
		case -1:
			req.WriteRune('D')
			_ = binary.Write(&req, binary.BigEndian, int32(z.Len()))
			_, _ = req.Write(z.Bytes())
			fmt.Println(req.String())
			if _, err := c.client.Write(req.Bytes()); err != nil {
				return err
			}
		case 1:
			_ = binary.Write(&req, binary.BigEndian, uint8(c.protocolVersion))
			_ = binary.Write(&req, binary.BigEndian, uint32(z.Len()))
			_, _ = req.Write(z.Bytes())
			fmt.Println(req.String())
			if _, err := c.client.Write(req.Bytes()); err != nil {
				return err
			}
		}
	} else {
		// deVer==1
		fmt.Println(z.String())
		if _, err := c.client.Write(z.Bytes()); err != nil {
			return err
		}
	}
	wlock.Unlock()
	return nil
}

func (c DeType) detectVersion() error {
	_ = c.sendCall(1, -1, "daemon.info")
	_ = c.sendCall(2, -1, "daemon.info")
	_ = c.sendCall(2, 1, "daemon.info")
	var buf bytes.Buffer
	// Seems I could read all the data from client.Read
	// resp data is like head...bodylen...body
	// and body is zlib compressed...
	// go to sleep and make this note (zzzz
	rlock.Lock()
	_ = c.client.SetDeadline(time.Now().Add(10 * time.Second))
	_, err := io.Copy(&buf, c.client)
	rlock.Unlock()
	resp := buf.Bytes()
	if err != nil {
		return err
	}
	if resp[0] == byte('D') {
		c.version = 2
		c.protocolVersion = -1
		// receive daemon_version!!!
	} else if resp[0] == 1 {
		c.version = 2
		c.protocolVersion = 1
		// receive daemon_version!!!
	} else {
		c.version = 1
		// Deluge 1 doesn't recover well from the bad request. Re-connect!
		if err := c.Reconnect(); err != nil {
			return err
		}
	}
	return nil
}

// recvResp: return resp body, reqID, isDeError, error
func (c DeType) recvResp() (*rencode.Decoder, error) {
	rlock.Lock()
	_, err := io.Copy(&rbuf, c.client)
	rlock.Unlock()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	var body bytes.Buffer
	bufreader := bufio.NewReader(&rbuf)
	w := bufio.NewWriter(&buf)
	bw := bufio.NewWriter(&body)
	//r := bufio.NewReader(&buf)

	if c.version == 2 {
		header, _ := bufreader.ReadByte()
		switch c.protocolVersion {
		case -1:
			if header != 'D' {
				return nil, fmt.Errorf("Expected D as first byte in reply.")
			}
			var n int32
			if err := binary.Read(bufreader, binary.BigEndian, &n); err != nil {
				return nil, err
			}
			if _, err := io.CopyN(w, bufreader, int64(n)); err != nil {
				return nil, err
			}
			r, _ := zlib.NewReader(&buf)
			_, _ = io.Copy(bw, r)
			r.Close()
			return rencode.NewDecoder(&body), nil

		case 1:
			if header != uint8(1) {
				return nil, fmt.Errorf("Expected protocol version %d as first byte in reply", 1)
			}
			var n uint32
			if err := binary.Read(bufreader, binary.BigEndian, &n); err != nil {
				return nil, err
			}
			if _, err := io.CopyN(w, bufreader, int64(n)); err != nil {
				return nil, err
			}
			r, _ := zlib.NewReader(&buf)
			_, _ = io.Copy(bw, r)
			r.Close()
			return rencode.NewDecoder(&body), nil
		}
	}

	return nil, nil
}
