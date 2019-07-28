package RSS

import (
	"html"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mmcdole/gofeed"
)

type RssRespType struct {
	Title       string
	Description string
	Author      string
	Categories  []string
	DURL        string
	Length      int //this is total length...
	Date        string
	GUID        string
}

func RssFetch(rurl string, client *http.Client) ([]RssRespType, error) {
	resp, err := client.Get(rurl)
	if err != nil {
		log.Printf("Caution: Failed to get rss meta: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	fp := gofeed.NewParser()
	rssFeed, _ := fp.Parse(resp.Body)
	Rresp := make([]RssRespType, len(rssFeed.Items))
	for i, v := range rssFeed.Items {
		tmp, err := strconv.Atoi(v.Enclosures[0].Length)
		if err != nil {
			tmp = 0
		}
		Rresp[i] = RssRespType{
			Title:       v.Title,
			Description: html.UnescapeString(v.Description),
			//Author:      v.Author.Name,
			//Categories:  v.Categories,
			DURL:   v.Enclosures[0].URL,
			Length: tmp,
			Date:   v.Published,
			GUID:   v.GUID,
		}
		if v.Author != nil {
			Rresp[i].Author = v.Author.Name
		}
	}
	return Rresp, nil
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
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}

func GetFileInfo(furl string, headermap http.Header) string {
	var urlname, headername = "", ""
	for p := len(furl) - 1; furl[p] != '/'; p-- {
		urlname = string(furl[p]) + urlname
	}
	if urlname == "" {
		urlname = "download" // In case of a blank name.
	}
	if headermap["Content-Disposition"] != nil {
		headername = headermap["Content-Disposition"][0]
		headername = headername[strings.Index(headername, "filename=")+9 : len(headername)]
		for headername[0] == ' ' || headername[0] == '"' {
			headername = headername[1:]
		}
		if strings.Index(headername, ";") != -1 {
			headername = headername[:strings.Index(headername, ";")]
		}
		for headername[len(headername)-1] == ' ' || headername[len(headername)-1] == '"' {
			headername = headername[:len(headername)-1]
		}
	}
	if headername == "" {
		headername = urlname // In case of a blank name.
	}

	rh, _ := url.QueryUnescape(headername)
	return rh
}
