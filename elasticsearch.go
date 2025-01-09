package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/color"
	"log"
	"strings"
)

type Elasticsearch struct {
	client *elasticsearch.Client
	indexs []string
}

func NewElasticsearch(client *elasticsearch.Client) *Elasticsearch {
	indexsStr := facades.Config().Get("elasticsearch.tables")
	if val, ok := indexsStr.([]string); ok {
		return &Elasticsearch{
			client: client,
			indexs: val,
		}
	}
	return nil
}

// 搜索文档
func (e *Elasticsearch) SearchDocuments(ctx context.Context, query string) (*esapi.Response, error) {
	req := esapi.SearchRequest{
		Index: e.indexs,
		Body:  strings.NewReader(query),
	}
	return req.Do(ctx, e.client)
}

// 创建索引
func (e *Elasticsearch) InitIndex(ctx context.Context, indexs []string) (*esapi.Response, error) {
	req := esapi.IndicesExistsRequest{
		Index: indexs,
	}

	resp, err := req.Do(ctx, e.client)
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		if resp.StatusCode == 404 {
			color.Blue().Println(fmt.Printf("Index %s does not exist\n", indexs))
			for _, index := range indexs {
				req_crt := esapi.IndicesCreateRequest{
					Index: index,
				}
				resp_crt, err_ := req_crt.Do(ctx, e.client)
				if err_ != nil {
					log.Fatalf("Error creating index: %s", err_)
				} else {
					if resp_crt.IsError() {
						log.Printf("[%s] Error creating index: %s", resp_crt.Status(), resp_crt.String())
					} else {
						fmt.Printf("Index %s created\n", index)
					}
				}
			}
		} else {
			log.Printf("[%s] Error checking index existence: %s", resp.Status(), resp.String())
			return nil, err
		}
	} else {
		color.Green().Println("ES bootstrap..." + fmt.Sprintf("%v", indexs) + " index exists")
		return nil, nil
	}
	return nil, nil
}

// 索引文档
func (e *Elasticsearch) IndexDocument(ctx context.Context, index string, docID string, body interface{}) (*esapi.Response, error) {
	docBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(string(docBytes)),
		Refresh:    "true",
		Pretty:     true,
	}
	return req.Do(ctx, e.client)
}

// 获取文档
func (e *Elasticsearch) GetDocument(ctx context.Context, index string, docID string) (*esapi.Response, error) {
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: docID,
	}
	return req.Do(ctx, e.client)
}

// 更新文档
func (e *Elasticsearch) UpdateDocument(ctx context.Context, index string, docID string, body interface{}) (*esapi.Response, error) {
	docBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(string(docBytes)),
		OpType:     "index",
	}
	return req.Do(ctx, e.client)
}

// 删除文档
func (e *Elasticsearch) DeleteDocument(ctx context.Context, index string, docID string) (*esapi.Response, error) {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: docID,
	}
	return req.Do(ctx, e.client)
}

// 删除索引
func (e *Elasticsearch) DeleteIndex(ctx context.Context, index string) (*esapi.Response, error) {
	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}
	return req.Do(ctx, e.client)
}

// 删除索引文章
func (e *Elasticsearch) DeleteIndexDocuments(ctx context.Context, index string) (*esapi.Response, error) {
	req := esapi.DeleteByQueryRequest{
		Index: []string{index},
	}
	return req.Do(ctx, e.client)
}
