package core

import (
	"net/http"
	"regexp"
	"time"

	"github.com/capric98/t-rss/client"
	"github.com/capric98/t-rss/torrents"
)

type worker struct {
	name     string
	loglevel int
	Config   Conf
	ticker   chan []torrents.Individ
	client   *http.Client
	cancel   func()
}
type clientConfig = map[string]interface{}

type ymlConf struct {
	RSSLink     string `yaml:"rss"`
	Cookie      string `yaml:"cookie"`
	EditTracker struct {
		Delete []string `yaml:"delete"`
		Add    []string `yaml:"add"`
	} `yaml:"edit_tracker"`
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
	RSSLink     string
	Cookie      string
	Strict      bool
	Interval    time.Duration
	Latency     time.Duration
	Download_to string

	Min int64
	Max int64

	Accept  []*regexp.Regexp
	Reject  []*regexp.Regexp
	DeleteT []*regexp.Regexp
	AddT    []string
	Client  []client.Client
}

var (
	DMode, TestOnly, Learn bool
	ConfigPath, CDir       string
)
