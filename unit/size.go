package unit

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	bReg = regexp.MustCompile(`[0-9]+B`)
	// KiBReg - don't use leading k in Go names; var kBReg should be bReg :(
	kiBReg = regexp.MustCompile(`[0-9]+[kK][i]{0,1}B`)
	miBReg = regexp.MustCompile(`[0-9]+[mM][i]{0,1}B`)
	giBReg = regexp.MustCompile(`[0-9]+[gG][i]{0,1}B`)
	tiBReg = regexp.MustCompile(`[0-9]+[tT][i]{0,1}B`)

	toD = regexp.MustCompile(`[0-9]+`)

	u = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "NF"}
)

// ParseSize parses a string to int64 size.
// If fails to parse a string, it will return 0.
func ParseSize(s string) int64 {
	b, _ := strconv.ParseInt(toD.FindString(bReg.FindString(s)), 10, 64)
	kiB, _ := strconv.ParseInt(toD.FindString(kiBReg.FindString(s)), 10, 64)
	miB, _ := strconv.ParseInt(toD.FindString(miBReg.FindString(s)), 10, 64)
	giB, _ := strconv.ParseInt(toD.FindString(giBReg.FindString(s)), 10, 64)
	tiB, _ := strconv.ParseInt(toD.FindString(tiBReg.FindString(s)), 10, 64)
	size := tiB<<40 + giB<<30 + miB<<20 + kiB<<10 + b
	if size == 0 {
		b, _ = strconv.ParseInt(toD.FindString(s), 10, 64)
		size = b
	}
	return size
}

// FormatSize :)
func FormatSize(n int64) string {
	f64n := float64(n)
	count := 0
	for f64n > 1024 {
		count++
		f64n /= 1024.0
	}
	if count >= 6 {
		count = 6
		f64n = 0.1
	}
	return fmt.Sprintf("%.1f%s", f64n, u[count])
}
