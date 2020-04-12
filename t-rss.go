package trss

import (
	"os"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

// WithConfigFile starts program using a config file.
func WithConfigFile(filename string, learn bool) {
	backgroundLogger := logrus.New()
	backgroundLogger.SetOutput(colorable.NewColorableStdout())

	fr, e := os.Open(filename)
	if e != nil {
		backgroundLogger.Fatal("open config file:", e)
	}
	defer fr.Close()
}
