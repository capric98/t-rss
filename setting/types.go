package setting

import (
	"io"
	"regexp"
	"time"

	"github.com/capric98/t-rss/unit"
	"gopkg.in/yaml.v2"
)

// Int64 is int64
type Int64 struct {
	I int64
}

// Duration wraps time.Duration
type Duration struct {
	T time.Duration
}

// C :)
type C struct {
	Global Global           `yaml:"GLOBAL"`
	Tasks  map[string]*Task `yaml:"TASKS"`
}

// Global is global configs.
type Global struct {
	LogFile string `yaml:"log_file"`
	History struct {
		MaxNum int    `yaml:"max_num"`
		Save   string `yaml:"save_to"`
	} `yaml:"history"`
	Timeout Duration `yaml:"timeout"`
}

// Task is task part.
type Task struct {
	Rss      *Rss     `yaml:"rss"`
	Filter   Filter   `yaml:"filter"`
	Quota    Quota    `yaml:"quota"`
	Edit     *Edit    `yaml:"edit"`
	Receiver Receiver `yaml:"receiver"`
}

// Rss :)
type Rss struct {
	URL      string            `yaml:"url"`
	Method   string            `yaml:"method"`
	Headers  map[string]string `yaml:"headers"`
	Interval Duration          `yaml:"interval"`
}

// Filter :)
type Filter struct {
	ContentSize ContentSize  `yaml:"content_size"`
	Regexp      RegexpConfig `yaml:"regexp"`
}

// Quota :)
type Quota struct {
	Num  int   `yaml:"num"`
	Size Int64 `yaml:"size"`
}

// Edit :)
type Edit struct {
	Tracker Tracker `yaml:"tracker"`
}

// Receiver defines tasks' receiver(s).
type Receiver struct {
	Delay  Duration                          `yaml:"delay"`
	Save   *string                           `yaml:"save_path"`
	Client map[string]map[string]interface{} `yaml:"client"`
}

// ContentSize :)
type ContentSize struct {
	Min Int64 `yaml:"min"`
	Max Int64 `yaml:"max"`
}

// RegexpConfig :)
type RegexpConfig struct {
	Accept []Reg `yaml:"accept"`
	Reject []Reg `yaml:"reject"`
}

// Reg :)
type Reg struct {
	R *regexp.Regexp
	C string
}

// Tracker :)
type Tracker struct {
	Delete []Reg    `yaml:"delete"`
	Add    []string `yaml:"add"`
}

// Parse :)
func Parse(r io.Reader) (config *C, e error) {
	config = new(C)
	e = yaml.NewDecoder(r).Decode(config)
	if e != nil {
		return
	}
	config.standardize()
	return
}

func (c *C) standardize() {
	if c.Global.Timeout.T == 0 {
		c.Global.Timeout.T = 30 * time.Second
	}
	if c.Global.History.MaxNum == 0 {
		c.Global.History.MaxNum = 500
	}
	if c.Global.History.Save == "" {
		c.Global.History.Save = ".t-rss_History/"
	}
	if l := len(c.Global.History.Save); l > 0 && c.Global.History.Save[l-1] != '/' {
		c.Global.History.Save = c.Global.History.Save + "/"
	}
	for _, v := range c.Tasks {
		if v.Rss.Method == "" {
			v.Rss.Method = "GET"
		}
		if v.Rss.Interval.T == 0 {
			v.Rss.Interval.T = 30 * time.Second
		}
		if v.Filter.ContentSize.Max.I == 0 {
			v.Filter.ContentSize.Max.I = 1 << 62
		}
		if v.Quota.Num == 0 {
			v.Quota.Num = 1 << 30
		}
		if v.Quota.Size.I == 0 {
			v.Quota.Size.I = 1 << 62
		}
	}
}

// UnmarshalYAML :)
func (t *Duration) UnmarshalYAML(uf func(interface{}) error) (e error) {
	var s string
	e = uf(&s)
	if e != nil {
		return
	}
	t.T = unit.ParseDuration(s)
	return nil
}

// UnmarshalYAML :)
func (r *Reg) UnmarshalYAML(uf func(interface{}) error) (e error) {
	var s string
	e = uf(&s)
	if e != nil {
		return
	}
	r.C = s
	r.R, e = regexp.Compile(s)
	return
}

// UnmarshalYAML :)
func (n *Int64) UnmarshalYAML(uf func(interface{}) error) (e error) {
	var s string
	e = uf(&s)
	if e != nil {
		return
	}
	n.I = unit.ParseSize(s)
	e = nil
	return
}
