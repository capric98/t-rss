package client

import (
	"strconv"
	"strings"
	"unicode"
)

// Client :)
type Client interface {
	Add(b []byte, name string) error
	Name() string
}

// UConvert: convert a string which may contain unit to a float64 with bytes unit.
func UConvert(s string) float64 {
	if s == "" {
		return 0
	}

	spNum := float64(0)
	number := make([]rune, 0)
	runit := make([]rune, 0)

	for _, r := range s {
		if unicode.IsDigit(r) || r == '.' || r == '-' {
			number = append(number, r)
		} else {
			runit = append(runit, r)
		}
	}

	sunit := strings.TrimSpace(string(runit))
	spNum, _ = strconv.ParseFloat(strings.TrimSpace(string(number)), 64)

	switch {
	case sunit == "K" || sunit == "k" || sunit == "KB" || sunit == "kB" || sunit == "KiB" || sunit == "kiB":
		spNum = spNum * 1024
	case sunit == "M" || sunit == "m" || sunit == "MB" || sunit == "mB" || sunit == "MiB" || sunit == "miB":
		spNum = spNum * 1024 * 1024
	case sunit == "G" || sunit == "g" || sunit == "GB" || sunit == "gB" || sunit == "GiB" || sunit == "giB":
		spNum = spNum * 1024 * 1024 * 1024
	case sunit == "T" || sunit == "t" || sunit == "TB" || sunit == "tB" || sunit == "TiB" || sunit == "tiB":
		spNum = spNum * 1024 * 1024 * 1024 * 1024
	default:
		spNum = spNum * 1024 * 1024
	}

	return spNum
}
