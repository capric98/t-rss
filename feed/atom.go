package feed

import (
	"bytes"
	"encoding/xml"

	"github.com/capric98/t-rss/unit"
	"golang.org/x/net/html/charset"
)

// AtomFeed :)
type AtomFeed struct {
	// Required
	Title       string `xml:"title"`
	Link        string `xml:"href,attr"`
	Description string `xml:"description"`

	// Optional
	PubDate   string `xml:"updated"`
	Generator string `xml:"generator"`

	Items []AtomItem `xml:"entry"`
}

// AtomItem :)
type AtomItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"-"`
	Description string `xml:"subtitle"`
	Author      string `xml:"author"`
	Category    struct {
		Domain string `xml:"term,attr"`
		Name   string `xml:"label,attr"`
	} `xml:"category"`
	Comments  string `xml:"-"`
	Enclosure struct {
		URL  string `xml:"rel,attr"`
		Len  int64  `xml:"-"`
		Type string `xml:"href,attr"`
	} `xml:"link"`
	GUID struct {
		IsPermaLink bool   `xml:"-"`
		Value       string `xml:",chardata"`
	} `xml:"id"`
	SpubDate string `xml:"updated"`
	Source   string `xml:"-"`
}

func parseAtom(body []byte) (f []Item, e error) {
	var feed []AtomFeed

	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charset.NewReaderLabel
	e = decoder.Decode(&feed)

	if e != nil {
		return
	}

	for k := range feed {
		for _, v := range feed[k].Items {
			f = append(f, Item{
				Title:       v.Title,
				Link:        v.Link,
				Description: v.Description,
				Author:      v.Author,
				URL:         v.Enclosure.URL,
				Len:         v.Enclosure.Len,
				Type:        v.Enclosure.Type,
				GUID:        v.GUID.Value,
				Date:        unit.ParseTime(v.SpubDate),
				Source:      v.Source,
			})
		}
	}

	return
}
