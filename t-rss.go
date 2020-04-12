package trss

import (
	"os"

	"github.com/capric98/t-rss/setting"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

// WithConfigFile starts program using a config file.
func WithConfigFile(filename string, learn bool) {
	backgroundLogger := logrus.New()
	formatter := &logrus.TextFormatter{
		ForceColors:     false,
		FullTimestamp:   true,
		PadLevelText:    true,
		TimestampFormat: "2006-01-02 15:04:05",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@time",
			logrus.FieldKeyLevel: "level",
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
		backgroundLogger.WithField("testField", "!!!").Info("Output log to stderr...")
		backgroundLogger.SetOutput(colorable.NewColorableStderr())
	}
	backgroundLogger.SetLevel(logrus.DebugLevel)
	backgroundLogger.Debug(config)
	backgroundLogger.Info("hahahaha")
}
