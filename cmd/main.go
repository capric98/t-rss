package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	config = flag.String("conf", "config.yml", "config file")
)

func init() {
	flag.Parse()
}

func main() {
	hd, e := os.UserHomeDir()
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(hd)
	if _, e := os.Stat(hd + "/audio.m4a"); !os.IsNotExist(e) {
		fmt.Println("file exists!")
	} else {
		fmt.Println(e)
	}
}
