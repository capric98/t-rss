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
	"github.com/capric98/t-rss/receiver"
	"github.com/capric98/t-rss/setting"
	"github.com/capric98/t-rss/ticker"
	"github.com/capric98/t-rss/unit"
	"github.com/sirupsen/logrus"
)

type worker struct {
	*ticker.Ticker
	*http.Client

	accept, reject []setting.Reg
	min, max       int64
	quota          setting.Quota
	edit           *setting.Edit
	receiver       []receiver.Receiver

	header      map[string]string
	ctx         context.Context
	logger      func() *logrus.Entry
	historyPath string
}

func doTask(ctx context.Context, n int, t *setting.Task, client *http.Client, nlfunc func() *logrus.Entry, wg *sync.WaitGroup, history string) {
	if _, e := os.Stat(history); os.IsNotExist(e) {
		e = os.MkdirAll(history, 0640)
		if e != nil {
			nlfunc().Fatal("cannot mkdir: ", e)
		}
	}
	// nlfunc().Debugf("%#v\n", t)
	w := &worker{
		accept:      t.Filter.Regexp.Accept,
		reject:      t.Filter.Regexp.Reject,
		min:         t.Filter.ContentSize.Min,
		max:         t.Filter.ContentSize.Max,
		quota:       t.Quota,
		edit:        t.Edit,
		header:      t.Rss.Headers,
		ctx:         ctx,
		logger:      nlfunc,
		historyPath: history,
	}
	w.Client = client

	if t.Rss != nil {
		req, _ := http.NewRequest(t.Rss.Method, t.Rss.URL, nil)
		for k, v := range t.Rss.Headers {
			req.Header.Add(k, v)
		}

		w.Ticker = ticker.NewRssTicker(n, req, client, nlfunc(), t.Rss.Interval.T)
	}
	// make receiver
	if t.Receiver.Save != nil {
		w.receiver = append(w.receiver, receiver.NewDownload(*t.Receiver.Save))
	}
	for k, v := range t.Receiver.Client {
		// nlfunc().Infof("%#v", v)
		w.receiver = append(w.receiver, receiver.NewClient(v["type"], v, k))
	}

	// nlfunc().Debugf("%#v\n", w)
	go w.do(wg)
}

func (w *worker) do(wg *sync.WaitGroup) {
	defer wg.Done()
	defer w.Ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case items, ok := <-w.C():
			if !ok {
				w.logger().Debug("ticker closed, return")
				return
			}

			var passed []feed.Item
			var accept, reject int
			quota := w.quota
			for _, v := range items {
				log := w.logger().WithFields(logrus.Fields{
					"author":   v.Author,
					"category": v.Category,
					"hash":     v.GUID,
					"size":     unit.FormatSize(v.Len),
				})

				// Check if have seen.
				if _, err := os.Stat(w.historyPath + v.GUID); !os.IsNotExist(err) {
					log.Debug(`reject "`, v.Title, `" - have seen before.`)
					reject++
					continue
				} else {
					hf, err := os.Create(w.historyPath + v.GUID)
					if err != nil {
						log.Warn("create history file: ", err)
					} else {
						hf.Close()
					}
				}
				// Check regexp filter.
				if match, ms := checkRegexp(v, w.reject); match {
					log.Info(`reject "`, v.Title, `" with matching Reject - `, ms)
					reject++
					continue
				}
				if match, _ := checkRegexp(v, w.accept); (len(w.accept) > 0) && (!match) {
					log.Info(`reject "`, v.Title, `" with no matching Accept`)
					reject++
					continue
				}
				// Check content_size.
				if v.Len != 0 && (v.Len < w.min || v.Len > w.max) {
					log.Info(`reject "`, v.Title, `" - `, v.Len, " not in [", w.min, ",", w.max, "]")
					reject++
					continue
				}
				// Check quota
				if quota.Num <= 0 || quota.Size < v.Len {
					log.Info(`reject "`, v.Title, `" - quota exceeded(left num=`, quota.Num, " size=", unit.FormatSize(quota.Size), ")")
					reject++
					continue
				}

				log.Info("accept ", v.Title)
				passed = append(passed, v)
				accept++
				quota.Num--
				quota.Size -= v.Len
			}
			w.logger().Info("accepted ", accept, " item(s), rejected ", reject, " item(s)")
			go w.push(passed)
		}
	}
}

func (w *worker) push(it []feed.Item) {
	for _, v := range it {
		go func(item feed.Item) {
			log := w.logger().WithFields(logrus.Fields{
				"title": item.Title,
			})
			req, _ := http.NewRequest("GET", item.URL, nil)
			for hk, hv := range w.header {
				req.Header.Add(hk, hv)
			}
			resp, e := w.Do(req)
			for retry := 0; e != nil && retry < 3; retry++ {
				resp, e = w.Do(req)
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

			if item.Len == 0 {
				if pass, len := w.checkTLength(body); !pass {
					log.Warn("reject in double check: ", len, " is not in [", w.min, ",", w.max, "]")
					return
				}
			}
			if w.edit != nil {
				log.Debug("edit torrent...")
				body = w.editTorrent(body)
			}

			for k := range w.receiver {
				go func(i int) {
					start := time.Now()
					err := w.receiver[i].Push(body, item.Title)
					if err != nil {
						log.Warn("push to ", w.receiver[i].Name(), " : ", err)
					} else {
						log.WithField("@cost", time.Since(start)).Info("push to ", w.receiver[i].Name())
					}
				}(k)
			}
		}(v)
	}
}

func checkRegexp(v feed.Item, reg []setting.Reg) (bool, string) {
	for _, r := range reg {
		// if r.R.MatchString(v.Description) {
		// 	return true, r.C
		// }
		if r.R.MatchString(v.Title) || r.R.MatchString(v.Author) {
			return true, r.C
		}
	}
	return false, ""
}

func (w *worker) checkTLength(data []byte) (bool, int64) {
	defer func() { _ = recover() }()

	result, err := bencode.Decode(data)
	if err != nil {
		return false, -1
	}
	info := result[0].Dict("info")
	pl := (info.Dict("piece length")).Value()
	ps := int64(len((info.Dict("pieces")).BStr())) / 20
	length := pl * ps

	if w.min <= length && length <= w.max {
		return true, length
	}
	return false, length
}
