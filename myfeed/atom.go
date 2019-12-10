package myfeed

import (
	"encoding/xml"
	"html"
	"io"
)

type AtomFeed struct {
	// Required
	Title       string `xml:"title"`
	Link        string `xml:"href,attr"`
	Description string `xml:"description"`

	// Optional
	PubDate   string `xml:"updated"`
	Generator string `xml:"generator"`

	Items []aItem `xml:"entry"`
}

type aItem struct {
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
		Url  string `xml:"rel,attr"`
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

func aParse(r io.ReadCloser) (f []Item, e error) {
	var feed []AtomFeed
	e = xml.NewDecoder(r).Decode(&feed)
	if e == nil && len(feed) == 0 {
		e = ErrNotAtomFormat
	}

	if e == nil {
		items = items[:0]
		for _, ff := range feed {
			for _, v := range ff.Items {
				i := Item{
					rItem:   toR(v),
					PubDate: strToTime(v.SpubDate),
				}
				i.Link = html.UnescapeString(i.Link)
				i.Enclosure.Url = html.UnescapeString(i.Enclosure.Url)
				i.Enclosure.Type = html.UnescapeString(i.Enclosure.Type)
				if i.Enclosure.Url == "" {
					i.Enclosure.Url = i.Enclosure.Type
				}
				i.Description = html.UnescapeString(i.Description)
				i.Comments = html.UnescapeString(i.Comments)
				items = append(items, i)
			}
		}
		f = items
	}
	return
}

func toR(v aItem) rItem {
	r := rItem{
		Title:       v.Title,
		Link:        v.Link,
		Description: v.Description,
		Author:      v.Author,
		Comments:    v.Comments,
		SpubDate:    v.SpubDate,
		Source:      v.Source,
	}
	r.Category.Domain = v.Category.Domain
	r.Category.Name = v.Category.Name
	r.Enclosure.Url = v.Enclosure.Url
	r.Enclosure.Len = v.Enclosure.Len
	r.Enclosure.Type = v.Enclosure.Type
	r.GUID.IsPermaLink = v.GUID.IsPermaLink
	r.GUID.Value = v.GUID.Value
	return r
}
