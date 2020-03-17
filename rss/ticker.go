package rss

import (
	"bytes"
	"context"
	"io/ioutil"
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
	ftype        int
	debug        bool
}

func NewTicker(name string, link string, cookie string, interval time.Duration, wc *http.Client, ctx context.Context, debug bool) (ch chan []torrents.Individ) {
	t := &ticker{
		name:     name,
		client:   wc,
		cookie:   cookie,
		link:     link,
		interval: interval,
		ctx:      ctx,
		ftype:    -1,
		debug:    debug,
	}
	ch = make(chan []torrents.Individ, 2)
	go t.tick(ch)
	return
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

	var respFeed []myfeed.Item

	for counter := 0; counter < 5 && respFeed == nil; counter++ {
		resp, e := t.client.Do(req)
		if e != nil {
			return
		}

		if t.ftype == -1 {
			fbody, _ := ioutil.ReadAll(resp.Body)
			rfeed, fe := myfeed.Parse(bytes.NewReader(fbody), myfeed.RSSType)
			afeed, ae := myfeed.Parse(bytes.NewReader(fbody), myfeed.AtomType)
			if fe == nil {
				t.ftype = myfeed.RSSType
				respFeed = rfeed
			} else {
				if ae == nil {
					t.ftype = myfeed.AtomType
					respFeed = afeed
				}
			}
		} else {
			respFeed, e = myfeed.Parse(resp.Body, t.ftype)
		}

		resp.Body.Close()
		if e != nil {
			log.Println("myfeed:", e)
		}
	}

	if respFeed == nil {
		log.Println("myfeed: Got nil result, please consider to retry latter.")
	} else {
		for k := range respFeed {
			if respFeed[k].Enclosure.Url == "" {
				respFeed[k].Enclosure.Url = respFeed[k].Link
			}
			if respFeed[k].GUID.Value == "" {
				respFeed[k].GUID.Value = myfeed.NameRegularize(respFeed[k].Title)
			}
			respFeed[k].GUID.Value = myfeed.NameRegularize(respFeed[k].GUID.Value)
		}

		if t.debug {
			log.Printf("%s fetched in %7.2fms.", t.name, time.Since(startT).Seconds()*1000.0)
		}

		ch <- respFeed
	}
}
