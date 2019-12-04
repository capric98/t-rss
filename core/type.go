package core

import (
	"github.com/capric98/t-rss/torrents"
)

type worker struct{
	loglevel int
	Config ymlConf
	ticker chan []torrents.Individ
}
type clientConfig = map[string]interface{}

type ymlConf struct {
	RSSLink     string `yaml:"rss"`
	Cookie      string `yaml:"cookie"`
	Strict      bool   `yaml:"strict"`
	Interval    int    `yaml:"interval"`
	Latency     int    `yaml:"latency"`
	Download_to string `yaml:"download_to"`

	Content_size struct {
		Min string `yaml:"min"`
		Max string `yaml:"max"`
	} `yaml:"content_size"`
	Regexp struct {
		Accept []string `yaml:"accept"`
		Reject []string `yaml:"reject"`
	} `yaml:"regexp"`
	Client map[string]clientConfig `yaml:"client"`
}

type Conf struct {
	RSSLink string
	Cookie string
	Strict bool
	Interval    int
	Latency     int
	Download_to string

	Content_size struct {
		Min string
		Max string
	}
	Regexp struct {
		Accept []string
		Reject []string
	}
	Client map[string]clientConfig
}

var (
	DMode,TestOnly bool
	ConfigPath, CDir string
)