package trss

import (
	"fmt"
	"runtime"
)

var (
	version = "v0.6.12"
	intro   = fmt.Sprintf("t-rss %v %v/%v (%v build)\n", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
)

func init() {
	fmt.Println(intro)
}
