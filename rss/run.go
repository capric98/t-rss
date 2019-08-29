package rss

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/capric98/t-rss/client"
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

func saveItem(r RssRespType, t TaskType, Client *http.Client, wg *sync.WaitGroup) {
	req, err := http.NewRequest("GET", r.DURL, nil)
	if err != nil {
		LevelPrintLog(fmt.Sprintf("%v\n", err), true)
		return
	}
	if t.Cookie != "" {
		req.Header.Add("Cookie", t.Cookie)
	}

	startT := time.Now()

	resp, err := Client.Do(req)
	for try := 0; try < 3; {
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if err == nil {
			resp.Body.Close()
		} // StatusCode != 200
		time.Sleep(200 * time.Millisecond)
		resp, err = Client.Do(req)
		try++
	}
	if err != nil {
		LevelPrintLog(fmt.Sprintf("%v\n", err), true)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		LevelPrintLog(fmt.Sprintf("Item \"%s\" met status code %d.", r.Title, resp.StatusCode), true)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LevelPrintLog(fmt.Sprintf("%v\n", err), true)
	}

	if TestOnly {
		PrintTimeInfo(fmt.Sprintf("Item \"%s\" done.", r.Title), time.Since(startT))
		return
	}

	filename := GetFileInfo(r.DURL, resp.Header)

	if t.DownPath != "" {
		err := ioutil.WriteFile(t.DownPath+string(os.PathSeparator)+filename, body, 0644)
		if err != nil {
			LevelPrintLog(fmt.Sprintf("Warning: %v\n", err), true)
			return
		}
		LevelPrintLog(fmt.Sprintf("Item \"%s\" is saved as \"%s\"\n", r.Title, filename), false)
	}

	// Add file to client.
	if t.Client != nil && !Learn {
		for _, v := range t.Client {
			for ec := 0; ec < 3; ec++ {
				switch v.Name {
				case "qBittorrent":
					if ec != 0 {
						_ = v.Client.(*client.QBType).Init()
					} // In case of the session timeout, reinitiallize it.
					err = v.Client.(*client.QBType).Add(body, filename)
				case "Deluge":
					err = v.Client.(*client.DeType).Add(body, filename)
				}
				if err != nil {
					// If fail to add torrent to client, try another 2 times.
					LevelPrintLog(fmt.Sprintf("%s: Failed to add item \"%s\" to %s client with message: \"%v\".\n", t.TaskName, r.Title, v.Name, err), true)
				} else {
					break
				}
				if ec == 2 {
					return
					// Failed 3 times, quit and do not save history.
				}

			}
		}
	}
	PrintTimeInfo(fmt.Sprintf("Item \"%s\" done.", r.Title), time.Since(startT))

	if f, err := os.Create(CDir + r.GUID); err != nil {
		LevelPrintLog(fmt.Sprintf("Warning: %v", err), true)
	} else {
		f.Close()
	}
	if Learn {
		wg.Done()
	}
	// Under test only mode, we do not create history file.
}

func runTask(t TaskType, wg *sync.WaitGroup) {
	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	for {
		LevelPrintLog(fmt.Sprintf("Run task: %s.\n", t.TaskName), true)
		startT := time.Now()

		Rresp, err := fetch(t.RSSLink, &client, t.Cookie)
		if err != nil {
			LevelPrintLog(fmt.Sprintf("Caution: Task %s failed to get RSS data and raised an error: %v.\n", t.TaskName, err), true)
			time.Sleep(time.Duration(t.Interval) * time.Second)
			continue
		}
		acCount := 0
		rjCount := 0
		for _, v := range Rresp {
			// Check if item had been accepted yet.
			if v.GUID == "" {
				v.GUID = NameRegularize(v.Title)
				if len(v.GUID) > 200 {
					v.GUID = v.GUID[:200]
				}
			} else {
				v.GUID = NameRegularize(v.GUID)
			} // Just in case.
			if _, err := os.Stat(CDir + v.GUID); !os.IsNotExist(err) {
				rjCount++
				continue
			}

			// Check regexp filter.
			if (t.RjcRegexp != nil) && (checkRegexp(v, t.RjcRegexp)) {
				LevelPrintLog(fmt.Sprintf("%s: Reject item \"%s\"\n", t.TaskName, v.Title), true)
				rjCount++
				continue
			}
			if t.AccRegexp != nil && (!checkRegexp(v, t.AccRegexp)) && (t.Strict) {
				LevelPrintLog(fmt.Sprintf("%s: Cannot accept item \"%s\" due to strict mode.\n", t.TaskName, v.Title), true)
				rjCount++
				continue
			}

			// Check content_size.
			if !((v.Length > t.MinSize && v.Length < t.MaxSize) || (v.Length == 0 && !t.Strict)) {
				LevelPrintLog(fmt.Sprintf("%s: Reject item \"%s\" due to content_size not fit.\n", t.TaskName, v.Title), true)
				LevelPrintLog(fmt.Sprintf("%d vs [%d,%d]\n", v.Length, t.MinSize, t.MaxSize), false)
				rjCount++
				continue
			}

			LevelPrintLog(fmt.Sprintf("%s: Accept item \"%s\"\n", t.TaskName, v.Title), true)
			acCount++

			if Learn {
				wg.Add(1)
			}
			go saveItem(v, t, &client, wg)
		}
		PrintTimeInfo(fmt.Sprintf("Task %s: Accept %d item(s), reject %d item(s).", t.TaskName, acCount, rjCount), time.Since(startT))
		if Learn {
			LevelPrintLog("Learning finished.", true)
			wg.Done()
			return
		}
		time.Sleep(time.Duration(t.Interval) * time.Second)
	}
}
