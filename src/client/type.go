package client

import (
	"fmt"
	"strconv"
	"unicode"
)

type ClientType struct {
	Name   string
	Client interface{}
}

func SpeedToInt(s string) (int64, error) {
	return strconv.ParseInt(speedParse(s), 10, 64)
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
	unit := string(u)
	switch {
	case unit == "K" || unit == "k" || unit == "KB" || unit == "kB" || unit == "KiB" || unit == "kiB":
		spNum = spNum * 1024
	case unit == "M" || unit == "m" || unit == "MB" || unit == "mB" || unit == "MiB" || unit == "miB":
		spNum = spNum * 1024 * 1024
	case unit == "G" || unit == "g" || unit == "GB" || unit == "gB" || unit == "GiB" || unit == "giB":
		spNum = spNum * 1024 * 1024 * 1024
	case unit == "T" || unit == "t" || unit == "TB" || unit == "tB" || unit == "TiB" || unit == "tiB":
		spNum = spNum * 1024 * 1024 * 1024 * 1024
	}
	return fmt.Sprintf("%d", spNum)
}
