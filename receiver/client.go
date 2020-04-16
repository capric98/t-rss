package receiver

import (
	"github.com/capric98/t-rss/client"
	"github.com/capric98/t-rss/feed"
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
func (c *Client) Push(i *feed.Item, b []byte) error {
	return c.Client.Add(b, i.Title)
}
