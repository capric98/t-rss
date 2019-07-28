package client

import (
	"log"
	"net/http"
	"time"
)

type QBType struct {
	client *http.Client
}

func NewqBclient(m map[interface{}]interface{}) ClientType {
	var nc ClientType
	nc.Settings = make(map[string]string)

	for k, v := range m {
		nc.Settings[k.(string)] = v.(string)
	} // Copy settings.
	nc.Client = QBType{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	fcount := 1
	err := nc.Client.(QBType).Init()
	for err != nil {
		fcount++
		if fcount == 3 {
			log.Fatal(err)
		}
		err = nc.Client.(QBType).Init()
	}
	return nc
}

func (c QBType) Init() error {
	return nil
}

func (c QBType) Add(data []byte) error {
	return nil
}
