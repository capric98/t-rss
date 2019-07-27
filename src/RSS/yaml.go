package RSS

import (
	"fmt"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"
)

type TaskType struct {
	TaskName  string
	RSS_Link  string
	Interval  int
	ExeAtTime []int
	DownPath  string
	Client    string
	Cookie    string
}

func ParseSettings(data []byte) ([]TaskType, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	T := make([]TaskType, len(m))
	n := 0
	//fmt.Printf("--- m:\n%v\n", m)
	for tname, task := range m {
		T[n].TaskName = tname.(string)
		for k, v := range task.(map[interface{}]interface{}) {
			if k.(string) == "rss" {
				RssFetch(v.(string), &http.Client{
					Timeout: time.Duration(10 * time.Second),
				})
			}
			//fmt.Println(v)
		}
		n++
	}

	return nil, nil
}
