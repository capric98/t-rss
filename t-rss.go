package trss

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/capric98/t-rss/setting"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

// WithConfigFile starts program using a config file.
func WithConfigFile(filename string, level string, learn bool) {
	backgroundLogger := logrus.New()
	formatter := &logrus.TextFormatter{
		ForceColors:      false,
		ForceQuote:       true,
		FullTimestamp:    true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		TimestampFormat:  "2006-01-02 15:04:05",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@time",
			logrus.FieldKeyLevel: "&",
			logrus.FieldKeyMsg:   "@msg",
		},
	}
	backgroundLogger.SetFormatter(formatter)
	backgroundLogger.SetLevel(toLogLevel(level))

	fr, e := os.Open(filename)
	if e != nil {
		backgroundLogger.Fatal("open config file: ", e)
	}
	config, e := setting.Parse(fr)
	fr.Close()
	if e != nil {
		backgroundLogger.Fatal("parse config file: ", e)
	}

	backgroundLogger.Tracef("%#v\n", *config)

	if config.Global.LogFile == "" {
		formatter.ForceColors = true
		backgroundLogger.SetOutput(colorable.NewColorableStderr())
	} else {
		fw, fe := os.OpenFile(
			config.Global.LogFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0640,
		)
		if fe != nil {
			backgroundLogger.Fatal(fe)
		}
		backgroundLogger.SetOutput(fw)
	}

	if config.Global.History.Save[len(config.Global.History.Save)-1] != '/' {
		config.Global.History.Save = config.Global.History.Save + "/"
	}
	checkAndWatchHistory(
		config.Global.History.Save,
		config.Global.History.MaxNum,
		backgroundLogger,
	)

	client := &http.Client{Timeout: config.Global.Timeout.T}
	bgCtx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	runNum := -1
	if learn {
		backgroundLogger.Info("Learning...")
		runNum = 1
	}
	for k, v := range config.Tasks {
		kk := k // make a copy
		wg.Add(1)

		nw := &worker{
			client: client,
			header: v.Rss.Headers.H,
			tick:   nil, // https://github.com/capric98/t-rss/blob/master/task.go#L40

			filters: nil, // https://github.com/capric98/t-rss/blob/master/task.go#L50
			recvers: nil, // https://github.com/capric98/t-rss/blob/master/task.go#L54
			quota:   v.Quota,
			delay:   v.Receiver.Delay.T,
			wpath:   config.Global.History.Save + kk + "/",

			edit: v.Edit,

			logger: func() *logrus.Entry {
				return backgroundLogger.WithField("@task", kk)
			},
			ctx: bgCtx,
			wg:  &wg,
		}
		// backgroundLogger.Debugf("%#v\n", nw.header)
		nw.prepare(v, runNum)
		go nw.loop()
	}

	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	wgDone := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		wgDone <- struct{}{}
	}()

	select {
	case sig := <-c:
		backgroundLogger.Info("receive signal: ", sig)
	case <-wgDone:
		backgroundLogger.Info("all tasks done")
	}

	cancel()
	backgroundLogger.Info("gracefully shutting down...")
	wg.Wait()
	backgroundLogger.Info("bye~")
	_ = backgroundLogger.Writer().Close()
}
