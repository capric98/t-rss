package rss

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

func PrintTimeInfo(info string, t time.Duration) {
	oLock.Lock()
	fmt.Fprintf(os.Stderr, time.Now().Format("2006/01/02 15:04:05 "))
	fmt.Fprintf(os.Stderr, info)
	color.Set(color.FgWhite, color.BgGreen)
	fmt.Fprintf(os.Stderr, "%s", t)
	color.Unset()
	fmt.Fprintf(os.Stderr, ".\n")
	oLock.Unlock()
}

func LevelPrintLog(s string, important bool) {
	if important || (DMode) {
		log.Println(s)
	}
}
