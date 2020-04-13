package feed

import (
	"errors"
	"time"

	"golang.org/x/net/html"
)

// Item :)
type Item struct {
	Title       string
	Link        string
	Description string
	Author      string
	Category    string
	Comments    string

	// Enclosure
	URL  string
	Len  int64
	Type string

	GUID   string
	Date   time.Time
	Source string
}

// Parse :)
func Parse(body []byte) (i []Item, e error) {
	var estr string
	i, e = parseRSS(body)
	if e != nil {
		estr = e.Error()
		i, e = parseAtom(body)
	}
	if e != nil {
		e = errors.New(estr + " | " + e.Error())
	}

	for k := range i {
		i[k].Title = html.UnescapeString(i[k].Title)
		i[k].Link = html.UnescapeString(i[k].Link)
		i[k].Description = html.UnescapeString(i[k].Description)
		i[k].Author = html.UnescapeString(i[k].Author)
		i[k].Category = html.UnescapeString(i[k].Category)
		i[k].Comments = html.UnescapeString(i[k].Comments)
		i[k].URL = html.UnescapeString(i[k].URL)
		i[k].Type = html.UnescapeString(i[k].Type)
		i[k].GUID = html.UnescapeString(i[k].GUID)
		i[k].Source = html.UnescapeString(i[k].Source)
	}

	return
}
