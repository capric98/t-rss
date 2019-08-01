package client

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/binary"
	"sync"

	"github.com/gdm85/go-rencode"
)

var (
	// messageHeaderFormat = "!BI"
	// messageHeaderSize   = 5
	request_id = 0
	glock      sync.Mutex
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

func (c DeType) Init() error {
	conn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
		InsecureSkipVerify: true,
	}) // Deluge use self-signed cert and seems must to use it...
	if err != nil {
		return err
	}
	c.client = conn
	c.newReqID()
	if err := c.detectVersion(); err != nil {
		return err
	}
	return nil
}

func (c DeType) sendCall(deVer int, protoVer int, method string, args ...interface{}) error {
	reqID := c.newReqID()
	var b, z bytes.Buffer
	e := rencode.NewEncoder(&b)
	if err := e.Encode(rencode.NewList(rencode.NewList(reqID, method))); err != nil {
		return err
	}
	tmp := zlib.NewWriter(&z)
	_, _ = tmp.Write(b.Bytes())
	tmp.Close()

	if c.version == 2 {
		var req bytes.Buffer
		switch c.protocolVersion {
		case -1:
			req.WriteRune('D')
			_ = binary.Write(&req, binary.BigEndian, int32(z.Len()))
			_, _ = req.Write(z.Bytes())
			if _, err := c.client.Write(req.Bytes()); err != nil {
				return err
			}
		case 1:
			_ = binary.Write(&req, binary.BigEndian, uint8(c.protocolVersion))
			_ = binary.Write(&req, binary.BigEndian, uint32(z.Len()))
			_, _ = req.Write(z.Bytes())
			if _, err := c.client.Write(req.Bytes()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c DeType) detectVersion() error {
	_ = c.sendCall(1, -1, "daemon.info")
	_ = c.sendCall(2, -1, "daemon.info")
	_ = c.sendCall(2, 1, "daemon.info")
	resp := make([]byte, 0, 1)
	// Seems I could read all the data from client.Read
	// resp data is like head...bodylen...body
	// and body is zlib compressed...
	// go to sleep and make this note (zzzz
	_, err := c.client.Read(resp)
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
		c.client.Close()
		conn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return err
		}
		c.client = conn
	}
	return nil
}
