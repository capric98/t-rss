package core

import (
	"log"
	"regexp"
	"time"
	"unicode"

	"github.com/capric98/t-rss/client"
	"gopkg.in/yaml.v2"
)

func parse(data []byte) (conf map[string]Conf) {
	tmp := make(map[string]ymlConf)
	conf = make(map[string]Conf)
	if err := yaml.Unmarshal(data, &tmp); err != nil {
		log.Println("Failed to parse config file:")
		log.Fatal(err)
	}

	for k, v := range tmp {
		tmp := Conf{
			RSSLink:     v.RSSLink,
			Cookie:      v.Cookie,
			Interval:    time.Duration(v.Interval) * time.Second,
			Latency:     time.Duration(v.Latency) * time.Second,
			Download_to: v.Download_to,
			Min:         UConvert(v.Content_size.Min),
			Max:         UConvert(v.Content_size.Max),
			Quota: Quota{
				Num:  v.Quota.Num,
				Size: UConvert(v.Quota.Size),
			},
			Accept:  regcompile(v.Regexp.Accept),
			Reject:  regcompile(v.Regexp.Reject),
			DeleteT: regcompile(v.EditTracker.Delete),
			AddT:    v.EditTracker.Add,
			Client:  parseClient(v.Client),
		}
		if v.Interval == 0 {
			tmp.Interval = 30 * time.Second
		}
		if tmp.Max == 0 {
			tmp.Max = 0x7FFFFFFFFFFFFFFF
		}
		if tmp.Quota.Size == 0 {
			tmp.Quota.Size = 0x7FFFFFFFFFFFFFFF
		}
		if tmp.Quota.Num == 0 {
			tmp.Quota.Num = 65535
		}
		conf[k] = tmp
	}
	return
}

func UConvert(s string) int64 {
	if s == "" {
		return 0
	}
	var spNum int64
	u := make([]rune, 0)
	for _, c := range s {
		if !unicode.IsDigit(c) {
			u = append(u, c)
		} else {
			spNum = spNum*10 + int64(c-'0')
		}
	}
	unit := string(u)
	switch {
	case unit == "K" || unit == "k" || unit == "KB" || unit == "kB" || unit == "KiB" || unit == "kiB":
		spNum = spNum * 1024
	case unit == "M" || unit == "m" || unit == "MB" || unit == "mB" || unit == "MiB" || unit == "miB":
		spNum = spNum * 1024 * 1024
	case unit == "G" || unit == "g" || unit == "GB" || unit == "gB" || unit == "GiB" || unit == "giB":
		spNum = spNum * 1024 * 1024 * 1024
	case unit == "T" || unit == "t" || unit == "TB" || unit == "tB" || unit == "TiB" || unit == "tiB":
		spNum = spNum * 1024 * 1024 * 1024 * 1024
	}
	return spNum
}

func regcompile(s []string) []Reg {
	if s == nil {
		return nil
	}

	rs := make([]Reg, 0, 1)

	for _, v := range s {
		r, e := regexp.Compile(v)
		if e != nil {
			log.Println("Failed to build regexp:", v)
			log.Fatal(e)
		}
		rs = append(rs, Reg{C: v, R: r})
	}

	return rs
}

func parseClient(raw map[string]clientConfig) (list []client.Client) {
	list = make([]client.Client, 0, 1)

	for k, v := range raw {
		if v["type"] == nil {
			log.Panicln("Invalid config: Client should have type attribute.")
		}
		if v["type"].(string) == "qBittorrent" {
			list = append(list, client.NewqBclient(k, v))
		}
		if v["type"].(string) == "Deluge" {
			list = append(list, client.NewDeClient(k, v))
		}
	}

	return list
}
