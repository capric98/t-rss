package core

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func cleanDaemon(CDir string) {
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
				log.Println("Error during autoclean:", err)
				continue
			}
			if info.IsDir() {
				continue
			} else if time.Since(info.ModTime()) > (10 * 259200 * time.Second) {
				err := os.Remove(file) //Remove history file which exists over 30 days.
				if err != nil {
					log.Println("Failed to remove file", file, "with error:", err)
					continue
				}
				log.Println("Delete history file", file)
			}
		}
		time.Sleep(86400 * time.Second) // Perform clean once every day.
	}
}
