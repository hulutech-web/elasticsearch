package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("elasticsearch", map[string]any{
		"address":  "http://localhost:9200",
		"username": "",
		"password": "",
		"schema":   "goravel",
		"canal":    true,  // 是否开启canal
		"log":      false, // 是否开启日志
		"tables": map[string][]string{
			"articles": {"title", "content", "subtitle"},
			"posts":    {"title", "content"},
		},
	})
}
