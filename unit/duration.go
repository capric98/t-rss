package unit

import (
	"regexp"
	"strconv"
	"time"
)

var (
	secondReg = regexp.MustCompile(`[0-9]+s`)
	minuteReg = regexp.MustCompile(`[0-9]+m`)
	hourReg   = regexp.MustCompile(`[0-9]+h`)
	dayReg    = regexp.MustCompile(`[0-9]+d`)
	residue   = regexp.MustCompile(`[0-9]+`)
)

// ParseDuration parses a string to time.Duration.
// If fails to parse a string, it will return 0.
func ParseDuration(s string) time.Duration {
	second, _ := strconv.ParseInt(shave(secondReg.FindString(s), 1), 10, 64)
	minute, _ := strconv.ParseInt(shave(minuteReg.FindString(s), 1), 10, 64)
	hour, _ := strconv.ParseInt(shave(hourReg.FindString(s), 1), 10, 64)
	day, _ := strconv.ParseInt(shave(dayReg.FindString(s), 1), 10, 64)
	t := time.Duration(day)*24*time.Hour +
		time.Duration(hour)*time.Hour +
		time.Duration(minute)*time.Minute +
		time.Duration(second)*time.Second
	if t == 0 {
		second, _ = strconv.ParseInt(residue.FindString(s), 10, 64)
		t = time.Duration(second) * time.Second
	}
	return t
}
