package core

import (
	"github.com/capric98/t-rss/torrents"
	"github.com/capric98/t-rss/torrents/bencode"
	"sync"
	"regexp"
	"fmt"
	"os"
)

func (w *worker)run(wg *sync.WaitGroup) {
	defer wg.Done()

	Min:=int64(0)
	Max:=int64(0x7FFFFFFF)

	for tasks := range w.ticker {
		w.log(fmt.Sprintf("Run task: %s.\n", w.name), 1)
				//startT := time.Now()
		
				acCount := 0
				rjCount := 0
				for _, v := range tasks {
					// Check if item had been accepted yet.
					if _, err := os.Stat(CDir + v.GUID); !os.IsNotExist(err) {
						rjCount++
						continue
					}
		
					// Check regexp filter.
					if (w.Config.Reject != nil) && (checkRegexp(v, w.Config.Reject)) {
						w.log(fmt.Sprintf("%s: Reject item \"%s\"\n", w.name, v.Title), 1)
						rjCount++
						continue
					}
					if w.Config.Accept != nil && (w.Config.Strict) && (!checkRegexp(v, w.Config.Accept)) {
						w.log(fmt.Sprintf("%s: Cannot accept item \"%s\" due to strict mode.\n", w.name, v.Title), 1)
						rjCount++
						continue
					}
		
					// Check content_size.
					if !(v.Length >= Min && v.Length <= Max) {
						w.log(fmt.Sprintf("%s: Reject item \"%s\" due to content_size not fit.\n", w.name, v.Title), 1)
						w.log(fmt.Sprintf("%d vs [%d,%d]\n", v.Length, Min, Max), 0)
						rjCount++
						continue
					}
		
					w.log(fmt.Sprintf("%s: Accept item \"%s\"\n", w.name, v.Title), 1)
					acCount++
		
				// 	if Learn {
				// 		savehistory(CDir + v.GUID)
				// 	} else {
				// 		go v.save(t, &fetcher)
				// 	}
				}
				// PrintTimeInfo(fmt.Sprintf("Task %s: Accept %d item(s), reject %d item(s).", t.TaskName, acCount, rjCount), time.Since(startT))
				// if Learn {
				// 	return
				// }
	}
}

func checkRegexp(v torrents.Individ, reg []*regexp.Regexp) bool {
	for _, r := range reg {
		if r.MatchString(v.Title) {
			return true
		}
		// matched, _ = regexp.MatchString(r, v.Description)
		// if matched {
		// 	return true
		// }
		if r.MatchString(v.Author) {
			return true
		}
	}
	return false
}

func checkTLength(data []byte, min int64, max int64) (bool, int64) {
	defer func() {
		_ = recover()
	}()

	result, err := bencode.Decode(data)
	if err != nil {
		return false, -1
	}
	info := result[0].Dict("info")
	pl := (info.Dict("piece length")).Value()
	ps := int64(len((info.Dict("pieces")).BStr())) / 20
	length := pl * ps
	if min < length && length < max {
		return true, length
	}
	return false, length
}