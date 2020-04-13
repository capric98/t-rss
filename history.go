package trss

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func checkAndWatchHistory(path string, maxAge time.Duration, log *logrus.Logger) {
	if _, e := os.Stat(path); os.IsNotExist(e) {
		log.WithFields(logrus.Fields{
			"@func": "checkAndWatchHistory",
			"path":  path,
		}).Info("path does not exsit, create it")
		e = os.MkdirAll(path, 0640)
		if e != nil {
			log.Fatal(e)
		}
	}
	go watchHistroy(path, maxAge, log.WithField("@func", "watchHistory"))
}

func watchHistroy(path string, maxAge time.Duration, log *logrus.Entry) {
	log.Debug("start to watch history dir")
	for {
		fl, e := walkDir(path)
		if e != nil {
			log.Warn("walkDir:", e)
			continue
		}
		for _, v := range fl {
			if info, err := os.Stat(v); err == nil {
				mod := time.Since(info.ModTime())
				log.WithFields(logrus.Fields{
					"filename": info.Name(),
				}).Trace(
					"moded ", mod, " ago",
				)

				if mod > maxAge {
					log.WithField("filename", info.Name()).Debug("delete old history file")
					err = os.Remove(v)
					if err != nil {
						log.Warn("delete: ", err)
					}
				}
			}
		}
		time.Sleep(12 * time.Hour)
	}
}

func walkDir(hd string) (fl []string, e error) {
	e = filepath.Walk(hd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fl = append(fl, path)
		}
		return nil
	})
	return
}
