package trss

import "github.com/sirupsen/logrus"

func toLogLevel(s string) (l logrus.Level) {
	switch s {
	case "trace":
		l = logrus.TraceLevel
	case "debug":
		l = logrus.DebugLevel
	case "info":
		l = logrus.InfoLevel
	case "warn":
		l = logrus.WarnLevel
	default:
		l = logrus.InfoLevel
	}
	return
}
