package bencode

import (
	"fmt"
	"testing"
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
	switch b.Type() {
	case IntValue:
		fmt.Println(b.Value())
	case ByteString:
		if len(b.BStr()) < 250 {
			fmt.Println(string(b.BStr()))
		} else {
			fmt.Println("[...Too long]")
		}
	case DictType:
		fmt.Println("[Dictionary]")
		for i := 0; i < b.Len(); i++ {
			printSpace(level + 1)
			k, _ := b.DictN(i)
			fmt.Println(k + ":")
			v := b.Dict(k)
			v.print(level + 2)
		}
	case ListType:
		fmt.Println("[List]")
		for i := 0; i < b.Len(); i++ {
			b.List(i).print(level + 1)
		}
	}
}

func (b *Body) idle() {
}

func TestDecodeByteSlice(t *testing.T) {
	// startT := time.Now()
	// f, _ := ioutil.ReadFile("vcb.torrent")
	// result, err := Decode(f)
	// if err != nil {
	// 	fmt.Println(err)
	// 	t.Fail()
	// }
	// fmt.Printf("%s\n", time.Since(startT))
	// result[0].print(0)

	// info := result[0].Dict("info")
	// pl := (info.Dict("piece length")).Value()
	// ps := int64(len((info.Dict("pieces")).BStr())) / 20
	// //ps := int64(1)
	// fmt.Println(float64(pl*ps) / 1024 / 1024 / 1024)
	// fmt.Println("Checked:", result[0].Check())

	// hash, err := result[0].Infohash()
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(hex.EncodeToString(hash))
	// }

	//enc, _ := result[0].Encode()
	//_ = ioutil.WriteFile("out", enc, 0644)
	//t.Fail()
}

func BenchmarkDecode(b *testing.B) {
	// f, _ := ioutil.ReadFile("TLMC.torrent")
	// b.ReportAllocs()
	// for i := 0; i < b.N; i++ {
	// 	result, err := Decode(f)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		b.Fail()
	// 	}
	// 	result[0].idle()
	// }
}
