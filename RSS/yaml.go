package RSS

import (
	"log"
	"regexp"

	"github.com/capric98/GoRSS/client"
	"gopkg.in/yaml.v2"
)

type ClientType = client.ClientType

type TaskType struct {
	TaskName  string
	RSS_Link  string
	Interval  int
	ExeAtTime []int
	DownPath  string
	Client    []ClientType
	Cookie    string
	MaxSize   int64
	MinSize   int64
	Strict    bool
	AccRegexp []*regexp.Regexp
	RjcRegexp []*regexp.Regexp
}

func ParseClientSettings(s map[interface{}]interface{}) []ClientType {
	ps := make([]ClientType, 0)
	for k, v := range s {
		switch k.(string) {
		case "qBittorrent":
			ps = append(ps, client.NewqBclient(v.(map[interface{}]interface{})))
		case "Deluge":
			ps = append(ps, client.NewDeClient(v.(map[interface{}]interface{})))
		default:
		}
	}
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
		T[n].MinSize = -1
		T[n].MaxSize, _ = client.SpeedToInt("2000TB") // 2000TiB+
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
						re, err := regexp.Compile(r.(string))
						if err != nil {
							log.Fatalf("Panic: %v\n", err)
						}
						T[n].AccRegexp = append(T[n].AccRegexp, re)
					}
				}
				if tmp := v.(map[interface{}]interface{})["rejectFilter"]; tmp != nil {
					for _, r := range tmp.([]interface{}) {
						re, err := regexp.Compile(r.(string))
						if err != nil {
							log.Fatalf("Panic: %v\n", err)
						}
						T[n].RjcRegexp = append(T[n].RjcRegexp, re)
					}
				}
			case "content_size":
				if tmp := v.(map[interface{}]interface{})["max"]; tmp != nil {
					switch tmp.(type) {
					case int:
						T[n].MaxSize = int64(tmp.(int)) * 1024 * 1024
					case string:
						T[n].MaxSize, _ = client.SpeedToInt(tmp.(string))
					}

				}
				if tmp := v.(map[interface{}]interface{})["min"]; tmp != nil {
					switch tmp.(type) {
					case int:
						T[n].MinSize = int64(tmp.(int)) * 1024 * 1024
					case string:
						T[n].MinSize, _ = client.SpeedToInt(tmp.(string))
					}
				}
			case "strict":
				T[n].Strict = v.(bool)
			case "interval":
				T[n].Interval = v.(int)
			default:
				log.Printf("Caution: Unknown config path: %s\n", k.(string))
			}
		}
		//fmt.Println(T[n].MinSize, T[n].MaxSize)
		n++
	}

	return ConfigCheck(T)
}
