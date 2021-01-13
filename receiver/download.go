package receiver

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/capric98/t-rss/feed"
)

type dReceiver struct {
	path string
}

// NewDownload news a download receiver
func NewDownload(path string) Receiver {
	if _, e := os.Stat(path); os.IsNotExist(e) {
		_ = os.MkdirAll(path, 0740)
	}

	return &dReceiver{path: path}
}

// Push implements Receiver interface.
func (r *dReceiver) Push(i *feed.Item, b []byte) (e error) {
	fn := i.Title
	fn = regularizeFilename(fn)
	e = ioutil.WriteFile(path.Join(r.path, fn+".torrent"), b, 0664)
	return
}

// Name implements Receiver interface.
func (r *dReceiver) Name() string {
	return "download"
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

	nameRune := []rune(name)
	for len(string(nameRune)) > 200 {
		nameRune = nameRune[:len(nameRune)-1]
	}
	return string(nameRune)
}
