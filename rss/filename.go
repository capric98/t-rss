package rss

import (
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
)

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
