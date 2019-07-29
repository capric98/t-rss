package client

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"time"
)

type QBType struct {
	client   *http.Client
	settings map[string]string
}

var (
	qBparalist = []string{"dlLimit", "upLimit", "savepath", "paused"}
)

func NewqBclient(m map[interface{}]interface{}) ClientType {
	var nc ClientType
	nc.Name = "qBittorrent"

	cookieJar, _ := cookiejar.New(nil)
	nc.Client = QBType{
		client: &http.Client{
			Timeout: 60 * time.Second,
			Jar:     cookieJar,
		},
		settings: make(map[string]string),
	}
	for k, v := range m {
		nc.Client.(QBType).settings[k.(string)] = v.(string)
	} // Copy settings.

	nc.Client.(QBType).settings["dlLimit"] = speedParse(nc.Client.(QBType).settings["dlLimit"])
	nc.Client.(QBType).settings["upLimit"] = speedParse(nc.Client.(QBType).settings["upLimit"])

	fcount := 1
	err := nc.Client.(QBType).Init()
	for err != nil {
		fcount++
		if fcount == 3 {
			log.Fatal(err)
		}
		err = nc.Client.(QBType).Init()
	}
	return nc
}

func (c QBType) Init() error {
	resp, err := c.client.PostForm(c.settings["host"]+"/login", url.Values{
		"username": {c.settings["username"]},
		"password": {c.settings["password"]},
	})
	if err != nil {
		log.Printf("Failed to initialize client: %v\n", err)
		return err
	}
	resp.Body.Close()
	return nil
}

func (c QBType) Add(data []byte, filename string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Write config.
	for _, v := range qBparalist {
		if c.settings[v] != "" {
			w.WriteField(v, c.settings[v])
		}
	}
	// Write torrent body.
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="torrents"; filename="%s"`, filename))
	p, _ := w.CreatePart(h)
	p.Write(data)
	w.Close()

	req, err := http.NewRequest("POST", c.settings["host"]+"/command/upload", &b)
	if err != nil {
		return err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != "Ok." {
		return errors.New(string(body))
	}

	resp.Body.Close()
	return nil
}
