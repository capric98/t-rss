package core

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/capric98/t-rss/rss"
	"github.com/capric98/t-rss/torrents"
	"github.com/capric98/t-rss/torrents/bencode"
)

func (w *worker) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for tasks := range w.ticker {
		w.log(fmt.Sprintf("Run task: %s.", w.name), 1)

		acCount := 0
		rjCount := 0
		for _, v := range tasks {
			// Check if item had been accepted yet.
			if _, err := os.Stat(CDir + v.GUID); !os.IsNotExist(err) {
				rjCount++
				continue
			} else {
				if !TestOnly {
					w.log(savehistory(CDir+v.GUID), 0)
				}
			}

			// Check regexp filter.
			if (w.Config.Reject != nil) && (checkRegexp(v, w.Config.Reject)) {
				w.log(fmt.Sprintf("%s: Reject item \"%s\"", w.name, v.Title), 1)
				rjCount++
				continue
			}
			if w.Config.Accept != nil && (w.Config.Strict) && (!checkRegexp(v, w.Config.Accept)) {
				w.log(fmt.Sprintf("%s: Cannot accept item \"%s\" due to strict mode.", w.name, v.Title), 1)
				rjCount++
				continue
			}

			// Check content_size.
			if !(v.Length >= w.Config.Min && v.Length <= w.Config.Max) {
				w.log(fmt.Sprintf("%s: Reject item \"%s\" due to content_size not fit.", w.name, v.Title), 1)
				w.log(fmt.Sprintf("%d vs [%d,%d]", v.Length, w.Config.Min, w.Config.Max), 0)
				rjCount++
				continue
			}

			w.log(fmt.Sprintf("%s: Accept item \"%s\"", w.name, v.Title), 1)
			acCount++

			if !Learn {
				go w.save(v)
			}
		}
		w.log(fmt.Sprintf("Task %s: Accept %d item(s), reject %d item(s).", w.name, acCount, rjCount), 1)
		if Learn {
			return
		}
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

func (w *worker) save(t torrents.Individ) {
	defer func() {
		if e := recover(); e != nil {
			w.log(e, 1)
		}
	}()

	time.Sleep(w.Config.Latency)
	startT := time.Now()

	data, filename, err := w.getTorrent(t)
	path := w.Config.Download_to
	if err != nil {
		w.log(fmt.Sprintf("Save: %v", err), 1)
		return
	}
	if t.Length == 0 && w.Config.Strict {
		if pass, tlen := checkTLength(data, w.Config.Min, w.Config.Max); !pass {
			w.log(fmt.Sprintf("%s: Reject item \"%s\" due to TORRENT content_size not fit.", w.name, t.Title), 1)
			w.log(fmt.Sprintf("%d vs [%d,%d]", tlen, w.Config.Min, w.Config.Max), 0)
			return
		}
	}

	if TestOnly {
		w.log(fmt.Sprintf("Item \"%s\" done.", t.Title), 1)
		return
	}

	if w.Config.DeleteT != nil || w.Config.AddT != nil {
		nd, err := w.editTorrent(data)
		if err != nil {
			w.log(fmt.Sprintf("Edit torrent: %v", err), 1)
		} else {
			data = nd
		}
	}

	if path != "" {
		err := ioutil.WriteFile(path+string(os.PathSeparator)+filename, data, 0644)
		if err != nil {
			w.log(fmt.Sprintf("Warning: %v", err), 1)
			return
		}
		w.log(fmt.Sprintf("Item is saved to \"%s\"", filename), 0)
	}

	//Add file to client.
	if w.Config.Client != nil && !Learn {
		for _, v := range w.Config.Client {
			e := v.Add(data, filename)
			if e != nil {
				w.log(fmt.Sprintf("Failed to add item \"%s\" to %s's %s client with message: \"%v\".", t.Title, v.Label(), v.Name(), e), 1)
				return
			}
		}
	}
	w.log(fmt.Sprintf("%7.2fms Item \"%s\" done.", time.Since(startT).Seconds()*1000.0, t.Title), 1)
}

func (w *worker) getTorrent(t torrents.Individ) ([]byte, string, error) {
	req, err := http.NewRequest("GET", t.DUrl, nil)
	if err != nil {
		return nil, "", fmt.Errorf("%v\n", err)
	}
	if w.Config.Cookie != "" {
		req.Header.Add("Cookie", w.Config.Cookie)
	}

	resp, err := w.client.Do(req)
	for try := 0; try < 3; {
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if err == nil {
			resp.Body.Close()
		} // StatusCode != 200
		time.Sleep(200 * time.Millisecond)
		resp, err = w.client.Do(req)
		try++
	}
	if err != nil {
		return nil, "", fmt.Errorf("Failed to download torrent file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("Item \"%s\" met status code %d.", t.Title, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to dump byte data: %v", err)
	}
	return body, rss.GetFileInfo(t.DUrl, resp.Header), nil
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

func savehistory(path string) error {
	if f, err := os.Create(path); err != nil {
		return err
	} else {
		f.Close()
	}
	return nil
}
