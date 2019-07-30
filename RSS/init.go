package RSS

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/capric98/GoRSS/client"
	"github.com/fatih/color"
)

//color.New(color.FgYellow).SprintFunc()
//color.New(color.FgRed).SprintFunc()
var (
	cGreen = color.New(color.FgWhite, color.BgGreen)
	cRed   = color.New(color.FgHiWhite, color.BgRed)
	oLock  = sync.Mutex{}
)

func CheckRegexp(v RssRespType, reg []*regexp.Regexp) bool {
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

func SaveItem(r RssRespType, t TaskType) {
	nClient := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	startT := time.Now()

	resp, err := nClient.Get(r.DURL)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%v\n", err)
	}

	if t.DownPath != "" {
		err := ioutil.WriteFile(t.DownPath+"/"+GetFileInfo(r.DURL, resp.Header), body, 0644)
		if err != nil {
			log.Printf("Warning: %v\n", err)
		}
		//log.Printf("Item \"%s\" is saved as \"%s\"\n", r.Title, GetFileInfo(r.DURL, resp.Header))
	}

	// Add file to client.
	if t.Client != nil {
		for _, v := range t.Client {
			switch v.Name {
			case "qBittorrent":
				err = v.Client.(client.QBType).Add(body, GetFileInfo(r.DURL, resp.Header))
			case "Deluge":
				//err= v.Client.(client.DeType).Add(body)
			}
			if err != nil {
				log.Printf("%s: Failed to add item \"%s\" to %s client with message: \"%v\".\n", t.TaskName, r.Title, v.Name, err)
			}
		}
	}
	PrintTimeInfo(fmt.Sprintf("Item \"%s\" uses ", r.Title), time.Since(startT))
	return
}

func PrintTimeInfo(info string, t time.Duration) {
	oLock.Lock()
	fmt.Fprintf(os.Stderr, time.Now().Format("2006/01/02 15:04:05 "))
	fmt.Fprintf(os.Stderr, info)
	color.Set(color.FgWhite, color.BgGreen)
	fmt.Fprintf(os.Stderr, "%s", t)
	color.Unset()
	fmt.Fprintf(os.Stderr, ".\n")
	oLock.Unlock()
}

func RunTask(t TaskType) {
	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	for {
		log.Printf("Run task: %s\n", t.TaskName)
		startT := time.Now()

		Rresp, err := RssFetch(t.RSS_Link, &client)
		if err != nil {
			log.Printf("Caution: Task %s failed to get RSS data and raised an error: %v.\n", t.TaskName, err)
			continue
		}
		ac_count := 0
		rj_count := 0
		for _, v := range Rresp {
			// Check if item had been accepted yet.
			if v.GUID == "" {
				v.GUID = NameRegularize(v.Title)
			} else {
				v.GUID = NameRegularize(v.GUID)
			} // Just in case.
			if _, err := os.Stat(".RSS-saved/" + v.GUID); !os.IsNotExist(err) {
				rj_count++
				continue
			}
			os.Create(".RSS-saved/" + v.GUID)

			// Check regexp filter.
			if (t.RjcRegexp != nil) && (CheckRegexp(v, t.RjcRegexp)) {
				log.Printf("%s: Reject item \"%s\"\n", t.TaskName, v.Title)
				rj_count++
				continue
			}
			if t.AccRegexp != nil && (!CheckRegexp(v, t.AccRegexp)) && (t.Strict) {
				log.Printf("%s: Cannot accept item \"%s\" due to strict mode.\n", t.TaskName, v.Title)
				rj_count++
				continue
			}

			// Check content_size.
			if !((v.Length > t.MinSize && v.Length < t.MaxSize) || (v.Length == 0 && !t.Strict)) {
				log.Printf("%s: Reject item \"%s\" due to content_size not fit.\n", t.TaskName, v.Title)
				//log.Println(v.Length, "vs", t.MinSize, t.MaxSize)
				rj_count++
				continue
			}

			log.Printf("%s: Accept item \"%s\"\n", t.TaskName, v.Title)
			ac_count++

			go SaveItem(v, t)
		}

		PrintTimeInfo(fmt.Sprintf("Accept %d item(s), reject %d item(s). Task %q costs ", ac_count, rj_count, t.TaskName), time.Since(startT))
		time.Sleep(time.Duration(t.Interval) * time.Second)
	}
}

func Init(conf string) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cdata, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	if _, err := os.Stat(".RSS-saved"); os.IsNotExist(err) {
		os.Mkdir(".RSS-saved", 0644)
		log.Println(".RSS-saved dir did not exist, make it!")
	}
	taskList := ParseSettings(cdata)

	qsignal := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		qsignal <- fmt.Errorf("%s", <-c)
	}()

	for _, t := range taskList {
		go RunTask(t)
	}
	<-qsignal
	log.Println("Receive signal 2, quit the program.")
	log.Println("Bye.")
}
