package myfeed

import (
	"errors"
)

var (
	ErrNotRSSFormat  = errors.New("Feed: Not a RSS format feed.")
	ErrNotAtomFormat = errors.New("Feed: Not a Atom format feed.")
)

func Parse(data []byte) (f Feed, e error) {
	f, e = rParse(data)
	if e == ErrNotRSSFormat {
		f, e = aParse(data)
	}
	return
}
