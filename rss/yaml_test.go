package rss

import (
	"fmt"
	"testing"

	"reflect"

	"gopkg.in/yaml.v2"
)

var yamlString = `
Name0:
  rss: https://bangumi.tv/latest/rss.xml
  strict: no
  content_size:
    min: 2048
    max: 9999
  regexp:
    accept:
      - Vol.*?Fin
    reject:
      - Test
  interval: 10
  download_to: "/home/WatchDir/"
  client:
    qBittorrent:
      host: http://127.0.0.1:8080
      username: admin
      password: adminadmin
      dlLimit: "10M"
      upLimit: "10M"
      paused: true
      savepath: "/home/Downloads/"`

func TestDecodeYaml(t *testing.T) {
	con := make(map[string]confYaml)
	err := yaml.Unmarshal([]byte(yamlString), &con)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	v := reflect.ValueOf(con["Name0"])
	for i := 0; i < v.Type().NumField(); i++ {
		vt := v.Type().Field(i)
		fmt.Printf("%v: ", vt.Tag.Get("yaml"))
		fmt.Println(v.Field(i))
	}
	t.Fail()
}
