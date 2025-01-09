package contracts

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Elasticsearch interface {
	SearchDocuments(ctx context.Context, query string) (*esapi.Response, error)
	PushIndex(ctx context.Context, indexs []string) (*esapi.Response, error)
	GetDocument(ctx context.Context, index string, docID string) (*esapi.Response, error)
	UpdateDocument(ctx context.Context, index string, docID string, body interface{}) (*esapi.Response, error)
	DeleteDocument(ctx context.Context, index string, docID string) (*esapi.Response, error)
	DeleteIndex(ctx context.Context, index string) (*esapi.Response, error)
	DeleteIndexDocuments(ctx context.Context, index string) (*esapi.Response, error)
}
