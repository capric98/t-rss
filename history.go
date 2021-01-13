package trss

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type file struct {
	path string
	info os.FileInfo
}

func checkAndWatchHistory(path string, maxNum int, log *logrus.Logger) {
	if _, e := os.Stat(path); os.IsNotExist(e) {
		log.WithFields(logrus.Fields{
			"@func": "checkAndWatchHistory",
			"path":  path,
		}).Info("path does not exsit, create it")
		e = os.MkdirAll(path, 0740)
		if e != nil {
			log.Fatal(e)
		}
	}
	go watchHistroy(path, maxNum, log.WithField("@func", "watchHistory"))
}

func watchHistroy(path string, maxNum int, log *logrus.Entry) {
	log.Debug("start to watch history dir")
	for {
		var subdir []string
		e := filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if i.IsDir() {
				if p != path {
					subdir = append(subdir, p)
					return filepath.SkipDir
				}
			}
			return nil
		})
		if e != nil {
			log.Warn("walk: ", e)
			continue
		}
		for k := range subdir {
			log.Debug("clean subdir: ", subdir[k])
			cleanDir(subdir[k], maxNum, log)
		}
		time.Sleep(12 * time.Hour)
	}
}

func cleanDir(path string, maxNum int, log *logrus.Entry) {
	var f []file
	e := filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !i.IsDir() {
			f = append(f, file{
				info: i,
				path: p,
			})
		} else {
			if p != path {
				return filepath.SkipDir
			}
		}
		return nil
	})
	if e != nil {
		log.Warn("walk: ", e)
		return
	}
	log.Debug("find ", len(f), " files in ", path)
	if len(f) > maxNum {
		sort(f)
		f = f[maxNum:]
		for k := range f {
			log.Debug("delete old history: ", f[k].info.Name())
			if e := os.Remove(f[k].path); e != nil {
				log.Warn("delete old history: ", f[k].info.Name(), " - ", e)
			}
		}
	}
}

func sort(a []file) {
	ll := len(a)
	l, r := 0, ll-1
	if l >= r {
		return
	}
	key := a[l+rand.Intn(r-l)].info.ModTime()
	for l <= r {
		for ; key.Before(a[l].info.ModTime()); l++ {
		}
		for ; a[r].info.ModTime().Before(key); r-- {
		}
		if l <= r {
			tmp := a[l]
			a[l] = a[r]
			a[r] = tmp
			l++
			r--
		}
	}

	if l < ll {
		sort(a[l:])
	}
	if 0 < r {
		sort(a[:r])
	}
}
