package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strconv"

	"github.com/gocolly/colly/queue"

	"github.com/gocolly/colly"
)

type Movie struct {
	Title       string
	Url         string
	DownloadAdd []string
}

func main() {
	var d *colly.Collector
	movies := []*Movie{}
	c := colly.NewCollector()

	d = c.Clone()

	q, _ := queue.New(15, &queue.InMemoryQueueStorage{MaxSize: 10000})

	pageCount, err := getPageCount()
	if err != nil {
		log.Fatal(err)
	}

	c.OnHTML("div.article h2", func(e *colly.HTMLElement) {
		// 获取电影链接
		e.ForEach("a", func(_ int, el *colly.HTMLElement) {
			movies = append(movies, &Movie{
				Url:   el.Attr("href"),
				Title: e.Text,
			})
		})
	})

	// 将url加入到队列中
	for i := 1; i <= pageCount; i++ {
		if i == 1 {
			q.AddURL("http://gaoqing.la/1080p")
		}

		q.AddURL(fmt.Sprintf("http://gaoqing.la/1080p/page/%d", i))
	}

	q.Run(c)

	for _, movie := range movies {
		getDownloadAddr(d, movie)
	}

	// file,err := os.OpenFile("movie.txt",os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	data, err := json.Marshal(movies)

	err = ioutil.WriteFile("movie.txt", data, 0644)
	if err != nil {
		log.Fatal(err)
	}

}

// 获取总页数
func getPageCount() (count int, err error) {

	c := colly.NewCollector()

	c.OnHTML("div.pagination a.extend", func(e *colly.HTMLElement) {
		count, err = strconv.Atoi(path.Base(e.Attr("href")))
		if err != nil {
			return
		}
	})

	c.Visit("http://gaoqing.la/1080p")

	return
}

func getDownloadAddr(d *colly.Collector, movie *Movie) {

	d.OnHTML("div#post_content p", func(e *colly.HTMLElement) {
		// fmt.Println(e.ChildAttrs("span a", "href"))
		movie.DownloadAdd = append(movie.DownloadAdd, e.ChildAttrs("span a", "href")...)
	})

	d.Visit(movie.Url)
}
