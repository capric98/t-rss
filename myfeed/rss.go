package myfeed

import (
	"encoding/xml"
	"html"
	"io"
)

type RSSFeed struct {
	Version string    `xml:"version,attr"`
	Channel []Channel `xml:"channel"`
}

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

	Items []rItem `xml:"item"`
}

type rItem struct {
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
		Url  string `xml:"url,attr"`
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

func rParse(r io.ReadCloser) (f []Item, e error) {
	var feed RSSFeed
	e = xml.NewDecoder(r).Decode(&feed)
	if e == nil && len(feed.Channel) == 0 {
		e = ErrNotRSSFormat
	}

	if e == nil {
		items = items[:0]
		for _, c := range feed.Channel {
			for _, v := range c.Items {
				i := Item{
					rItem:   v,
					PubDate: strToTime(v.SpubDate),
				}
				i.Link = html.UnescapeString(i.Link)
				i.Enclosure.Url = html.UnescapeString(i.Link)
				i.Description = html.UnescapeString(i.Description)
				i.Comments = html.UnescapeString(i.Comments)
				items = append(items, i)
			}
		}
		f = items
	}
	return
}
