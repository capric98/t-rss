package unit

import "time"

var (
	timeFormat = []string{time.ANSIC, time.UnixDate, time.RubyDate,
		time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano,
		time.RFC822, time.RFC822Z, time.RFC850, time.Kitchen,
		time.Stamp, time.StampMicro, time.StampMilli, time.StampNano}
)

// ParseTime parses a string to time.Time.
// If fails to parse a string, it will return time.Now().
func ParseTime(s string) time.Time {
	for k := range timeFormat {
		t, e := time.Parse(timeFormat[k], s)
		if e == nil {
			return t
		}
	}
	return time.Now()
}
