package setting

import (
	"fmt"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	fr, e := os.Open("../config.example.yml")
	if e != nil {
		fmt.Println(e)
		t.FailNow()
	}
	defer fr.Close()
	c, e := Parse(fr)
	if e != nil {
		fmt.Println(e)
		t.FailNow()
	}
	fmt.Printf("%#v\n", c)
}
