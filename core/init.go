package core

import (
	"github.com/capric98/t-rss/rss"
	"log"
	"path/filepath"
	"os"
	"sync"
	"os/signal"
	"io/ioutil"
)

func Init(DMode *bool, TestOnly *bool, Learn *bool, ConfigPath *string, CDir *string) {
	if *DMode {
		log.Println("Debug:", *DMode)
		log.Println("Test:", *TestOnly)
		log.Println("Config:", *ConfigPath)
		log.Println("History:", *CDir)
	}

	// Check config file path and determin to use default history save path or not.
	// ====================================================================================================
	if *ConfigPath != "config.yml" && *CDir == ".RSS-saved" {
		// Change config file path without setting CDir.
		*CDir = filepath.Dir(*ConfigPath) + string(os.PathSeparator) + ".RSS-saved" + string(os.PathSeparator)
	}
	*CDir = filepath.Dir(*CDir+string(os.PathSeparator)) + string(os.PathSeparator) // Just in case.
	log.Println("History will be saved to:", *CDir)
	if _, err := os.Stat(*CDir); os.IsNotExist(err) {
		merr := os.Mkdir(*CDir, 0644)
		if merr != nil {
			log.Fatal(merr)
		} else {
			log.Println(*CDir,"did not exist, make it!")
		}
	}
	// ====================================================================================================

	//Read the config file and parse to task list.
	cdata, err := ioutil.ReadFile(*ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	tasks := parse(cdata)
	if *TestOnly {
		log.Println(ConfigPath, "passes the test.")
		return
	}

	// Start all the task(s).
	wg := sync.WaitGroup{}
	for k, v := range tasks {
		wg.Add(1)
		
		nw := &worker{
			name:k,
			loglevel:1,
			Config: v,
		}
		if *DMode {
			nw.loglevel = 0
		}
		if nw.Config.RSSLink!="" {
			nw.ticker = rss.NewTicker(v.RSSLink, v.Interval, v.Cookie)
		}

		go nw.run(&wg)
	}
	go cleanDaemon(*CDir)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		log.Println("Receive signal 2, quit the program.")
	case <-done:
		log.Println("All task goroutines quit.")
	}
	log.Println("Bye.")
}
