package rss

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func cleanDaemon() {
	LevelPrintLog("Autoclean daemon start!\n", false)

	for {
		var files []string
		err := filepath.Walk(CDir, func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		})
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				LevelPrintLog(fmt.Sprintf("Error during autoclean: %v\n", err), true)
				continue
			}
			if info.IsDir() {
				continue
			} else if time.Since(info.ModTime()) > (10 * 259200 * time.Second) {
				err := os.Remove(file) //Remove history file which exists over 30 days.
				if err != nil {
					LevelPrintLog(fmt.Sprintf("Failed to remove file %s with error: %v\n", file, err), true)
					continue
				}
				LevelPrintLog(fmt.Sprintf("Delete history file %s\n", file), false)
			}
		}
		time.Sleep(86400 * time.Second) // Perform clean once every day.
	}
}
