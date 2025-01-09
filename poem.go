package elasticsearch

import (
	"encoding/json"
	"fmt"
	"github.com/goravel/framework/facades"
	"goravel/app/models"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// 测试数据
const (
	apiURL       = "https://whyta.cn/api/tx/poetries"
	apiKey       = "96f163cda80b"
	defaultCount = 100
	maxOccurs    = 5
	queryNum     = 10
	sleepTime    = 5000
)

type GoLimit struct {
	ch chan struct{}
}

func NewGoLimit(max int) *GoLimit {
	return &GoLimit{ch: make(chan struct{}, max)}
}

func (g *GoLimit) Add() {
	g.ch <- struct{}{}
}

func (g *GoLimit) Done() {
	<-g.ch
	fmt.Println(fmt.Sprintf("📢并发通道数：（%v/%v）\n", len(g.ch), maxOccurs))
}

type Ret struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result struct {
		Curpage int    `json:"curpage"`
		Allnum  int    `json:"allnum"`
		List    []Poem `json:"list"`
	} `json:"result"`
}

type Poem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

func HuntPoem() {
	g := NewGoLimit(maxOccurs)
	var wg sync.WaitGroup

	for i := 0; i < defaultCount; i++ {
		page := i + 1
		wg.Add(1)
		g.Add() // 在启动新的goroutine前添加到通道
		go func(g *GoLimit, index int) {
			defer func() {
				g.Done()
				wg.Done()
			}()

			time.Sleep(time.Duration(sleepTime) * time.Millisecond)

			client := &http.Client{}
			reqURL := fmt.Sprintf("%s?key=%s&num=%d&page=%d", apiURL, apiKey, queryNum, page)
			req, err := http.NewRequest("GET", reqURL, nil)
			if err != nil {
				fmt.Println("创建请求失败:", err)
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("发送请求失败:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("请求失败，状态码: %d\n", resp.StatusCode)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取响应体失败:", err)
				return
			}

			var ret Ret
			err = json.Unmarshal(body, &ret)
			if err != nil {
				fmt.Println("反序列化失败:", err)
				return
			}

			articles := make([]models.Article, 0, len(ret.Result.List))
			for _, r := range ret.Result.List {
				articles = append(articles, models.Article{
					Title:   r.Title,
					Content: r.Content,
					Author:  r.Author,
				})
			}

			if err_ := facades.Orm().Query().Model(&models.Article{}).Create(&articles); err_ != nil {
				fmt.Println("插入数据库失败:", err_.Error())
				return
			}
		}(g, i)
	}

	wg.Wait()
}
