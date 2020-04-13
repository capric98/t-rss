package bencode

import (
	"crypto/sha1"
	"errors"
)

var (
	ErrInfoNotFound = errors.New("bencode: Did not find Info to calc infohash.")
)

func (body *Body) Infohash() (r []byte, e error) {
	info := body.Dict("info")
	if info == nil {
		e = ErrInfoNotFound
		return
	}

	hasher := sha1.New()
	data, e := info.Encode()
	if e != nil {
		return
	}
	_, e = hasher.Write(data)
	if e != nil {
		return
	}

	r = hasher.Sum(nil)
	return
}
