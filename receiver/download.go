package receiver

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type dReceiver struct {
	path string
}

// NewDownload news a download receiver
func NewDownload(path string) Receiver {
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	return &dReceiver{path: path}
}

// Push implements Receiver interface.
func (r *dReceiver) Push(b []byte, i interface{}) (e error) {
	fn, ok := i.(string)
	if !ok {
		return fmt.Errorf("expected a string but got %T", i)
	}
	fn = regularizeFilename(fn)
	e = ioutil.WriteFile(r.path+fn, b, 0664)
	return
}

func regularizeFilename(name string) string {
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	name = strings.ReplaceAll(name, "\n", "_")
	name = strings.ReplaceAll(name, "\r", "_")
	name = strings.ReplaceAll(name, " ", "_")
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}
