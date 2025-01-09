# elasticsearch

## 一、安装
```bash
go clone github.com/hulutech-web/elasticsearch

```
#### 1.1 注册服务提供者:config/app.go
```go

func init() {
"providers": []foundation.ServiceProvider{
	....
	&elasticsearch.ServiceProvider{},
}
}
```

#### 1.2发布资源
```go
go run . artisan vendor:publish --package=./packages/elasticsearch
```

## 二、使用
#### 2.1 使用说明:es连接配置
发布资源后，config/elasticsearch.go中的配置文件中有默认的配置项信息，请自行修改
```go
config.Add("elasticsearch", map[string]any{
    "address":  "http://localhost:9200",
    "username": "",
    "password": "",
    "schema":   "goravel",
    "canal":    true,  // 是否开启canal
    "log":      false, // 是否开启日志
    "tables": []string{
    "articles",//索引的表名
    "posts", //索引的表名
    },
})
```
#### 2.2 使用说明:同步es及查询示例
- 查询及删除

```go
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/contracts/http"
	elasticfacades "goravel/packages/elasticsearch/facades"
	"log"
)

type EsController struct {
	// Dependent services
}

func NewEsController() *EsController {
	//client := &elasticsearch.NewClient(elasticsearch.Config{}
	return &EsController{
	}
}

// 查询出数据，并按照关键词进行高亮显示，给定一个html的class类名为highlight,前端请自行添加高亮的样式
func (r *EsController) Index(ctx http.Context) http.Response {
  content := ctx.Request().Query("content")
  fields := ctx.Request().QueryArray("fields")
  query := map[string]interface{}{
    "query": map[string]interface{}{
      "multi_match": map[string]interface{}{
        "query":  content,
        "fields": fields,
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
    return ctx.Response().Json(http.StatusOK, result)
  }
  instance:= elasticfacades.Elasticsearch()
  resp, err := instance.SearchDocuments(ctx, string(queryStr))
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

func (r *EsController) Destroy(ctx http.Context) http.Response {
	index := ctx.Request().Input("index")
	docID := ctx.Request().Input("docID")
    _, err := elasticfacades.Elasticsearch().DeleteDocument(ctx, index, docID)
    if err != nil {
      ctx.Response().Json(http.StatusInternalServerError,map[string]interface{}{
        "error":err.Error(),
      }).Abort()
    }
	return ctx.Response().Success().Json(map[string]any{
		"message": "success",
		"data":    "删除成功",
	})
}
```
- 同步
```go
func (r *ArticleController) Store(ctx http.Context) http.Response {
	var articleRequest requests.ArticleRequest
	errors, err := ctx.Request().ValidateRequest(&articleRequest)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", err.Error())
	}
	if errors != nil {
		return httpfacade.NewResult(ctx).ValidError("", errors.All())
	}
	article := models.Article{
		Title:    articleRequest.Title,
		Subtitle: articleRequest.Subtitle,
		Content:  articleRequest.Content,
		Author:   articleRequest.Author,
	}
	//todo add request values
	facades.Orm().Query().Model(&models.Article{}).Create(&article)
	docID := fmt.Sprintf("%d", article.ID)
    elasticfacades.Elasticsearch().PushIndex(ctx, []string{"article"})
	if err != nil {
		log.Fatalf("Error indexing document: %s", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		log.Printf("[%s] Error indexing document: %s", resp.Status(), resp.String())
	} else {
		fmt.Printf("Document %s indexed successfully in index %s\n", docID, index)
	}
	return httpfacade.NewResult(ctx).Success("创建成功", nil)
}
```

#### 2.3 自动同步
为方便的使用es的能力，采用canal扩展包来监听mysql数据库的变化，当数据发生变化时，自动同步到es中。这样与业务进行解耦，避免了不必要的操作。
- 安装canal:阿里巴巴 MySQL binlog 增量订阅&消费组件
下载安装包：[git地址](https://github.com/alibaba/canal/releases)
- 解压后，修改配置文件
  1. canal配置文件``canal.properties``:
    ```properties
    # tcp bind ip
    canal.ip = 127.0.0.1
    # register ip to zookeeper
    canal.register.ip =
    canal.port = 11111
    canal.metrics.pull.port = 11112
    # canal instance user/passwd
    # canal.user = canal
    # canal.passwd =
    
    
    # MySQL主库地址
    canal.instance.master.address = 127.0.0.1:3306
    # MySQL用户名
    canal.instance.dbUsername = root
    # MySQL密码
    canal.instance.dbPassword = Dazhouquxian2012.
    
    canal.destinations = example
    ```
  2. canal配置文件``example/instance.properties``:
    ```properties
    # position info
    canal.instance.master.address=127.0.0.1:3306
    # table regex
    canal.instance.filter.regex=goravel\\.articles,goravel\\.roles,goravel\\.users
    # table black regex
    canal.instance.filter.black.regex=mysql\\.slave_.*
    ```
    
  
- mysql配置binlog,``my.cnf``文件
```properties
# Default Homebrew MySQL server config
[mysqld]
# Only allow connections from localhost
bind-address = 127.0.0.1
mysqlx-bind-address = 127.0.0.1
server_id = 1  #配置mysql replication需要定义，不能和canal的slaveId重复
binlog-format = ROW
log-bin = mysql-bin #开启binlog
```
- mysql用户权限配置
```sql
use `goravel`;
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'root'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON goravel.* TO 'root'@'localhost';
FLUSH PRIVILEGES;
```

- 安装扩展golang包
```go
go get -u github.com/withlin/canal-go/client
```
- 启动canal:
``yourpath/canal.deployer-1.1.8-SNAPSHOT/bin`` 下执行,``./startup.sh``命令启动canal，默认端口为11111，关闭则执行``./stop.sh``命令
```shell
sh ./startup.sh
```
- 同步日志打印
```bash
【ES INFO】================> binlog[mysql-bin.000003 : 55642],Schema:[goravel],tablename:[articles],docID:[1] eventType: INSERT
id : 1  update= true
title : 出塞二首  update= true
subtitle : 出塞二首  update= true
content : 秦时明月汉时关，万里长征人未还。 但使龙城飞将在，不教胡马度阴山。 骝马新跨白玉鞍，战罢沙场月色寒。 城头铁鼓声犹震，匣里金刀血未干。  update= true
author : 王昌龄〔唐代〕  update= true
created_at : 2025-01-09 12:11:12  update= true
updated_at : 2025-01-09 12:11:12  update= true
deleted_at :   update= true
```
