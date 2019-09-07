package bencode

import (
	"fmt"
	"testing"
)

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}
func TestEncode(t *testing.T) {
	e := NewEncoder()
	_ = e.NewDict("")
	_ = e.Add("name", "Tadokoro Kouji")
	_ = e.Add("age", 24)
	_ = e.NewList("List Test")
	check(e.Add("", "Line0"))
	check(e.Add("", "Line1"))
	check(e.NewDict(""))
	check(e.Add("Ooops", "bilibili"))
	check(e.EndPart())
	check(e.EndPart())
	check(e.Add("A", "Add to head."))
	check(e.Add("X", 114514))
	result := e.End()
	result[0].print(0)
	result[0].Delete("A")
	result[0].Dict("X").Edit(1919810)
	check(result[0].AddPart("Copy", result[0].Dict("name")))
	result[0].Dict("name").Edit("Yajuu Senpai")
	result[0].print(0)
	//t.Fail()
}
