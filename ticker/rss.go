package ticker

import (
	"io/ioutil"
	"net/http"
	"time"
	"unsafe"

	"github.com/capric98/t-rss/feed"
	"github.com/sirupsen/logrus"
)

// NewRssTicker :)
func NewRssTicker(n int, req *http.Request, client *http.Client, log *logrus.Entry, interval time.Duration) *Ticker {
	ch := make(chan []feed.Item, 10)
	go rssTicker(n, req, client, ch, log, interval)
	return &Ticker{c: ch}
}

func rssTicker(n int, req *http.Request, client *http.Client, ch chan []feed.Item, log *logrus.Entry, interval time.Duration) {
	log = log.WithField("@func", "rssTicker")

	times := byte(0)
	retry := 0
	for {
		if retry == 3 {
			log.Debug("reset ticker due to too many retry")
			retry = 0
			time.Sleep(interval)
		}
		retry++

		resp, e := client.Do(req)
		if e != nil {
			log.Warn(e)
			continue
		}
		body, e := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		log.Trace("\n", *(*string)(unsafe.Pointer(&body)))

		items, e := feed.Parse(body)
		if e != nil {
			log.Warn("parse: ", e)
			continue
		}
		ch <- items

		if times++; int(times) == n {
			close(ch)
			return
		}
		time.Sleep(interval)
		retry = 0
	}
}
