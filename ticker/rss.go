package ticker

import (
	"io/ioutil"
	"net/http"
	"unsafe"

	"github.com/capric98/t-rss/feed"
	"github.com/sirupsen/logrus"
)

// NewRssTicker :)
func NewRssTicker(req *http.Request, client *http.Client, log *logrus.Entry) *Ticker {
	ch := make(chan []feed.Item, 10)
	go rssTicker(req, client, ch, log)
	return &Ticker{c: ch}
}

func rssTicker(req *http.Request, client *http.Client, ch chan []feed.Item, log *logrus.Entry) {
	log = log.WithField("@func", "rssTicker")
	for {
		resp, e := client.Do(req)
		if e != nil {
			log.Warn(e)
			continue
		}
		body, e := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		log.Trace("\n", *(*string)(unsafe.Pointer(&body)))
	}
}
