package receiver

import (
	"fmt"

	"github.com/capric98/t-rss/client"
)

// Client :)
type Client struct {
	client.Client
}

// NewClient :)
func NewClient(tYPE interface{}, conf map[string]interface{}, name string) Receiver {
	stype := tYPE.(string) // let it crash if got non string value

	var cc client.Client
	switch stype {
	case "qBittorrent":
		cc = client.NewqBclient(name, conf)
	case "Deluge":
		cc = client.NewDeClient(name, conf)
	}
	rc := &Client{}
	rc.Client = cc
	return rc
}

// Push :)
func (c *Client) Push(b []byte, i interface{}) error {
	fn, ok := i.(string)
	if !ok {
		return fmt.Errorf("wanted string but got %T", i)
	}
	return c.Client.Add(b, fn)
}
