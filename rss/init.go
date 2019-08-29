package rss

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	oLock      = sync.Mutex{}
	DMode      bool
	ConfigPath string
	TestOnly   bool
	CDir       string
	Learn      bool
)

func Init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if DMode {
		LevelPrintLog(fmt.Sprintf("Debug: %t\n", DMode), true)
		LevelPrintLog(fmt.Sprintf("Test: %t\n", TestOnly), true)
		LevelPrintLog(fmt.Sprintf("Config: %s\n", ConfigPath), true)
		LevelPrintLog(fmt.Sprintf("History: %s\n", CDir), true)
	}

	cdata, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		LevelPrintLog(fmt.Sprintf("Error: %v\n", err), true)
		os.Exit(2)
	}

	if ConfigPath != "config.yml" && CDir == ".RSS-saved" {
		// Change config file path without setting CDir.
		CDir = filepath.Dir(ConfigPath) + string(os.PathSeparator) + ".RSS-saved" + string(os.PathSeparator)
	}
	CDir = filepath.Dir(CDir+string(os.PathSeparator)) + string(os.PathSeparator) // Just in case.
	LevelPrintLog(fmt.Sprintf("History will be saved to: %s\n", CDir), false)
	if _, err := os.Stat(CDir); os.IsNotExist(err) {
		merr := os.Mkdir(CDir, 0644)
		if merr != nil {
			LevelPrintLog(fmt.Sprintf("%v\n", merr), true)
			os.Exit(2)
		} else {
			LevelPrintLog(fmt.Sprintf("%s did not exist, make it!", CDir), false)
		}

	}
	taskList := parseSettings(cdata)
	if TestOnly {
		log.Println(ConfigPath, "passes the test.")
		//return
	}

	qsignal := make(chan error, 2)
	var wg sync.WaitGroup
	done := make(chan bool)
	go func() {
		c := make(chan os.Signal, 10) // bufferd
		signal.Notify(c, os.Interrupt)
		qsignal <- fmt.Errorf("%s", <-c)
	}()

	for _, t := range taskList {
		wg.Add(1)
		go runTask(t, &wg)
	}
	go cleanDaemon()
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-qsignal:
		LevelPrintLog(fmt.Sprintf("Receive signal 2, quit the program.\n"), true)
	case <-done:
		LevelPrintLog(fmt.Sprintf("All task goroutines quit.\n"), true)
	}
	LevelPrintLog(fmt.Sprintf("Bye.\n"), true)
}

func NInit() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if DMode {
		LevelPrintLog(fmt.Sprintf("Debug: %t\n", DMode), true)
		LevelPrintLog(fmt.Sprintf("Test: %t\n", TestOnly), true)
		LevelPrintLog(fmt.Sprintf("Config: %s\n", ConfigPath), true)
		LevelPrintLog(fmt.Sprintf("History: %s\n", CDir), true)
	}

	// Check config file path and determin to use default history save path or not.
	// ====================================================================================================
	if ConfigPath != "config.yml" && CDir == ".RSS-saved" {
		// Change config file path without setting CDir.
		CDir = filepath.Dir(ConfigPath) + string(os.PathSeparator) + ".RSS-saved" + string(os.PathSeparator)
	}
	CDir = filepath.Dir(CDir+string(os.PathSeparator)) + string(os.PathSeparator) // Just in case.
	LevelPrintLog(fmt.Sprintf("History will be saved to: %s\n", CDir), false)
	if _, err := os.Stat(CDir); os.IsNotExist(err) {
		merr := os.Mkdir(CDir, 0644)
		if merr != nil {
			LevelPrintLog(fmt.Sprintf("%v\n", merr), true)
			os.Exit(2)
		} else {
			LevelPrintLog(fmt.Sprintf("%s did not exist, make it!", CDir), false)
		}
	}
	// ====================================================================================================

	// Read the config file and parse to task list.
	cdata, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		LevelPrintLog(fmt.Sprintf("Error: %v\n", err), true)
		os.Exit(2)
	}
	taskList := parse(cdata)
	if TestOnly {
		log.Println(ConfigPath, "passes the test.")
		//return
	}

	// Start all the task.
	qsignal := make(chan error, 2)
	var wg sync.WaitGroup
	done := make(chan bool)
	go func() {
		c := make(chan os.Signal, 10) // bufferd
		signal.Notify(c, os.Interrupt)
		qsignal <- fmt.Errorf("%s", <-c)
	}()

	for _, t := range taskList {
		wg.Add(1)
		fmt.Println(t)
		//go runTask(t, &wg)
	}
	go cleanDaemon()
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-qsignal:
		LevelPrintLog(fmt.Sprintf("Receive signal 2, quit the program.\n"), true)
	case <-done:
		LevelPrintLog(fmt.Sprintf("All task goroutines quit.\n"), true)
	}
	LevelPrintLog(fmt.Sprintf("Bye.\n"), true)
}
