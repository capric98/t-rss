package bencode

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func printSpace(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf("  ")
	}
}

func (b *Body) print(level int) {
	if b == nil {
		return
	}

	printSpace(level)
	switch b.Type {
	case IntValue:
		fmt.Println(b.Value)
	case ByteString:
		if len(b.ByteStr) < 250 {
			fmt.Println(string(b.ByteStr))
		} else {
			fmt.Println("[...Too long]")
		}
	case DictType:
		fmt.Println("[Dictionary]")
		for k := range b.Dict {
			printSpace(level + 1)
			fmt.Println(string(b.Dict[k].key) + ":")
			(b.Get(string(b.Dict[k].key))).print(level + 2)
			//b.Dict[k].value.print(level + 2)
		}
	case ListType:
		fmt.Println("[List]")
		for _, v := range b.List {
			v.print(level + 1)
		}
	}
}

func (b *Body) idle() {
}

func TestDecodeByteSlice(t *testing.T) {
	startT := time.Now()
	f, _ := ioutil.ReadFile("vcb.torrent")
	result, err := Decode(f)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Printf("%s\n", time.Since(startT))
	result[0].print(0)

	info := result[0].Get("info")
	pl := (info.Get("piece length")).Value
	ps := int64(len((info.Get("pieces")).ByteStr)) / 20
	//ps := int64(1)
	fmt.Println(float64(pl*ps) / 1024 / 1024 / 1024)
	//t.Fail()
}

func BenchmarkDecode(b *testing.B) {
	f, _ := ioutil.ReadFile("TLMC.torrent")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result, err := Decode(f)
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}
		result[0].idle()
	}
}
