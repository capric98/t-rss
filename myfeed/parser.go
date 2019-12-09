package myfeed

import (
	"errors"
	"io"
)

var (
	ErrNotRSSFormat  = errors.New("Feed: Not a RSS format feed.")
	ErrNotAtomFormat = errors.New("Feed: Not a Atom format feed.")

	items = []Item{}
)

func Parse(r io.ReadCloser, ftype int) (f []Item, e error) {
	if ftype == RSSType {
		return rParse(r)
	}
	return aParse(r)
}
