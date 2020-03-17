package myfeed

import (
	"errors"
	"io"
)

var (
	ErrNotRSSFormat  = errors.New("Feed: Not a RSS format feed.")
	ErrNotAtomFormat = errors.New("Feed: Not an Atom format feed.")
)

func Parse(r io.Reader, ftype int) (f []Item, e error) {
	if ftype == RSSType {
		return rParse(r)
	}
	return aParse(r)
}
