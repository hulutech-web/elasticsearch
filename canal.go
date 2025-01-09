package elasticsearch

import (
	"context"
	"fmt"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/color"
	"github.com/withlin/canal-go/client"
	pbe "github.com/withlin/canal-go/protocol/entry"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"time"
)

// StartCanalSync 启动Canal同步到ES
func StartCanalSync() error {
	ctx := context.Background()
	NewElastic(ctx)
	mode := facades.Config().GetBool("elasticsearch.canal")

	if !mode {
		return nil
	}
	//参考canal.properties配置文件中的内容，修改成自己的配置
	connector := client.NewSimpleCanalConnector("localhost", 11111, "", "", "example", 60000, 60*60*1000)
	err := connector.Connect()
	defer connector.DisConnection()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	indexs, err := GetElasticsearchTables()
	schema := facades.Config().GetString("elasticsearch.schema")
	subscribeStr := ""
	for _, v := range indexs {
		subscribeStr += schema + "\\." + v + ","
	}
	//去除最后一个逗号
	subscribeStr = subscribeStr[:len(subscribeStr)-1]
	err = connector.Subscribe(subscribeStr)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	color.Magenta().Println("===Elasticsearch is Started!!!===")
	for {

		message, err := connector.Get(100, nil, nil)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		batchId := message.Id
		if batchId == -1 || len(message.Entries) <= 0 {
			time.Sleep(300 * time.Millisecond)
			continue
		} else {
			printEntry(ctx, message.Entries)

		}
	}
}

func printEntry(ctx context.Context, entrys []pbe.Entry) {
	log_open := facades.Config().Get("elasticsearch.log")
	log_open_val := log_open.(bool)
	for _, entry := range entrys {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}
		rowChange := new(pbe.RowChange)

		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		checkError(err)
		//获取主键
		primaryKey := ""
		if len(rowChange.GetRowDatas()) > 0 {
			for _, column := range rowChange.GetRowDatas()[0].GetAfterColumns() {
				if column.GetIsKey() {
					primaryKey = column.GetValue()
					break
				}
			}
		}
		if rowChange != nil {
			eventType := rowChange.GetEventType()
			header := entry.GetHeader()
			fmt.Println(fmt.Sprintf("【ES INFO】================> binlog[%s : %d],Schema:[%s],tablename:[%s],docID:[%s] eventType: %s", header.GetLogfileName(), header.GetLogfileOffset(), header.GetSchemaName(), header.GetTableName(), primaryKey, header.GetEventType()))
			//提取rowChange.GetRowDatas()中的数据，转换为一个map结构，进行es同步
			for _, rowData := range rowChange.GetRowDatas() {
				if eventType == pbe.EventType_DELETE {
					printColumn(rowData.GetBeforeColumns())
					ES.DeleteDocument(ctx, header.GetTableName(), primaryKey)
				} else if eventType == pbe.EventType_INSERT {
					printColumn(rowData.GetAfterColumns())
					toMap := columnsToMap(rowData.GetAfterColumns())
					ES.IndexDocument(ctx, header.GetTableName(), primaryKey, toMap)
				} else if eventType == pbe.EventType_TRUNCATE {
					ES.DeleteIndexDocuments(ctx, header.GetTableName())
				} else {
					//fmt.Println("-------> before")
					//printColumn(rowData.GetBeforeColumns())
					if log_open_val {
						fmt.Println("-------> after update")

						printColumn(rowData.GetAfterColumns())
					}

					toMap := columnsToMap(rowData.GetAfterColumns())
					ES.UpdateDocument(ctx, header.GetTableName(), primaryKey, toMap)
				}
			}
		}
	}
}

func printColumn(columns []*pbe.Column) {
	for _, col := range columns {
		fmt.Println(fmt.Sprintf("%s : %s  update= %t", col.GetName(), col.GetValue(), col.GetUpdated()))
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func columnsToMap(columns []*pbe.Column) map[string]interface{} {
	result := make(map[string]interface{})
	for _, column := range columns {
		result[column.GetName()] = column.GetValue()
	}
	return result
}
