package rss

import (
	"log"
	"regexp"
	"time"
	"unicode"

	"github.com/capric98/t-rss/client"
	"gopkg.in/yaml.v2"
)

type confYaml struct {
	RSSLink     string `yaml:"rss"`
	Cookie      string `yaml:"cookie"`
	Strict      bool   `yaml:"strict"`
	Interval    int    `yaml:"interval"`
	Download_to string `yaml:"download_to"`

	Content_size struct {
		Min string `yaml:"min"`
		Max string `yaml:"max"`
	} `yaml:"content_size"`
	Regexp struct {
		Accept []string `yaml:"accept"`
		Reject []string `yaml:"reject"`
	} `yaml:"regexp"`
	Client struct {
		Qb map[string]interface{} `yaml:"qBittorrent"`
		De map[string]interface{} `yaml:"Deluge"`
	} `yaml:"client"`
}

type Config struct {
	TaskName    string
	RSSLink     string
	Cookie      string
	Strict      bool
	Interval    time.Duration
	Download_to string

	Min, Max       int64
	Accept, Reject []*regexp.Regexp
	Client         []client.Client
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

func regcompile(s string) *regexp.Regexp {
	r, e := regexp.Compile(s)
	if e != nil {
		log.Println("Failed to build regexp:", s)
		log.Fatal(e)
	}
	return r
}

func parse(data []byte) (conf []Config) {
	m := make(map[string]confYaml)
	if err := yaml.Unmarshal(data, &m); err != nil {
		log.Println("Failed to parse config file:")
		log.Fatal(err)
		return nil
	}

	conf = make([]Config, 0)
	for k, v := range m {
		tmp := Config{
			TaskName:    k,
			RSSLink:     v.RSSLink,
			Cookie:      v.Cookie,
			Strict:      v.Strict,
			Interval:    time.Duration(v.Interval) * time.Second,
			Download_to: v.Download_to,
			Min:         UConvert(v.Content_size.Min),
			Max:         UConvert(v.Content_size.Max),
		}
		if tmp.Max == 0 {
			tmp.Max = 0x7FFFFFFFFFFFFFFF
		}
		if v.Regexp.Accept != nil {
			tmp.Accept = make([]*regexp.Regexp, len(v.Regexp.Accept))
			for i, r := range v.Regexp.Accept {
				tmp.Accept[i] = regcompile(r)
			}
		}
		if v.Regexp.Reject != nil {
			tmp.Reject = make([]*regexp.Regexp, len(v.Regexp.Reject))
			for i, r := range v.Regexp.Reject {
				tmp.Accept[i] = regcompile(r)
			}
		}
	}
	return
}
