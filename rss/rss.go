package rss

import (
	"html"
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
	Length      int64 //this is total length...
	Date        string
	GUID        string
}

func fetch(rurl string, client *http.Client, cookie string) ([]RssRespType, error) {
	req, err := http.NewRequest("GET", rurl, nil)
	if err != nil {
		return nil, err
	}
	if cookie != "" {
		req.Header.Add("Cookie", cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fp := gofeed.NewParser()
	rssFeed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	Rresp := make([]RssRespType, len(rssFeed.Items))
	for i, v := range rssFeed.Items {
		Rresp[i] = RssRespType{
			Title:       v.Title,
			Description: html.UnescapeString(v.Description),
			Date:        v.Published,
			GUID:        v.GUID,
		}
		if v.Enclosures != nil {
			tmp, err := strconv.Atoi(v.Enclosures[0].Length)
			if err != nil {
				tmp = 0
			}
			Rresp[i].DURL = v.Enclosures[0].URL
			Rresp[i].Length = int64(tmp)
		} else {
			Rresp[i].DURL = v.Link
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
		headername = headername[strings.Index(headername, "filename=")+9:]
		for headername[0] == ' ' || headername[0] == '"' {
			headername = headername[1:]
		}
		if strings.Contains(headername, ";") {
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
