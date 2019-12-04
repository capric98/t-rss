package core

import (
	"github.com/capric98/t-rss/client"
	"net/http"
	"time"
	"regexp"
	"github.com/capric98/t-rss/torrents"
)

type worker struct{
	name string
	loglevel int
	Config Conf
	ticker chan []torrents.Individ
	client *http.Client
	cancel func()
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
	Interval    time.Duration
	Latency     time.Duration
	Download_to string

	Min int64
	Max int64

	Accept []*regexp.Regexp
	Reject []*regexp.Regexp
	Client []client.Client
}

var (
	DMode,TestOnly,Learn bool
	ConfigPath, CDir string
)