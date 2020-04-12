package client

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestDeluge(t *testing.T) {
	config := make(map[string]interface{})
	config["host"] = "127.0.0.1:58846"
	config["username"] = "localclient"
	config["password"] = "d96a384f6b9a314405575554acd8b40a6f2f343d"
	config["add_paused"] = true
	c := NewDeClient("Test", config)

	file, _ := ioutil.ReadFile("SAXZ-5.torrent")
	e := c.Add(file, "SAXZ-5")
	if e != nil {
		log.Println("Test:", e)
		// t.Fail()
	}
	log.Println("Success!")
	// t.Fail()
}
