package elasticsearch

import (
	"encoding/json"
	"fmt"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"log"
)

func ElasticSearch(app foundation.Application) {
	go StartCanalSync()

	router := app.MakeRoute()
	//操作面板
	router.Get("/es", func(ctx http.Context) http.Response {
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
		fmt.Sprintf("%s", queryStr)
		resp, err := ES.SearchDocuments(ctx, string(queryStr))
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
		}
		return ctx.Response().Json(http.StatusOK, result)
	})

	router.Post("/es/sync", func(ctx http.Context) http.Response {
		table := ctx.Request().Input("table")
		model_id := ctx.Request().Input("model_id")
		docID := fmt.Sprintf("%v", model_id)
		result := map[string]interface{}{}
		facades.Orm().Query().Table(table).Where("id=?", docID).Scan(&result)
		resp, err := ES.IndexDocument(ctx, table, docID, result)
		if err != nil {
			log.Fatalf("Error indexing document: %s", err)
		}
		defer resp.Body.Close()
		if resp.IsError() {
			log.Printf("[%s] Error indexing document: %s", resp.Status(), resp.String())
		} else {
			fmt.Printf("Document %s indexed successfully in index %s\n", docID, table)
		}
		return ctx.Response().Success().Json(http.Json{
			"message": "同步成功",
		})
	})
}
