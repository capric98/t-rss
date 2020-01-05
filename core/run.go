package core

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/capric98/t-rss/rss"
	"github.com/capric98/t-rss/torrents"
	"github.com/capric98/t-rss/torrents/bencode"
)

func (w *worker) run(wg *sync.WaitGroup) {
	defer wg.Done()

	var q Quota

	for tasks := range w.ticker {
		w.log(fmt.Sprintf("Run task: %s.", w.name), 0)
		q = w.Config.Quota

		acCount := 0
		rjCount := 0
		for _, v := range tasks {
			w.log(" + debug Title = "+v.Title, 0)
			w.log(" + debug URL   = "+v.Enclosure.Url, 0)
			w.log(" + debug Size  = "+strconv.FormatInt(v.Enclosure.Len, 10), 0)

			// Check if item had been accepted yet.
			if _, err := os.Stat(CDir + v.GUID.Value); !os.IsNotExist(err) {
				rjCount++
				continue
			} else {
				if !TestOnly {
					w.log(savehistory(CDir+v.GUID.Value), 0)
				}
			}

			// Check regexp filter.
			if result, r := checkRegexp(v, w.Config.Reject); result {
				w.log(fmt.Sprintf("%s: Reject item \"%s\", reject regexp match \"%s\".", w.name, v.Title, r), 1)
				rjCount++
				continue
			}
			if result, _ := checkRegexp(v, w.Config.Accept); (w.Config.Accept != nil) && !result {
				w.log(fmt.Sprintf("%s: Reject item \"%s\", accept regexp no match.", w.name, v.Title), 1)
				rjCount++
				continue
			}

			// Check content_size.
			if !(v.Enclosure.Len >= w.Config.Min && v.Enclosure.Len <= w.Config.Max) {
				w.log(fmt.Sprintf("%s: Reject item \"%s\" due to content_size not fit.", w.name, v.Title), 1)
				w.log(fmt.Sprintf("%d vs [%d,%d]", v.Enclosure.Len, w.Config.Min, w.Config.Max), 0)
				rjCount++
				continue
			}

			if q.Num > 0 && q.Size >= v.Enclosure.Len {
				w.log(fmt.Sprintf("%s: Accept item \"%s\"", w.name, v.Title), 1)
				acCount++
				q.Num--
				q.Size -= v.Enclosure.Len
			} else {
				rjCount++
				w.log(fmt.Sprintf("%s: Quota exceeded, drop item \"%s\"", w.name, v.Title), 1)
				continue
			}

			if !Learn {
				go w.save(v, &q)
			}
		}
		w.log(fmt.Sprintf("Task %s: Accept %d item(s), reject %d item(s).", w.name, acCount, rjCount), 0)
		if Learn {
			return
		}
	}
}

func checkRegexp(v torrents.Individ, reg []Reg) (bool, string) {
	for _, r := range reg {
		if r.R.MatchString(v.Title) {
			return true, r.C
		}
		// matched, _ = regexp.MatchString(r, v.Description)
		// if matched {
		// 	return true
		// }
		if r.R.MatchString(v.Author) {
			return true, r.C
		}
	}
	return false, ""
}

func (w *worker) save(t torrents.Individ, quota *Quota) {
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
	if t.Enclosure.Len == 0 {
		if pass, tlen := checkTLength(data, w.Config.Min, w.Config.Max); !pass {
			w.log(fmt.Sprintf("%s: Reject item \"%s\" due to TORRENT content_size not fit.", w.name, t.Title), 1)
			w.log(fmt.Sprintf("%d vs [%d,%d]", tlen, w.Config.Min, w.Config.Max), 0)
			return
		} else {
			w.log(fmt.Sprintf(" + double check: \"%s\" torrent content_size=%9d", t.Title, tlen), 0)
			if quota.Size >= t.Enclosure.Len {
				quota.Size -= t.Enclosure.Len
			} else {
				w.log(fmt.Sprintf(" + Quota exceeded, drop item \"%s\"", t.Title), 1)
			}
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
	req, err := http.NewRequest("GET", t.Enclosure.Url, nil)
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
	return body, rss.GetFileInfo(t.Enclosure.Url, resp.Header), nil
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
