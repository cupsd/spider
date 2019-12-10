package main

import (
	"fmt"
	"log"
	"path"
	"strconv"

	"github.com/gocolly/colly/queue"

	"github.com/gocolly/colly"
)

// TODO: 获取电影下载链接
type Movie struct {
	Title string
	Url   string
}

func main() {

	movies := []*Movie{}
	c := colly.NewCollector()

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
		fmt.Printf("%#v\n", movie)
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
