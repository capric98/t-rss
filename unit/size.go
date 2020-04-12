package unit

import (
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
)

// ParseSize parses a string to int64 size.
// If fails to parse a string, it will return 0.
func ParseSize(s string) int64 {
	b, _ := strconv.ParseInt(toD.FindString(bReg.FindString(s)), 10, 64)
	kiB, _ := strconv.ParseInt(toD.FindString(kiBReg.FindString(s)), 10, 64)
	miB, _ := strconv.ParseInt(toD.FindString(miBReg.FindString(s)), 10, 64)
	giB, _ := strconv.ParseInt(toD.FindString(giBReg.FindString(s)), 10, 64)
	tiB, _ := strconv.ParseInt(toD.FindString(tiBReg.FindString(s)), 10, 64)
	return tiB<<40 + giB<<30 + miB<<20 + kiB<<10 + b
}
