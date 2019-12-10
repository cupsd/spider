package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/gocolly/colly/queue"

	"github.com/gocolly/colly"
)

type Movie struct {
	Title       string
	Url         string
	DownloadAdd []string
}

const (
	downloadRe = `<a style="color: #ff0000;" href="(.+)?">.*?</a>`
)

func main() {

	movies := []*Movie{}
	c := colly.NewCollector()

	q, _ := queue.New(15, &queue.InMemoryQueueStorage{MaxSize: 10000})

	pageCount, err := getPageCount()
	if err != nil {
		log.Fatal(err)
	}

	c.OnHTML("div.article h2 a", func(e *colly.HTMLElement) {
		movie := &Movie{}
		movie.Title = e.Text
		url := e.Attr("href")
		movie.Url = url

		subMatch, err := getDownloadAdds(url)
		if err != nil {
			log.Println(err)
			return
		}

		// 正则匹配失败，没有下载链接
		if subMatch == nil {
			movie.DownloadAdd = append(movie.DownloadAdd, "没有下载链接或链接已过期")
			return
		}

		// 将多个链接加入到切片中
		for _, match := range subMatch {
			movie.DownloadAdd = append(movie.DownloadAdd, match[1])
		}

		movies = append(movies, movie)
	})

	// 将url加入到队列中
	for i := 1; i <= pageCount; i++ {
		if i == 1 {
			q.AddURL("http://gaoqing.la/1080p")
		}

		q.AddURL(fmt.Sprintf("http://gaoqing.la/1080p/page/%d", i))
	}

	q.Run(c)

	data, err := json.Marshal(movies)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("1.txt", data, 0644)
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

// 获取页面
func getPage(url string) (data []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("return code is not 200")
		return
	}

	e := determineEncoding(resp.Body)

	Reader := transform.NewReader(resp.Body, e.NewDecoder())

	data, err = ioutil.ReadAll(Reader)
	if err != nil {
		return
	}
	return
}

// 自动决定使用哪种编码格式
func determineEncoding(r io.Reader) encoding.Encoding {

	bytes, err := bufio.NewReader(r).Peek(1024)
	if err != nil {
		log.Fatal(err)
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")

	return e
}

// 获取所有的下载链接
func getDownloadAdds(url string) (matchs [][]string, err error) {
	var data []byte
	re := regexp.MustCompile(downloadRe)
	data, err = getPage(url)
	if err != nil {
		return
	}

	matchs = re.FindAllStringSubmatch(string(data), -1)
	if matchs == nil {
		return
	}

	return
}
