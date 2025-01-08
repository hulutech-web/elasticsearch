package elasticsearch

import "github.com/goravel/framework/contracts/foundation"

const Binding = "elasticsearch"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return nil, nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("github.com/hulutech-web/elasticsearch", map[string]string{
		"config/elasticsearch.go": app.ConfigPath("elasticsearch.go"),
	})
}
