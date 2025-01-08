package elasticsearch

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"strings"
)

type Elasticsearch struct {
	client *elasticsearch.Client
}

func NewElasticsearch(client *elasticsearch.Client) *Elasticsearch {
	return &Elasticsearch{
		client: client,
	}
}

func (e *Elasticsearch) Search(ctx context.Context, index string, query string) (*esapi.Response, error) {
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
	}
	return req.Do(ctx, e.client)
}

// 创建索引
func (e *Elasticsearch) CreateIndex(ctx context.Context, index string) (*esapi.Response, error) {
	req := esapi.IndicesCreateRequest{
		Index: index,
	}
	return req.Do(ctx, e.client)
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

// 搜索文档
func (e *Elasticsearch) SearchDocuments(ctx context.Context, query string, indexs ...string) (*esapi.Response, error) {

	req := esapi.SearchRequest{
		Index: indexs,
		Body:  strings.NewReader(query),
	}
	return req.Do(ctx, e.client)
}

// 更新文档
func (e *Elasticsearch) UpdateDocument(ctx context.Context, index string, docID string, body interface{}) (*esapi.Response, error) {
	docBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(string(docBytes)),
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
