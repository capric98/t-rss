package trss

import (
	"context"
	"net/http"

	"github.com/capric98/t-rss/setting"
	"github.com/capric98/t-rss/ticker"
	"github.com/sirupsen/logrus"
)

type worker struct {
	ticker.Ticker
	*http.Client

	logger *logrus.Entry
}

func doTask(ctx context.Context, t setting.Task, client *http.Client, logger *logrus.Entry) {
}

func (w *worker) do(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}
