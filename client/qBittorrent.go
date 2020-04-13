package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"sync"
	"time"
)

// QBType :)
type QBType struct {
	client   *http.Client
	settings map[string]string
	name     string
	label    string
	mu       sync.RWMutex
}

var (
	qBparalist      = []string{"dlLimit", "upLimit", "savepath", "paused", "category", "skip_checking", "root_folder", "rename", "autoTMM", "sequentialDownload", "firstLastPiecePrio"}
	privateIPBlocks []*net.IPNet
)

// NewqBclient :)
func NewqBclient(key string, m map[string]interface{}) *QBType {
	nc := &QBType{
		client:   nil,
		settings: make(map[string]string),
		name:     "qBittorrent",
		label:    key,
		mu:       sync.RWMutex{},
	}

	for k, v := range m {
		switch v := v.(type) {
		case string:
			nc.settings[k] = v
		case bool:
			if v {
				nc.settings[k] = "true"
			} else {
				nc.settings[k] = "false"
			}
		case int:
			nc.settings[k] = fmt.Sprintf("%d", v)
		}
	} // Copy settings.

	if length := len(nc.settings["url"]); nc.settings["url"][length-1] == '/' {
		nc.settings["url"] = nc.settings["url"][:length-1]
	}
	nc.settings["dlLimit"] = UConvert(nc.settings["dlLimit"])
	nc.settings["upLimit"] = UConvert(nc.settings["upLimit"])

	fcount := 1
	initPrivateIP()
	err := nc.init()
	for err != nil {
		fcount++
		if fcount == 3 {
			log.Fatal(err)
		}
		err = nc.init()
	}
	return nc
}

func (c *QBType) init() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cookieJar, _ := cookiejar.New(nil)
	c.client = &http.Client{
		Timeout: 30 * time.Second,
		Jar:     cookieJar,
	}

	if c.settings["password"] == "" && isPrivateURL(c.settings["url"]) {
		log.Println(c.label + " qBittorrent client: You do not set username or password.")
		log.Println("Please make sure the client is running on local network, and make sure you have enabled no authentication for local user.")
		return nil
	}

	resp, err := c.client.PostForm(c.settings["url"]+"/api/v2/auth/login", url.Values{
		"username": {c.settings["username"]},
		"password": {c.settings["password"]},
	})
	if err != nil {
		log.Printf("Failed to initialize %s client: %v\n", c.label, err)
		return err
	}
	resp.Body.Close()
	return nil
}

// Name :)
func (c *QBType) Name() string {
	return c.label
}

// Add :)
func (c *QBType) Add(data []byte, filename string) (e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}() // In case of init fails.

	var try int
	for {
		e = c.call(data, filename)
		if e == nil {
			return
		}
		try++
		if try == 3 {
			return
		}
		_ = c.init()
	}
}

func (c *QBType) call(data []byte, filename string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Write config.
	for _, v := range qBparalist {
		if c.settings[v] != "" {
			if w.WriteField(v, c.settings[v]) != nil {
				return fmt.Errorf("Failed to write field %s", v)
			}
		}
	}
	// Write torrent body.
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="torrents"; filename="%s"`, filename))
	p, _ := w.CreatePart(h)
	if _, perr := p.Write(data); perr != nil {
		return perr
	}
	w.Close()

	req, err := http.NewRequest("POST", c.settings["url"]+"/api/v2/torrents/add", &b)
	if err != nil {
		return err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	c.mu.RLock()
	resp, err := c.client.Do(req)
	c.mu.RUnlock()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP code: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != "Ok." {
		return fmt.Errorf(c.label + "'s webui returns \"" + string(body) + "\" rather than \"Ok.\"")
	}
	return nil
}

func initPrivateIP() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateURL(webuiurl string) bool {
	u, err := url.Parse(webuiurl)
	if err != nil {
		log.Panicln("qBittorrent client: Cannot parse webui url.")
	}
	ip, err := net.LookupIP(u.Hostname())
	if err != nil {
		log.Printf("qBittorrent client: Cannot resolve %s, assuming you are running on local network.\n", webuiurl)
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip[0]) {
			return true
		}
	}
	return false
}
