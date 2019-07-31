package main

import (
	"flag"

	"github.com/capric98/GoRSS/rss"
)

// import _ "net/http/pprof"

var (
	DMode    = flag.Bool("debug", false, "enable debug mode with a more detailed log output.") // Debug log output.
	Config   = flag.String("config", "config.yml", "config file path")
	TestOnly = flag.Bool("test", false, "dry run a test without caching the history/save to file/add to client")
	CDir     = flag.String("history", ".RSS-saved", "directory to save rss history")
)

func flagInit() {
	rss.DMode = *DMode
	rss.Config = *Config
	rss.TestOnly = *TestOnly
	rss.CDir = *CDir
}

func main() {
	flag.Parse()
	flagInit()

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	rss.Init()
}
