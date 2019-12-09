package rss

import (
	"context"
	"html"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/capric98/t-rss/torrents"
	"github.com/mmcdole/gofeed"
)

type ticker struct {
	name         string
	client       *http.Client
	link, cookie string
	interval     time.Duration
	ctx          context.Context
}

var (
	fp = gofeed.NewParser()
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

	t.fetch(ch)
	for {
		select {
		case <-t.ctx.Done():
			close(ch)
			return
		case <-tt.C:
			t.fetch(ch)
		}
	}
}

func (t *ticker) fetch(ch chan []torrents.Individ) {
	defer func() {
		e := recover()
		if e != nil {
			log.Println("rss ticker:", e)
		}
	}()
	startT := time.Now()

	req, _ := http.NewRequest("GET", t.link, nil)
	if t.cookie != "" {
		req.Header.Add("Cookie", t.cookie)
	}

	resp, _ := t.client.Do(req)
	rssFeed, _ := fp.Parse(resp.Body)
	resp.Body.Close()

	ind := make([]torrents.Individ, len(rssFeed.Items))
	for i, v := range rssFeed.Items {
		ind[i] = torrents.Individ{
			Title:    v.Title,
			Descript: html.UnescapeString(v.Description),
			Date:     v.Published,
			GUID:     NameRegularize(v.GUID),
		}
		if v.Enclosures != nil {
			tmp, err := strconv.Atoi(v.Enclosures[0].Length)
			if err != nil {
				tmp = 0
			}
			ind[i].DUrl = v.Enclosures[0].URL
			ind[i].Length = int64(tmp)
		} else {
			ind[i].DUrl = v.Link
		}
		if v.Author != nil {
			ind[i].Author = v.Author.Name
		}
		if ind[i].GUID == "" {
			ind[i].GUID = NameRegularize(v.Title)
			if len(ind[i].GUID) > 200 {
				ind[i].GUID = ind[i].GUID[:200]
			}
		}
	}

	log.Printf("%s fetched in %7.2fms.", t.name, time.Since(startT).Seconds()*1000.0)
	ch <- ind
}
