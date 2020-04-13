package trss

import (
	"context"
	"net/http"
	"sync"

	"github.com/capric98/t-rss/feed"
	"github.com/capric98/t-rss/setting"
	"github.com/capric98/t-rss/ticker"
	"github.com/sirupsen/logrus"
)

type worker struct {
	*ticker.Ticker
	*http.Client

	logger *logrus.Entry
}

func doTask(ctx context.Context, t setting.Task, client *http.Client, logger *logrus.Entry, wg *sync.WaitGroup) {
}

func (w *worker) do(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case items, ok := <-w.C():
			if !ok {
				return
			}
			var passed []feed.Item
			for _, v := range items {
				passed = append(passed, v)
			}
		}
	}
}
