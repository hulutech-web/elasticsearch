package feature

import (
	"fmt"
	"github.com/goravel/framework/facades"
	"github.com/withlin/canal-go/client"
	pbe "github.com/withlin/canal-go/protocol/entry"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"slices"
	"testing"
	"time"
)

// StartCanalSync 启动Canal同步到ES
func TestStartCanalSync(t *testing.T) {
	//facades.Config().Env("CANAL")
	//config := facades.Config()
	//mode := config.GetBool("elasticsearch.canal")
	//if !mode {
	//	os.Exit(1)
	//}
	connector := client.NewSimpleCanalConnector("127.0.0.1", 11111, "", "", "example", 60000, 60*60*1000)
	err := connector.Connect()
	//defer connector.DisConnection()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	tables := facades.Config().Get("elasticsearch.tables")
	// 单类型断言
	if v, ok := tables.([]string); ok {
		// 成功断言为 []string
		fmt.Println("Type assertion succeeded:", v)
	} else {
		// 断言失败
		fmt.Println("Type assertion failed")
	}
	schema := facades.Config().GetString("elasticsearch.schema")
	subscribeStr := ""
	for _, v := range tables.([]string) {
		subscribeStr += schema + "\\." + v + ","
	}
	//去除最后一个逗号
	subscribeStr = subscribeStr[:len(subscribeStr)-1]
	err = connector.Subscribe(subscribeStr)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("===start sync es data===")
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
			printEntry(message.Entries)
			syncEs(message.Entries, schema, tables.([]string))
		}
	}
}
func syncEs(entrys []pbe.Entry, schema string, tables []string) error {
	//进行elasticsearch的同步操作
	for _, entry := range entrys {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}
		rowChange := new(pbe.RowChange)

		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		checkError(err)
		if rowChange != nil {
			eventType := rowChange.GetEventType()
			header := entry.GetHeader()
			if header.GetSchemaName() == schema && slices.Contains(tables, header.GetTableName()) {
				if eventType == pbe.EventType_UPDATE {
					fmt.Println("更新数据")
				} else if eventType == pbe.EventType_INSERT {
					fmt.Println("插入数据")
				} else if eventType == pbe.EventType_DELETE {
					fmt.Println("删除数据")
				}
			}
		}
	}
	return nil
}

func printEntry(entrys []pbe.Entry) {

	for _, entry := range entrys {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}
		rowChange := new(pbe.RowChange)

		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		checkError(err)
		if rowChange != nil {
			eventType := rowChange.GetEventType()
			header := entry.GetHeader()
			fmt.Println(fmt.Sprintf("================> binlog[%s : %d],name[%s,%s], eventType: %s", header.GetLogfileName(), header.GetLogfileOffset(), header.GetSchemaName(), header.GetTableName(), header.GetEventType()))

			for _, rowData := range rowChange.GetRowDatas() {
				if eventType == pbe.EventType_DELETE {
					printColumn(rowData.GetBeforeColumns())
				} else if eventType == pbe.EventType_INSERT {
					printColumn(rowData.GetAfterColumns())
				} else {
					fmt.Println("-------> before")
					printColumn(rowData.GetBeforeColumns())
					fmt.Println("-------> after")
					printColumn(rowData.GetAfterColumns())

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
