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
		ftype:    myfeed.RSSType,
		debug:    debug,
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
	defer resp.Body.Close()
	rssFeed, e := myfeed.Parse(resp.Body, t.ftype)

	// debug, e := ioutil.ReadAll(resp.Body)
	// resp.Body.Close()
	// if e != nil {
	// 	return
	// }
	// log.Println(string(debug))
	// rssFeed, e := myfeed.Parse(ioutil.NopCloser(bytes.NewReader(debug)), t.ftype)

	if e != nil {
		if t.debug {
			log.Println("myfeed:", e)
		}
		if t.ftype == myfeed.AtomType {
			t.ftype = myfeed.RSSType
		} else {
			t.ftype = myfeed.AtomType
		}
	}

	for k := range rssFeed {
		if rssFeed[k].Enclosure.Url == "" {
			rssFeed[k].Enclosure.Url = rssFeed[k].Link
		}
		if rssFeed[k].GUID.Value == "" {
			rssFeed[k].GUID.Value = myfeed.NameRegularize(rssFeed[k].Title)
		}
		rssFeed[k].GUID.Value = myfeed.NameRegularize(rssFeed[k].GUID.Value)
	}

	if t.debug {
		log.Printf("%s fetched in %7.2fms.", t.name, time.Since(startT).Seconds()*1000.0)
	}
	ch <- rssFeed
}
