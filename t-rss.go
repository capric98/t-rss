package trss

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/capric98/t-rss/setting"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

// WithConfigFile starts program using a config file.
func WithConfigFile(filename string, learn bool) {
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

	fr, e := os.Open(filename)
	if e != nil {
		backgroundLogger.Fatal("open config file:", e)
	}
	config, e := setting.Parse(fr)
	fr.Close()
	if e != nil {
		backgroundLogger.Fatal("parse config file:", e)
	}
	if config.Global.LogConfig.Save == "" {
		formatter.ForceColors = true
		backgroundLogger.SetOutput(colorable.NewColorableStderr())
	} else {
		fw, fe := os.OpenFile(
			config.Global.LogConfig.Save,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0640,
		)
		if fe != nil {
			backgroundLogger.Fatal(fe)
		}
		backgroundLogger.SetOutput(fw)
	}
	backgroundLogger.SetLevel(toLogLevel(config.Global.LogConfig.Level))

	checkAndWatchHistory(
		config.Global.History.Save,
		config.Global.History.MaxAge.T,
		backgroundLogger,
	)

	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	backgroundLogger.Info("receive signal: ", <-c)
	_ = backgroundLogger.Writer().Close()
}
