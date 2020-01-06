package myfeed

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestRSSFeed(t *testing.T) {
	data, e := ioutil.ReadFile("1.xml")
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	r := bytes.NewReader(data)
	result, e := rParse(ioutil.NopCloser(r))
	if e != nil {
		fmt.Println(e)
		//t.Fail()
	}
	fmt.Println(len(result))
	for range result {
		//fmt.Println(v)
	}
}
