package main

import (
	"RSS"
	"fmt"
	"io/ioutil"
)

func main() {
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		fmt.Println(err)
		return
	}
	RSS.ParseSettings(data)
}
