package facades

import (
	"log"

	"goravel/packages/elasticsearch"
	"goravel/packages/elasticsearch/contracts"
)

func Elasticsearch() contracts.Elasticsearch {
	instance, err := elasticsearch.App.Make(elasticsearch.Binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	return instance.(contracts.Elasticsearch)
}
