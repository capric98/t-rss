package feed

import (
	"bytes"
	"encoding/xml"

	"github.com/capric98/t-rss/unit"
	"golang.org/x/net/html/charset"
)

// RSSFeed :)
type RSSFeed struct {
	Version string    `xml:"version,attr"`
	Channel []Channel `xml:"channel"`
}

// Channel :)
type Channel struct {
	// Required
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`

	// Optional
	Language  string `xml:"language"`
	Copyright string `xml:"copyright"`
	//managingEditor
	//webMaster
	PubDate string `xml:"pubDate"`
	//lastBuildDate
	//category
	Generator string `xml:"generator"`
	//docs
	//cloud
	//ttl
	//image
	//textInput
	//skipHours
	//skipDays

	Items []RSSItem `xml:"item"`
}

// RSSItem :)
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	Category    struct {
		Domain string `xml:"domain,attr"`
		Name   string `xml:",chardata"`
	} `xml:"category"`
	Comments  string `xml:"comments"`
	Enclosure struct {
		URL  string `xml:"url,attr"`
		Len  int64  `xml:"length,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
	GUID struct {
		IsPermaLink bool   `xml:"type,attr"`
		Value       string `xml:",chardata"`
	} `xml:"guid"`
	SpubDate string `xml:"pubDate"`
	Source   string `xml:"source"`
}

func parseRSS(body []byte) (f []Item, e error) {
	var feed RSSFeed

	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charset.NewReaderLabel
	e = decoder.Decode(&feed)

	if e != nil {
		return
	}

	for _, c := range feed.Channel {
		for _, v := range c.Items {
			i := Item{
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
			}
			f = append(f, i)
		}
	}

	return
}
