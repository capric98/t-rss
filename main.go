package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/capric98/t-rss/core"
)

var (
	DMode      = flag.Bool("debug", false, "enable debug mode with a more detailed log output.") // Debug log output.
	TestOnly   = flag.Bool("test", false, "dry run a test without caching the history/save to file/add to client")
	Learn      = flag.Bool("learn", false, "run rss without adding to client")
	ConfigPath = flag.String("config", "config.yml", "config file path")
	CDir       = flag.String("history", ".RSS-saved", "directory to save rss history")
)

func main() {
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	core.Init(DMode, TestOnly, Learn, ConfigPath, CDir)
}
