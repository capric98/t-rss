package trss

import (
	"fmt"
	"runtime"
)

var (
	version = "0.6.0"
	intro   = fmt.Sprintf("t-rss %v %v/%v (%v build)\n", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
)

func init() {
	fmt.Println(intro)
}
