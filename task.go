package trss

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/capric98/t-rss/bencode"
	"github.com/capric98/t-rss/feed"
	"github.com/capric98/t-rss/filter"
	"github.com/capric98/t-rss/receiver"
	"github.com/capric98/t-rss/setting"
	"github.com/capric98/t-rss/ticker"
	"github.com/capric98/t-rss/unit"
	"github.com/sirupsen/logrus"
)

type worker struct {
	client *http.Client
	header map[string][]string
	tick   *ticker.Ticker

	filters []filter.Filter
	recvers []receiver.Receiver
	quota   setting.Quota
	delay   time.Duration
	wpath   string

	edit *setting.Edit

	logger func() *logrus.Entry
	ctx    context.Context
	wg     *sync.WaitGroup
}

func (w *worker) prepare(t *setting.Task, num int) {
	// make ticker
	if t.Rss != nil {
		req, err := http.NewRequest(t.Rss.Method, t.Rss.URL, nil)
		if err != nil {
			w.logger().Fatal(err)
		}
		req.Header = w.header
		w.tick = ticker.NewRssTicker(num, req, w.client, w.logger(), t.Rss.Interval.T)
	}

	// make filter
	w.filters = append(w.filters, filter.NewRegexpFilter(t.Filter.Regexp.Accept, t.Filter.Regexp.Reject))
	w.filters = append(w.filters, filter.NewContentSizeFilter(t.Filter.ContentSize.Min.I, t.Filter.ContentSize.Max.I))

	// make receiver only if learn==false (num==-1)
	if num == -1 {
		if t.Receiver.Save != nil {
			w.recvers = append(w.recvers, receiver.NewDownload(*t.Receiver.Save))
		}
		for k, v := range t.Receiver.Client {
			w.recvers = append(w.recvers, receiver.NewClient(v["type"], v, k))
		}
	}
}

func (w *worker) loop() {
	defer w.wg.Done()
	defer w.tick.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case items, ok := <-w.tick.C():
			if !ok {
				w.logger().Debug("ticker closed, return")
				return
			} // under "learn" mode or ticker crashed

			var passed []feed.Item
			var accept, reject int
			for k := range items {
				log := w.logger().WithFields(logrus.Fields{
					"author":   items[k].Author,
					"category": items[k].Category,
					"GUID":     items[k].GUID,
					"size":     unit.FormatSize(items[k].Len),
				})

				// Check if have seen.
				historyPath := w.wpath + items[k].GUID
				if _, err := os.Stat(historyPath); !os.IsNotExist(err) {
					log.Trace("(reject) have seen ", items[k].Title, " before.")
					reject++
					continue
				} else {
					hf, err := os.Create(historyPath)
					if err != nil {
						log.Warn("create history file: ", err)
					} else {
						hf.Close()
					}
				}

				flag := true
				for _, f := range w.filters {
					if e := f.Check(&items[k]); e != nil {
						flag = false
						reject++
						log.Info("(reject) ", e)
						break
					}
				}
				if flag {
					accept++
					log.Info("(accept) ", items[k].Title)
					passed = append(passed, items[k])
				}
			}
			w.logger().Info("accepted ", accept, " item(s), rejected ", reject, " item(s).")
			go w.push(passed)
		}
	}
}

func (w *worker) push(items []feed.Item) {
	quota := w.quota
	mu := sync.Mutex{}
	for k := range items {
		go func(item feed.Item) {
			time.Sleep(w.delay)

			// preperation
			start := time.Now()
			log := w.logger().WithFields(logrus.Fields{
				"title": item.Title,
			})

			// request torrent's body
			req, e := http.NewRequest("GET", item.URL, nil)
			if e != nil {
				log.Warn("new request: ", e)
				return
			}
			req.Header = w.header
			resp, e := w.client.Do(req)
			for retry := 0; e != nil && retry < 3; retry++ {
				resp, e = w.client.Do(req)
			}
			if e != nil {
				log.Warn("client.Do(): ", e)
				return
			}
			body, e := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if e != nil {
				log.Warn("read response body: ", e)
				return
			}

			// double check in case of Len=0
			if item.Len == 0 {
				item.Len = tLen(body)
				for _, f := range w.filters {
					if e := f.Check(&item); e != nil {
						log.Info("(reject in double check) ", e)
						return
					}
				}
			}

			// check quota
			mu.Lock()
			if quota.Num > 0 && quota.Size.I >= item.Len {
				quota.Num--
				quota.Size.I -= item.Len
			} else {
				log.Info("(drop) quota exceeded (left Num=", quota.Num, " Size=", unit.FormatSize(quota.Size.I), ")")
				mu.Unlock()
				return
			}
			mu.Unlock()

			if w.edit != nil {
				log.Debug("edit torrent...")
				body, e = w.edit.EditTorrent(body)
				if e != nil {
					w.logger().WithFields(logrus.Fields{
						"@func": "editTorrent",
					}).Warn(e)
					return
				}
			}

			// push to every receiver
			recvwg := sync.WaitGroup{}
			for i := range w.recvers {
				recvwg.Add(1)
				go func(recv receiver.Receiver) {
					err := recv.Push(&item, body)
					if err != nil {
						log.Warn("push to ", recv.Name(), " : ", err)
					} else {
						log.WithField("@cost", time.Since(start)).Info("push to ", recv.Name())
					}
					recvwg.Done()
				}(w.recvers[i])
			}
			// preserve (start)
			recvwg.Wait()
		}(items[k])
	}
}

func tLen(data []byte) (l int64) {
	defer func() { _ = recover() }()

	result, err := bencode.Decode(data)
	if err != nil {
		return
	}
	info := result[0].Dict("info")
	pl := (info.Dict("piece length")).Value()
	ps := int64(len((info.Dict("pieces")).BStr())) / 20
	l = pl * ps

	return
}
