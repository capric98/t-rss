package client

import (
	"fmt"
	"strings"
	"unicode"
)

type ClientType struct {
	Name   string
	Client interface{}
}

func speedParse(s string) string {
	if s == "" {
		return ""
	}
	var spNum int64
	u := make([]rune, 0)
	for _, c := range s {
		if !unicode.IsDigit(c) {
			u = append(u, c)
		} else {
			spNum = spNum*10 + int64(c-'0')
		}
	}
	unit := strings.ToLower(string(u))
	switch {
	case unit == "K" || unit == "k" || unit == "KB" || unit == "kB" || unit == "KiB" || unit == "kiB":
		spNum = spNum * 1024
	case unit == "M" || unit == "m" || unit == "MB" || unit == "mB" || unit == "MiB" || unit == "miB":
		spNum = spNum * 1024 * 1024
	case unit == "G" || unit == "g" || unit == "GB" || unit == "gB" || unit == "GiB" || unit == "giB":
		spNum = spNum * 1024 * 1024 * 1024
	}
	return fmt.Sprintf("%d", spNum)
}
