package myfeed

import (
	"errors"
	"io"
	"io/ioutil"
)

var (
	ErrNotRSSFormat  = errors.New("Feed: Not a RSS format feed.")
	ErrNotAtomFormat = errors.New("Feed: Not a Atom format feed.")

	items = []Item{}
)

func Parse(r io.ReadCloser) (f *Feed, e error) {
	data, _ := ioutil.ReadAll(r)
	f, e = rParse(data)
	if e == ErrNotRSSFormat {
		f, e = aParse(data)
	}
	return
}
