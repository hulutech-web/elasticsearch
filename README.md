# elasticsearch

## 一、安装
```bash
go get -u github.com/hulutech-web/elasticsearch

```
#### 1.1 注册服务提供者:config/app.go
```go
import	"github.com/hulutech-web/elasticsearch"

func init() {
"providers": []foundation.ServiceProvider{
	....
	&elasticsearch.ServiceProvider{},
}
}
```

#### 1.2发布资源
```go
go run . artisan vendor:publish --package=github.com/hulutech-web/elasticsearch
```

## 二、使用
#### 2.1 使用说明:es连接配置
发布资源后，config/elasticsearch.go中的配置文件中有默认的配置项信息，请自行修改
```go
config.Add("elasticsearch", map[string]any{
    "address":  "http://localhost:9200",
    "username": "",
    "password": "",
})
```
#### 2.2 使用说明:同步es及查询示例
- 查询及删除

```go
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/contracts/http"
    "github.com/hulutech-web/elasticsearch"
	"log"
)

type EsController struct {
	Elasticsearch
	// Dependent services
}

func NewEsController() *EsController {
	client, _ := elasticsearch.NewDefaultClient()
	//client := &elasticsearch.NewClient(elasticsearch.Config{}
	es := NewElasticsearch(client)
	return &EsController{
		//Inject services
		Elasticsearch: *es,
	}
}

func (r *EsController) Index(ctx http.Context) http.Response {
	content := ctx.Request().Query("content")
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  content,
				"fields": []string{"title", "subtitle", "content", "author"},
			},
		},
		"highlight": map[string]interface{}{
			"pre_tags":  []string{"<span class='highlight'>"},
			"post_tags": []string{"</span>"},
			"fields": map[string]interface{}{
				"title": map[string]interface{}{
					"fragment_size":       100,
					"number_of_fragments": 1,
				},
				"subtitle": map[string]interface{}{
					"fragment_size":       100,
					"number_of_fragments": 1,
				},
				"content": map[string]interface{}{
					"fragment_size":       200,
					"number_of_fragments": 3,
				},
				"author": map[string]interface{}{
					"fragment_size":       50,
					"number_of_fragments": 1,
				},
			},
		},
	}
	// query转换为str
	queryStr, err := json.Marshal(query)
	if err != nil {
		log.Fatalf("Error marshalling query: %s", err)
	}
	index := "article"
	resp, err := r.SearchDocuments(ctx, string(queryStr), index)
	if err != nil {
		log.Fatalf("Error searching documents: %s", err)
	}
	var result map[string]interface{}
	defer resp.Body.Close()
	if resp.IsError() {
		log.Printf("[%s] Error searching documents: %s", resp.Status(), resp.String())
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Fatalf("Error decoding response: %s", err)
		}
		// 提取高亮部分并添加到结果中
		hits, ok := result["hits"].(map[string]interface{})
		if ok {
			hitList, ok := hits["hits"].([]interface{})
			if ok {
				for i := range hitList {
					hit, ok := hitList[i].(map[string]interface{})
					if ok {
						highlight, ok := hit["highlight"].(map[string]interface{})
						if ok {
							hit["highlighted"] = highlight
							delete(hit, "highlight")
						}
					}
				}
			}
		}
		fmt.Printf("Search results: %+v\n", result)
	}
	return ctx.Response().Json(http.StatusOK, result)
}

func (r *EsController) Destroy(ctx http.Context) http.Response {
	index := ctx.Request().Input("index")
	docID := ctx.Request().Input("docID")
	r.DeleteDocument(ctx, index, docID)
	return ctx.Response().Success().Json(map[string]any{
		"message": "success",
		"data":    "删除成功",
	})
}
```
- 同步
```go
func (r *ArticleController) Store(ctx http.Context) http.Response {
	var articleRequest requests.ArticleRequest
	errors, err := ctx.Request().ValidateRequest(&articleRequest)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", err.Error())
	}
	if errors != nil {
		return httpfacade.NewResult(ctx).ValidError("", errors.All())
	}
	article := models.Article{
		Title:    articleRequest.Title,
		Subtitle: articleRequest.Subtitle,
		Content:  articleRequest.Content,
		Author:   articleRequest.Author,
	}
	//todo add request values
	facades.Orm().Query().Model(&models.Article{}).Create(&article)
	docID := fmt.Sprintf("%d", article.ID)
	index := "article"
	resp, err := r.IndexDocument(ctx, index, docID, article)
	if err != nil {
		log.Fatalf("Error indexing document: %s", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		log.Printf("[%s] Error indexing document: %s", resp.Status(), resp.String())
	} else {
		fmt.Printf("Document %s indexed successfully in index %s\n", docID, index)
	}
	return httpfacade.NewResult(ctx).Success("创建成功", nil)
}
```