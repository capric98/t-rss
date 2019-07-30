package main

import (
	"flag"

	"github.com/capric98/GoRSS/RSS"
)

var (
	DMode    = flag.Bool("debug", false, "enable debug mode with a more detailed log output.") // Debug log output.
	Config   = flag.String("config", "config.yml", "config file path")
	TestOnly = flag.Bool("test", false, "dry run a test without caching the history/save to file/add to client")
	CDir     = flag.String("history", ".RSS-saved", "directory to save rss history")
)

func flagInit() {
	RSS.DMode = *DMode
	RSS.Config = *Config
	RSS.TestOnly = *TestOnly
	RSS.CDir = *CDir
}

func main() {
	flag.Parse()
	flagInit()

	RSS.Init()
}
