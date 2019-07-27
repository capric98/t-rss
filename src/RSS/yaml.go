package RSS

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
)

type ClientType struct {
	Name     string
	Host     string
	Port     int
	UserName string
	Password string
}

type TaskType struct {
	TaskName  string
	RSS_Link  string
	Interval  int
	ExeAtTime []int
	DownPath  string
	Client    ClientType
	Cookie    string
	MaxSize   int
	MinSize   int
	Strict    bool
	AccRegexp []string
	RjcRegexp []string
}

func ParseClientSettings(s map[interface{}]interface{}) ClientType {
	var ps ClientType
	ps.Name = s["name"].(string)
	ps.Host = s["host"].(string)
	ps.UserName = s["user"].(string)
	ps.Password = s["pass"].(string)
	return ps
}

func ConfigCheck(ts []TaskType) []TaskType {
	for i := 0; i < len(ts); i++ {
		if ts[i].Interval <= 0 {
			log.Printf("Task %s misses Interval, no bother to set it to default 60s.\n", ts[i].TaskName)
			ts[i].Interval = 60
		}
		if ts[i].Interval <= 3 {
			log.Printf("Caution: Task %s has too low Interval of %ds\n", ts[i].TaskName, ts[i].Interval)
		}
	}
	return ts
}

func ParseSettings(data []byte) []TaskType {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	T := make([]TaskType, len(m))
	n := 0
	//fmt.Printf("--- m:\n%v\n", m)
	for tname, task := range m {
		T[n].TaskName = tname.(string)
		for k, v := range task.(map[interface{}]interface{}) {
			switch k.(string) {
			case "rss":
				T[n].RSS_Link = v.(string)
			case "download_to":
				T[n].DownPath = v.(string)
			case "client":
				T[n].Client = ParseClientSettings(v.(map[interface{}]interface{}))
			case "regexp":
				// We'd better check the validity of regexps after...

				if tmp := v.(map[interface{}]interface{})["acceptFilter"]; tmp != nil {
					for _, r := range tmp.([]interface{}) {
						T[n].AccRegexp = append(T[n].AccRegexp, r.(string))
					}
				}
				if tmp := v.(map[interface{}]interface{})["rejectFilter"]; tmp != nil {
					for _, r := range tmp.([]interface{}) {
						T[n].RjcRegexp = append(T[n].RjcRegexp, r.(string))
					}
				}
			case "content_size":
				if tmp := v.(map[interface{}]interface{})["max"]; tmp != nil {
					T[n].MaxSize = tmp.(int)
				}
				if tmp := v.(map[interface{}]interface{})["min"]; tmp != nil {
					T[n].MinSize = tmp.(int)
				}
			case "strict":
				T[n].Strict = v.(bool)
			default:
				log.Printf("Caution: Unknown config path: %s\n", k.(string))
			}
		}
		fmt.Println(T[n])
		n++
	}

	return ConfigCheck(T)
}
