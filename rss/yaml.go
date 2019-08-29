package rss

import (
	"log"
	"regexp"

	"gopkg.in/yaml.v2"
)

type confYaml struct {
	RSSLink     string `yaml:"rss"`
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
	Strict      bool
	Interval    int
	Download_to string

	Min, Max       int64
	Accept, Reject []*regexp.Regexp
	Client         []string
}

func convert(s string) int64 {
	return 0
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
		log.Prinln("Failed to parse config file:")
		log.Fatal(err)
		return nil
	}

	conf = make([]Config, 0)
	for k, v := range m {
		tmp := Config{
			TaskName:    k,
			RSSLink:     v.RSSLink,
			Strict:      v.Strict,
			Interval:    v.Interval,
			Download_to: v.Download_to,
			Min:         convert(v.Content_size.Min),
			Max:         convert(v.Content_size.Max),
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
