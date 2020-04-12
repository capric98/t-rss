package setting

import (
	"io"
	"regexp"
	"time"

	"gopkg.in/yaml.v2"
)

// C :)
type C struct {
	Global Global          `yaml:"GLOBAL"`
	Tasks  map[string]Task `yaml:"TASKS"`
}

// Global is global configs.
type Global struct {
	LogConfig struct {
		Level string `yaml:"level"`
		Save  string `yaml:"save_to"`
	} `yaml:"log"`
	History struct {
		MaxAge time.Duration `yaml:"max_age"`
		Save   string        `yaml:"save_to"`
	} `yaml:"history"`
	Timeout time.Duration `yaml:"timeout"`
}

// Task is task part.
type Task struct {
	Rss      *Rss     `yaml:"rss"`
	Filter   Filter   `yaml:"filter"`
	Quota    Quota    `yaml:"quota"`
	Receiver Receiver `yaml:"receiver"`
}

// Rss :)
type Rss struct {
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
}

// Filter :)
type Filter struct {
	ContentSize ContentSize  `yaml:"content_size"`
	Regexp      RegexpConfig `yaml:"regexp"`
}

// Quota :)
type Quota struct {
	Num  int   `yaml:"num"`
	Size int64 `yaml:"size"`
}

// Edit :)
type Edit struct {
	Tracker Tracker `yaml:"tracker"`
}

// Receiver defines tasks' receiver(s).
type Receiver struct {
	Save   string                            `yaml:"save_to"`
	Client map[string]map[string]interface{} `yaml:"client"`
}

// ContentSize :)
type ContentSize struct {
	Min int64 `yaml:"min"`
	Max int64 `yaml:"max"`
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
	return
}
