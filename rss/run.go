package rss

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/capric98/t-rss/bencode"
)

func checkRegexp(v RssRespType, reg []*regexp.Regexp) bool {
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
	result, err := bencode.Decode(data)
	if err != nil {
		return false, -1
	}
	info := result[0].Get("info")
	pl := (info.Get("piece length")).Value
	ps := int64(len((info.Get("pieces")).ByteStr)) / 20
	length := pl * ps
	if min < length && length < max {
		return true, length
	}
	return false, length
}

func (r RssRespType) getTorrent(t Config, c *http.Client) ([]byte, string, error) {
	req, err := http.NewRequest("GET", r.DURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("%v\n", err)
	}
	if t.Cookie != "" {
		req.Header.Add("Cookie", t.Cookie)
	}

	resp, err := c.Do(req)
	for try := 0; try < 3; {
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if err == nil {
			resp.Body.Close()
		} // StatusCode != 200
		time.Sleep(200 * time.Millisecond)
		resp, err = c.Do(req)
		try++
	}
	if err != nil {
		return nil, "", fmt.Errorf("Failed to download torrent file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("Item \"%s\" met status code %d.", r.Title, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to dump byte data: %v", err)
	}
	return body, GetFileInfo(r.DURL, resp.Header), nil
}

func savehistory(path string) {
	if f, err := os.Create(path); err != nil {
		LevelPrintLog(fmt.Sprintf("Warning: %v", err), true)
	} else {
		f.Close()
	}
}

func (r RssRespType) save(task Config, fetcher *http.Client) {
	startT := time.Now()

	data, filename, err := r.getTorrent(task, fetcher)
	path := task.Download_to
	if err != nil {
		LevelPrintLog(fmt.Sprintf("%v", err), true)
		return
	}
	if r.Length == 0 {
		if pass, tlen := checkTLength(data, task.Min, task.Max); !pass {
			LevelPrintLog(fmt.Sprintf("%s: Reject item \"%s\" due to TORRENT content_size not fit.\n", task.TaskName, r.Title), true)
			LevelPrintLog(fmt.Sprintf("%d vs [%d,%d]\n", tlen, task.Min, task.Max), false)
		}
	}

	if TestOnly {
		PrintTimeInfo(fmt.Sprintf("Item \"%s\" done.", r.Title), time.Since(startT))
		return
	}

	if path != "" {
		err := ioutil.WriteFile(path+string(os.PathSeparator)+filename, data, 0644)
		if err != nil {
			LevelPrintLog(fmt.Sprintf("Warning: %v\n", err), true)
			return
		}
		LevelPrintLog(fmt.Sprintf("Item is saved to \"%s\"\n", filename), false)
	}

	// Add file to client.
	if task.Client != nil && !Learn {
		for _, v := range task.Client {
			e := v.Add(data, filename)
			if e != nil {
				LevelPrintLog(fmt.Sprintf("Failed to add item \"%s\" to %s's %s client with message: \"%v\".\n", r.Title, v.Label(), v.Name(), e), true)
				return
			}
		}
	}
	PrintTimeInfo(fmt.Sprintf("Item \"%s\" done.", r.Title), time.Since(startT))
	savehistory(CDir + r.GUID)
}

func runTask(t Config, signal chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	fetcher := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}

	for range signal {
		LevelPrintLog(fmt.Sprintf("Run task: %s.\n", t.TaskName), true)
		startT := time.Now()

		rssresp, err := fetch(t.RSSLink, &fetcher, t.Cookie)
		if err != nil {
			LevelPrintLog(fmt.Sprintf("Caution: Task %s failed to get RSS data and raised an error: %v.\n", t.TaskName, err), true)
			time.Sleep(t.Interval)
			continue
		}

		acCount := 0
		rjCount := 0
		for _, v := range rssresp {
			// Check if item had been accepted yet.
			if _, err := os.Stat(CDir + v.GUID); !os.IsNotExist(err) {
				rjCount++
				continue
			}

			// Check regexp filter.
			if (t.Reject != nil) && (checkRegexp(v, t.Reject)) {
				LevelPrintLog(fmt.Sprintf("%s: Reject item \"%s\"\n", t.TaskName, v.Title), true)
				rjCount++
				continue
			}
			if t.Accept != nil && (t.Strict) && (!checkRegexp(v, t.Accept)) {
				LevelPrintLog(fmt.Sprintf("%s: Cannot accept item \"%s\" due to strict mode.\n", t.TaskName, v.Title), true)
				rjCount++
				continue
			}

			// Check content_size.
			if !(v.Length >= t.Min && v.Length <= t.Max) {
				LevelPrintLog(fmt.Sprintf("%s: Reject item \"%s\" due to content_size not fit.\n", t.TaskName, v.Title), true)
				LevelPrintLog(fmt.Sprintf("%d vs [%d,%d]\n", v.Length, t.Min, t.Max), false)
				rjCount++
				continue
			}

			LevelPrintLog(fmt.Sprintf("%s: Accept item \"%s\"\n", t.TaskName, v.Title), true)
			acCount++

			if Learn {
				savehistory(CDir + v.GUID)
			} else {
				go v.save(t, &fetcher)
			}
		}
		PrintTimeInfo(fmt.Sprintf("Task %s: Accept %d item(s), reject %d item(s).", t.TaskName, acCount, rjCount), time.Since(startT))
		if Learn {
			return
		}
		time.Sleep(t.Interval)
	}
}

func tick(signal chan struct{}) {
	if Learn {
		signal <- struct{}{}
		runtime.Gosched()
		close(signal)
		return
	}
	for {
		signal <- struct{}{}
	}
}
