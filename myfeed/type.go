package myfeed

import (
	"strings"
	"time"
)

const (
	RSSType = iota
	AtomType
)

type Feed struct {
	Type  int
	Items []Item
}

type Item struct {
	rItem
	PubDate time.Time
}

var (
	timeFormat = []string{time.ANSIC, time.UnixDate, time.RubyDate,
		time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano,
		time.RFC822, time.RFC822Z, time.RFC850, time.Kitchen,
		time.Stamp, time.StampMicro, time.StampMilli, time.StampNano}
)

func strToTime(s string) time.Time {
	var t time.Time
	var e error
	for k := range timeFormat {
		t, e = time.Parse(timeFormat[k], s)
		if e == nil {
			break
		}
	}
	return t
}

func NameRegularize(name string) string {
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	name = strings.ReplaceAll(name, "\n", "_")
	name = strings.ReplaceAll(name, "\r", "_")
	name = strings.ReplaceAll(name, " ", "_")
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}
