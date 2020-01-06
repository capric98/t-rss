package myfeed

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestAtomFeed(t *testing.T) {
	data, e := ioutil.ReadFile("1.xml")
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	r := bytes.NewReader(data)
	result, e := aParse(ioutil.NopCloser(r))
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
	fmt.Println(len(result))
	for _, v := range result {
		fmt.Println(v)
	}
}
