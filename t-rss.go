package trss

import (
	"context"
	"net/http"
	"os"
	"os/signal"
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

	checkAndWatchHistory(
		config.Global.History.Save,
		config.Global.History.MaxAge.T,
		backgroundLogger,
	)

	client := &http.Client{Timeout: config.Global.Timeout.T}
	bgCtx, cancel := context.WithCancel(context.Background())
	for k, v := range config.Tasks {
		doTask(bgCtx, v, client, backgroundLogger.WithField("task", k))
	}

	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	backgroundLogger.Info("receive signal: ", <-c)
	cancel()
	_ = backgroundLogger.Writer().Close()
}
