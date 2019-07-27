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
	"time"
)

func CheckRegexp(v RssRespType, reg []string) bool {
	for _, r := range reg {
		matched, _ := regexp.MatchString(r, v.Title)
		if matched {
			return true
		}
		matched, _ = regexp.MatchString(r, v.Description)
		if matched {
			return true
		}
		matched, _ = regexp.MatchString(r, v.Author)
		if matched {
			return true
		}
	}
	return false
}

func SaveItem(r RssRespType, t TaskType) {
	nClient := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
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
		err := ioutil.WriteFile(t.DownPath+GetFileInfo(r.DURL, resp.Header), body, 0644)
		if err != nil {
			log.Printf("Warning: %v\n", err)
		}
		log.Printf("Item \"%s\" is saved as \"%s\"\n", r.Title, GetFileInfo(r.DURL, resp.Header))
	}

	// Add file to client.
	return
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
			log.Printf("Caution: Task %s failed to get RSS response and raised an error: %v.\n", t.TaskName, err)
			continue
		}
		for _, v := range Rresp {
			if (t.RjcRegexp != nil) && (CheckRegexp(v, t.RjcRegexp)) {
				log.Printf("%s: Reject item \"%s\"\n", t.TaskName, v.Title)
				continue
			}
			if t.AccRegexp != nil && (!CheckRegexp(v, t.RjcRegexp)) && (t.Strict) {
				log.Printf("%s: Cannot accept item \"%s\" due to strict mode.\n", t.TaskName, v.Title)
				continue
			}

			// Check content_size here!!!
			// Check history here!!!!

			log.Printf("%s: Accept item \"%s\"\n", t.TaskName, v.Title)
			go SaveItem(v, t)
		}

		log.Printf("Task %s executed once in %s.\n", t.TaskName, time.Since(startT))
		log.Printf("Sleep %d seconds.\n", t.Interval)
		time.Sleep(time.Duration(t.Interval) * time.Second)
	}
}

func Init(conf string) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cdata, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
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
