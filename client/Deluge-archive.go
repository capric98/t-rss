package client

// import (
// 	"bufio"
// 	"bytes"
// 	"compress/zlib"
// 	"crypto/tls"
// 	"encoding/base64"
// 	"encoding/binary"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"sync"
// 	"time"

// 	"github.com/gdm85/go-rencode"
// )

// var (
// 	glock      sync.Mutex
// 	rlock      sync.Mutex
// 	wlock      sync.Mutex
// 	rbuf       bytes.Buffer
// 	rechan     = make(chan bool)
// 	bufreader  = bufio.NewReader(&rbuf)
// 	request_id = 0
// 	wait       = 100
// )

// type DeType struct {
// 	Name            string
// 	settings        map[string]string
// 	client          *tls.Conn
// 	version         int
// 	protocolVersion int // -1 for None
// }

// func keepRead(c *tls.Conn) {
// 	for {
// 		select {
// 		case <-rechan:
// 			return
// 		default:
// 			_ = c.SetReadDeadline(time.Now().Add(1 * time.Second))
// 			if _, err := io.CopyN(&rbuf, c, 1); err != nil {
// 				continue
// 			}
// 		}
// 	}
// }

// func (c DeType) newReqID() int {
// 	glock.Lock()
// 	request_id++
// 	defer glock.Unlock()
// 	return request_id
// }

// func Test() {
// 	var buf bytes.Buffer
// 	e := rencode.NewEncoder(&buf)
// 	var dict rencode.Dictionary
// 	dict.Add("client_version", "deluge-client")
// 	_ = e.Encode(dict)
// 	fmt.Printf("%+q\n", buf.String())
// }

// func NewDeClient(m map[interface{}]interface{}) ClientType {
// 	Test()

// 	var nc ClientType
// 	nc.Name = "Deluge"
// 	tclient := &DeType{
// 		client:   nil,
// 		settings: make(map[string]string),
// 	}
// 	for k, v := range m {
// 		tclient.settings[k.(string)] = v.(string)
// 	} // Copy settings.

// 	// tclient.settings["dlLimit"] = speedParse(tclient.settings["dlLimit"])
// 	// tclient.settings["upLimit"] = speedParse(tclient.settings["upLimit"])

// 	fcount := 0
// 	err := tclient.Init()
// 	for err != nil {
// 		fmt.Println(err)
// 		fcount++
// 		if fcount == 3 {
// 			log.Fatal(err)
// 		}
// 		err = tclient.Init()
// 	}

// 	nc.Client = tclient
// 	return nc
// }

// func (c *DeType) Init() error {
// 	conn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
// 		InsecureSkipVerify: true,
// 	})
// 	// Deluge use self-signed cert and seems must to use it...
// 	if err != nil {
// 		return err
// 	}
// 	c.client = conn
// 	go keepRead(conn)
// 	if err := c.detectVersion(); err != nil {
// 		return err
// 	}

// 	rbuf.Truncate(0)
// 	fmt.Println("rbuf len=", rbuf.Len())

// 	if c.version == 2 {
// 		if err := c.Call("daemon.login", makeList(c.settings["username"], c.settings["password"]), makeDict(nil)); err != nil {
// 			return err
// 		}
// 		d, err := c.recvResp()
// 		if err != nil {
// 			return err
// 		}
// 		for {
// 			i, e := d.DecodeNext()
// 			if e == io.EOF {
// 				break
// 			}
// 			fmt.Println(i)
// 		}
// 	} else {
// 		if err := c.Call("daemon.login", makeList(c.settings["username"], c.settings["password"]), makeDict(nil)); err != nil {
// 			log.Fatal(err)
// 			return err
// 		}
// 		d, err := c.recvResp()
// 		if err != nil {
// 			fmt.Println(err)
// 			return err
// 		}
// 		var l, ll rencode.List
// 		var rtype, id int
// 		_ = d.Scan(&l)
// 		_ = l.Scan(&rtype, &id, &ll)
// 		if rtype != 1 {
// 			var estring string
// 			for _, v := range ll.Values() {
// 				estring = estring + fmt.Sprintf("%s\n", string(v.([]uint8)))
// 			}
// 			return fmt.Errorf(estring)
// 		} else {
// 			return nil
// 		}
// 	}
// 	os.Exit(1)
// 	return nil
// }

// func (c *DeType) Reconnect() error {
// 	if err := c.CloseSocket(); err != nil {
// 		return err
// 	}

// 	nconn, err := tls.Dial("tcp", c.settings["host"], &tls.Config{
// 		InsecureSkipVerify: true,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	c.client = nconn

// 	go keepRead(nconn)

// 	return nil
// }

// func (c *DeType) CloseSocket() error {
// 	rechan <- true
// 	if err := c.client.Close(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (c *DeType) detectVersion() error {
// 	_ = c.sendCall(1, -1, "daemon.info", makeList(), makeDict(nil))
// 	_ = c.sendCall(2, -1, "daemon.info", makeList(), makeDict(nil))
// 	_ = c.sendCall(2, 1, "daemon.info", makeList(), makeDict(nil))
// 	var buf bytes.Buffer

// 	time.Sleep(time.Duration(wait) * time.Millisecond)

// 	rlock.Lock()
// 	_, err := io.CopyN(&buf, bufreader, 1)
// 	rlock.Unlock()

// 	resp := buf.Bytes()
// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}
// 	if resp[0] == byte('D') {
// 		c.version = 2
// 		c.protocolVersion = -1
// 		// receive daemon_version!!!
// 	} else if resp[0] == 1 {
// 		c.version = 2
// 		c.protocolVersion = 1
// 		// receive daemon_version!!!
// 	} else {
// 		c.version = 1
// 		//Deluge 1 doesn't recover well from the bad request. Re-connect!
// 		if err := c.Reconnect(); err != nil {
// 			return err
// 		}
// 	}
// 	fmt.Println("Deluge version:", c.version)
// 	fmt.Println("Protocal version:", c.protocolVersion)

// 	rbuf.Truncate(0)
// 	return nil
// }

// func (c *DeType) Call(method string, args *rencode.List, kargs *rencode.Dictionary) error {
// 	return c.sendCall(c.version, c.protocolVersion, method, args, kargs)
// }

// func (c *DeType) Add(data []byte, name string) error {
// 	b64 := base64.StdEncoding.EncodeToString(data)
// 	m := make(map[interface{}]interface{})
// 	m["download_location"] = "/home/Downloads/"
// 	_ = c.sendCall(c.version, c.protocolVersion, "core.add_torrent_file", makeList(name, b64), makeDict(m))
// 	d, e := c.recvResp()
// 	if e != nil {
// 		fmt.Println(e)
// 		return e
// 	}
// 	fmt.Println(d.DecodeNext())
// 	return nil
// }
