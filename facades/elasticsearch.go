package facades

import (
	"github.com/hulutech-web/elasticsearch"
	"github.com/hulutech-web/elasticsearch/contracts"
	"log"
)

func Elasticsearch() contracts.Elasticsearch {
	instance, err := elasticsearch.App.Make(elasticsearch.Binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	return instance.(contracts.Elasticsearch)
}
