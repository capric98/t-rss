package RSS

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
)

func RssFetch(url string, client *http.Client) {

	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("Failed to get rss meta: %v", err)
		return
	}
	// bodyData, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatalf("Failed to read rss resp body: %v", err)
	// 	return
	// }
	// //fmt.Println(string(bodyData))
	// ioutil.WriteFile("resp.conf", bodyData, 0666)

	fp := gofeed.NewParser()
	rssFeed, _ := fp.Parse(resp.Body)
	file, _ := os.OpenFile("resp.conf", os.O_CREATE|os.O_WRONLY, 0644)
	for _, v := range rssFeed.Items {
		fmt.Println("==========================================================")
		fmt.Println(v.Title)
		//fmt.Println(v.Description)
		file.WriteString(html.UnescapeString(v.Description))
		fmt.Println(v.Author)
		fmt.Println(v.Categories)
		fmt.Println(v.Enclosures[0].URL)
		fmt.Println(v.Published)
	}
	file.Close()
}

func RssRun(task TaskType) {
	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	RssFetch(task.RSS_Link, &client)
}
