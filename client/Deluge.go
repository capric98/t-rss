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
	"os"
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
	rechan     chan bool
	cchan      chan bool
	bufreader  = bufio.NewReader(&rbuf)
)

type DeType struct {
	Name            string
	settings        map[string]string
	client          *tls.Conn
	version         int
	protocolVersion int // -1 for None
}

func keepRead(c *tls.Conn) {
	select {
	case <-cchan:
		_, _ = io.Copy(&rbuf, c)
	case <-rechan:
		return
	}
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
	tclient := &DeType{
		client:   nil,
		settings: make(map[string]string),
	}
	for k, v := range m {
		tclient.settings[k.(string)] = v.(string)
	} // Copy settings.

	// tclient.settings["dlLimit"] = speedParse(tclient.settings["dlLimit"])
	// tclient.settings["upLimit"] = speedParse(tclient.settings["upLimit"])

	fcount := 0
	err := tclient.Init()
	for err != nil {
		fmt.Println(err)
		fcount++
		if fcount == 3 {
			log.Fatal(err)
		}
		err = tclient.Init()
	}

	nc.Client = tclient
	return nc
}

func makeList(args ...interface{}) *rencode.List {
	list := rencode.NewList()
	for _, v := range args {
		list.Add(v)
	}
	return &list
}

func (c *DeType) Init() error {
	conn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
		InsecureSkipVerify: true,
	})
	// Deluge use self-signed cert and seems must to use it...
	fmt.Println("Dial success!")
	if err != nil {
		return err
	}
	c.client = conn
	//go keepRead(conn)
	if err := c.detectVersion(); err != nil {
		return err
	}

	if c.version == 2 {
		if err := c.Call("daemon.login", makeList(c.settings["username"], c.settings["password"]), nil); err != nil {
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
		if err := c.Call("daemon.login", makeList(c.settings["username"], c.settings["password"]), nil); err != nil {
			log.Fatal(err)
			return err
		}
		fmt.Printf("%+q\n", rbuf.String())
		fmt.Println("===========================1")
		d, err := c.recvResp()
		if err != nil {
			fmt.Println(err)
			return err
		}
		var l rencode.List
		d.Scan(&l)
		fmt.Println(l)
	}
	os.Exit(1)
	return nil
}

func (c *DeType) Reconnect() error {
	//rechan <- true
	c.client.Close()

	nconn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	c.client = nconn

	//go keepRead(nconn)

	return nil
}

func (c *DeType) detectVersion() error {
	_ = c.sendCall(1, -1, "daemon.info", nil, nil)
	_ = c.sendCall(2, -1, "daemon.info", nil, nil)
	_ = c.sendCall(2, 1, "daemon.info", nil, nil)
	var buf bytes.Buffer
	// Seems I could read all the data from client.Read
	// resp data is like head...bodylen...body
	// and body is zlib compressed...
	// go to sleep and make this note (zzzz
	fmt.Println("Detect resp.")
	rlock.Lock()
	_, err := io.CopyN(&buf, c.client, 1)
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
		//Deluge 1 doesn't recover well from the bad request. Re-connect!
		if err := c.Reconnect(); err != nil {
			return err
		}
	}
	fmt.Println("Deluge version:", c.version)
	fmt.Println("Protocal version:", c.protocolVersion)

	for {
		_ = c.client.SetReadDeadline(time.Now().Add(1 * time.Second))
		if _, err := io.CopyN(&rbuf, c.client, 1); err != nil {
			break
		}
	}
	fmt.Println(rbuf.Len())
	rbuf.Truncate(0)
	fmt.Println(rbuf.Len())
	return nil
}

func (c *DeType) Call(method string, args *rencode.List, kargs *rencode.Dictionary) error {
	return c.sendCall(c.version, c.protocolVersion, method, args, kargs)
}

func (c *DeType) sendCall(deVer int, protoVer int, method string, args *rencode.List, kargs *rencode.Dictionary) error {
	reqID := c.newReqID()
	fmt.Println("Request ID:", reqID)
	var b, z bytes.Buffer
	var reql rencode.List
	e := rencode.NewEncoder(&b)
	if args == nil {
		if kargs == nil {
			reql = rencode.NewList(reqID, method, nil, nil)
		} else {
			reql = rencode.NewList(reqID, method, nil, *kargs)
		}
	} else {
		if kargs == nil {
			reql = rencode.NewList(reqID, method, *args, nil)
		} else {
			reql = rencode.NewList(reqID, method, *args, *kargs)
		}
	}
	if err := e.Encode(rencode.NewList(reql)); err != nil {
		return err
	}
	tmp := zlib.NewWriter(&z)
	_, _ = tmp.Write(b.Bytes())
	// The output from the python example isn't a "complete" stream,
	// its just flushing the buffer after compressing the first string.
	// You can get the same output from the Go code by replacing Close() with Flush()
	//tmp.Flush()
	tmp.Close()
	fmt.Println("z len:", z.Len())

	wlock.Lock()
	_ = c.client.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if deVer == 2 {
		var req bytes.Buffer
		switch protoVer {
		case -1:
			req.WriteRune('D')
			_ = binary.Write(&req, binary.BigEndian, int32(z.Len()))
			_, _ = req.Write(z.Bytes())
			//fmt.Printf("%+q\n", z.String())
			if _, err := c.client.Write(req.Bytes()); err != nil {
				return err
			}
		case 1:
			_ = binary.Write(&req, binary.BigEndian, uint8(c.protocolVersion))
			_ = binary.Write(&req, binary.BigEndian, uint32(z.Len()))
			_, _ = req.Write(z.Bytes())
			//fmt.Printf("%+q\n", z.String())
			if _, err := c.client.Write(req.Bytes()); err != nil {
				return err
			}
		}
	} else {
		// deVer==1
		fmt.Printf("%+q\n", z.String())
		if _, err := c.client.Write(z.Bytes()); err != nil {
			return err
		}
	}
	wlock.Unlock()
	return nil
}

func (c *DeType) recvResp() (*rencode.Decoder, error) {
	var buf bytes.Buffer
	var body bytes.Buffer
	w := bufio.NewWriter(&buf)
	bw := bufio.NewWriter(&body)
	//r := bufio.NewReader(&buf)

	for {
		_ = c.client.SetReadDeadline(time.Now().Add(1 * time.Second))
		if _, err := io.CopyN(&rbuf, c.client, 1); err != nil {
			break
		}
	}

	tmp := rbuf.Bytes()
	rtmp := bytes.NewReader(tmp)
	//
	//fmt.Println(rbuf.Len())

	fmt.Println("Read done.")

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
		}
		r, _ := zlib.NewReader(&buf)
		_, _ = io.Copy(bw, r)
		r.Close()
		return rencode.NewDecoder(&body), nil
	}

	r, _ := zlib.NewReader(rtmp)
	_, _ = io.Copy(bw, r)
	r.Close()
	return rencode.NewDecoder(&body), nil
}
