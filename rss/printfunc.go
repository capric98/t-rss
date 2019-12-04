package rss

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"

// 	"github.com/fatih/color"
// )

// func PrintTimeInfo(info string, t time.Duration) {
// 	oLock.Lock()
// 	fmt.Fprint(os.Stderr, time.Now().Format("2006/01/02 15:04:05 "))
// 	color.Set(color.FgGreen)
// 	fmt.Fprintf(os.Stderr, "%7.2fms", t.Seconds()*1000.0)
// 	color.Unset()
// 	fmt.Fprint(os.Stderr, " ")
// 	fmt.Fprint(os.Stderr, info)
// 	fmt.Fprint(os.Stderr, "\n")
// 	oLock.Unlock()
// }

// func LevelPrintLog(s string, important bool) {
// 	oLock.Lock()
// 	if important || (DMode) {
// 		log.Print(s)
// 	}
// 	oLock.Unlock()
// }
