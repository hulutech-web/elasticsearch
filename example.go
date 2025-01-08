package elasticsearch

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"log"
)

type ExampleController struct {
	Elasticsearch
	// Dependent services
}

func NewExampleController() *ExampleController {
	address := facades.Config().GetString("elasticsearch.address")
	username := facades.Config().GetString("elasticsearch.username")
	password := facades.Config().GetString("elasticsearch.password")
	cfg := elasticsearch.Config{
		Addresses: []string{
			address,
		},
		Username: username,
		Password: password,
	}
	client, _ := elasticsearch.NewClient(cfg)
	//or you can use the default config
	//eg: client, _ := elasticsearch.NewDefaultClient()

	es := NewElasticsearch(client)
	return &ExampleController{
		//Inject services
		Elasticsearch: *es,
	}
}

// 查询出数据，并按照关键词进行高亮显示，给定一个html的class类名为highlight,前端请自行添加高亮的样式
func (r *ExampleController) Index(ctx http.Context) http.Response {
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
