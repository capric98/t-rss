package rss

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/capric98/t-rss/myfeed"
	"github.com/capric98/t-rss/torrents"
)

type ticker struct {
	name         string
	client       *http.Client
	link, cookie string
	interval     time.Duration
	ctx          context.Context
}

var (
	chunk = make([]byte, 1)
	data  = []byte{}
	rerr  error
)

func NewTicker(name string, link string, cookie string, interval time.Duration, wc *http.Client, ctx context.Context) (ch chan []torrents.Individ) {
	t := &ticker{
		name:     name,
		client:   wc,
		cookie:   cookie,
		link:     link,
		interval: interval,
		ctx:      ctx,
	}
	ch = make(chan []torrents.Individ)
	go t.tick(ch)
	return ch
}

func (t *ticker) tick(ch chan []torrents.Individ) {
	tt := time.NewTicker(t.interval)
	defer tt.Stop()

	req, _ := http.NewRequest("GET", t.link, nil)
	if t.cookie != "" {
		req.Header.Add("Cookie", t.cookie)
	}
	t.fetch(req, ch)
	for {
		select {
		case <-t.ctx.Done():
			close(ch)
			return
		case <-tt.C:
			t.fetch(req, ch)
		}
	}
}

func (t *ticker) fetch(req *http.Request, ch chan []torrents.Individ) {
	defer func() {
		e := recover()
		if e != nil {
			log.Println("rss ticker:", e)
		}
	}()
	startT := time.Now()

	resp, e := t.client.Do(req)
	if e != nil {
		return
	}
	data = data[:0]
	_, rerr = resp.Body.Read(chunk)
	for rerr == nil {
		data = append(data, chunk[0])
		_, rerr = resp.Body.Read(chunk)
	}
	resp.Body.Close()
	rssFeed, _ := myfeed.Parse(data)

	for k := range rssFeed.Items {
		if rssFeed.Items[k].Enclosure.Url == "" {
			rssFeed.Items[k].Enclosure.Url = rssFeed.Items[k].Link
		}
		if rssFeed.Items[k].GUID.Value == "" {
			rssFeed.Items[k].GUID.Value = myfeed.NameRegularize(rssFeed.Items[k].Title)
		}
		rssFeed.Items[k].GUID.Value = myfeed.NameRegularize(rssFeed.Items[k].GUID.Value)
	}

	log.Printf("%s fetched in %7.2fms.", t.name, time.Since(startT).Seconds()*1000.0)
	ch <- rssFeed.Items
	//runtime.GC()
}
