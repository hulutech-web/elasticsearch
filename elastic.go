package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	gcolor "github.com/gookit/color"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/color"
	"log"
	"strings"
	"sync"
)

var (
	once sync.Once
	ES   *Elastic
)

type Elastic struct {
	client *elasticsearch.Client
	indexs []string
}

// 获取并处理elasticsearch.tables配置
func GetElasticsearchTables() ([]string, error) {
	tablesInterface := facades.Config().Get("elasticsearch.tables")
	tables, ok := tablesInterface.(map[string][]string)
	if !ok {
		return nil, fmt.Errorf("Invalid configuration for 'elasticsearch.tables'")
	}

	indexs := make([]string, 0, len(tables))
	for index := range tables {
		indexs = append(indexs, index)
		//fmt.Printf("Index: %s, Fields: %v\n", index, tables[index])
	}
	return indexs, nil
}

// 创建Elasticsearch客户端
func createElasticsearchClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{facades.Config().GetString("elasticsearch.address")},
		Username:  facades.Config().GetString("elasticsearch.username"),
		Password:  facades.Config().GetString("elasticsearch.password"),
	}
	return elasticsearch.NewClient(cfg)
}
func viewPackage() {
	str := fmt.Sprintf(" ▗▄▄▖ ▗▄▖ ▗▄▄▖  ▗▄▖ ▗▖  ▗▖▗▄▄▄▖▗▖       ▗▄▄▄▖▗▖   ▗▄▄▄▖ ▗▄▄▖ ▗▄▄▖▗▄▄▄▖▗▄▄▄▖ ▗▄▄▖\n▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌  ▐▌▐▌   ▐▌       ▐▌   ▐▌   ▐▌   ▐▌   ▐▌     █    █  ▐▌   \n▐▌▝▜▌▐▌ ▐▌▐▛▀▚▖▐▛▀▜▌▐▌  ▐▌▐▛▀▀▘▐▌       ▐▛▀▀▘▐▌   ▐▛▀▀▘ ▝▀▚▖▐▌     █    █  ▐▌   \n▝▚▄▞▘▝▚▄▞▘▐▌ ▐▌▐▌ ▐▌ ▝▚▞▘ ▐▙▄▄▖▐▙▄▄▖    ▐▙▄▄▖▐▙▄▄▖▐▙▄▄▖▗▄▄▞▘▝▚▄▄▖  █  ▗▄█▄▖▝▚▄▄▖\n                                                                                \n                                                                                \n                                                                                ")
	color.Green().Println(str)
}

func NewElastic(ctx context.Context) (*Elastic, error) {

	var err error

	once.Do(func() {
		var client *elasticsearch.Client
		client, err = createElasticsearchClient()
		if err != nil {
			color.Red().Println(fmt.Sprintf("Error creating Elasticsearch client: %s", err.Error()))
			return
		}

		var indexs []string
		indexs, err = GetElasticsearchTables()
		if err != nil {
			color.Red().Println(fmt.Sprintf("Elastic services maybe stoped,Error getting Elasticsearch tables: %v", err.Error()))
			return
		}

		ES = &Elastic{
			client: client,
			indexs: indexs,
		}
		viewPackage()

		_, err := ES.PushIndex(ctx, indexs)
		if err != nil {
			color.Red().Println(fmt.Sprintf("Error pushing index: %v", err.Error()))
		}
	})
	if err != nil {
		return nil, err
	}
	return ES, nil
}

// 搜索文档
func (e *Elastic) SearchDocuments(ctx context.Context, query string) (*esapi.Response, error) {
	req := esapi.SearchRequest{
		Index: e.indexs,
		Body:  strings.NewReader(query),
	}
	return req.Do(ctx, e.client)
}

// 创建索引
func (e *Elastic) PushIndex(ctx context.Context, indexs []string) (*esapi.Response, error) {
	req := esapi.IndicesExistsRequest{
		Index: indexs,
	}

	resp, err := req.Do(ctx, e.client)
	if err != nil {
		color.Red().Println(fmt.Sprintf("Elastic services maybe stoped,Error checking if index exists: %s", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		if resp.StatusCode == 404 {
			color.Red().Println(fmt.Printf("Index %s does not exist\n", indexs))
			for _, index := range indexs {
				req_crt := esapi.IndicesCreateRequest{
					Index: index,
				}
				resp_crt, err_ := req_crt.Do(ctx, e.client)
				if err_ != nil {
					color.Yellow().Println(fmt.Sprintf("Error creating index: %s", err_))
					return nil, err_
				} else {
					if resp_crt.IsError() {
						color.Yellow().Println(fmt.Sprintf("[%s] Error creating index: %s", resp_crt.Status(), resp_crt.String()))
						return nil, err_
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
		gcolor.Success.Tips(fmt.Sprintf("Elastic services start up..."))
		gcolor.Success.Tips(fmt.Sprintf("Elastic Indexs %s has be loaded!!!", indexs))
		return nil, nil
	}
	return nil, nil
}

// 索引文档
func (e *Elastic) IndexDocument(ctx context.Context, index string, docID string, body interface{}) (*esapi.Response, error) {
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
func (e *Elastic) GetDocument(ctx context.Context, index string, docID string) (*esapi.Response, error) {
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: docID,
	}
	return req.Do(ctx, e.client)
}

// 更新文档
func (e *Elastic) UpdateDocument(ctx context.Context, index string, docID string, body interface{}) (*esapi.Response, error) {
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
func (e *Elastic) DeleteDocument(ctx context.Context, index string, docID string) (*esapi.Response, error) {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: docID,
	}
	return req.Do(ctx, e.client)
}

// 删除索引
func (e *Elastic) DeleteIndex(ctx context.Context, index string) (*esapi.Response, error) {
	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}
	return req.Do(ctx, e.client)
}

// 删除索引文章
func (e *Elastic) DeleteIndexDocuments(ctx context.Context, index string) (*esapi.Response, error) {
	req := esapi.DeleteByQueryRequest{
		Index: []string{index},
	}
	return req.Do(ctx, e.client)
}
