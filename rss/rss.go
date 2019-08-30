package rss

import (
	"html"
	"mime"
	"net/http"
	"net/url"
	"path"
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
			GUID:        NameRegularize(v.GUID),
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
		if v.GUID == "" {
			v.GUID = NameRegularize(v.Title)
			if len(v.GUID) > 200 {
				v.GUID = v.GUID[:200]
			}
		} else {
			v.GUID = NameRegularize(v.GUID)
		} // Just in case.
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
	var urlname, headername = path.Base(furl), ""

	if urlname == "" {
		urlname = "download.torrent" // In case of a blank name.
	}

	if headermap["Content-Disposition"] != nil {
		_, params, _ := mime.ParseMediaType(headermap["Content-Disposition"][0])
		headername = params["filename"]
	}
	if headername == "" {
		headername = urlname // In case of a blank name.
	}

	rh, _ := url.QueryUnescape(headername)
	return rh
}
