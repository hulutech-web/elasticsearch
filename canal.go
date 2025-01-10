package elasticsearch

import (
	"context"
	"fmt"
	gcolor "github.com/gookit/color"
	"github.com/goravel/framework/facades"
	"github.com/sirupsen/logrus"
	"github.com/withlin/canal-go/client"
	pbe "github.com/withlin/canal-go/protocol/entry"
	"google.golang.org/protobuf/proto"
	"os"
	"time"
)

// StartCanalSync 启动Canal同步到ES
func StartCanalSync() error {

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	mode := facades.Config().GetBool("elasticsearch.canal")
	if !mode {
		return nil
	}
	ctx := context.Background()
	NewElastic(ctx)
	//参考canal.properties配置文件中的内容，修改成自己的配置
	connector := client.NewSimpleCanalConnector("localhost", 11111, "", "", "example", 60000, 60*60*1000)
	err := connector.Connect()
	if err != nil {
		logrus.Errorln(fmt.Sprintf("CanalConnector connect fail: %v", err.Error()))
		//color.Red().Println(fmt.Sprintf("CanalConnector init fail: %v", err.Error()))
		return err
	}
	defer connector.DisConnection()
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
		facades.Log().Error(err.Error())
		return err
	}
	gcolor.Success.Tips(fmt.Sprintf("Canal Subscribe Success, tables: %v\n", subscribeStr))
	for {

		message, err := connector.Get(100, nil, nil)
		if err != nil {
			facades.Log().Error(err.Error())
			return err
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
			logrus.Info(fmt.Sprintf("【ES INFO】 binlog[%s : %d],Schema:[%s],tablename:[%s],docID:[%s] eventType: %s", header.GetLogfileName(), header.GetLogfileOffset(), header.GetSchemaName(), header.GetTableName(), primaryKey, header.GetEventType()))
			//提取rowChange.GetRowDatas()中的数据，转换为一个map结构，进行es同步
			for _, rowData := range rowChange.GetRowDatas() {
				if log_open_val {
					fmt.Println("{")
					printColumn(rowData.GetAfterColumns())
					fmt.Println("}")
				}
				if eventType == pbe.EventType_DELETE {
					ES.DeleteDocument(ctx, header.GetTableName(), primaryKey)
				} else if eventType == pbe.EventType_INSERT {
					toMap := columnsToMap(rowData.GetAfterColumns())
					ES.IndexDocument(ctx, header.GetTableName(), primaryKey, toMap)
				} else if eventType == pbe.EventType_TRUNCATE {
					ES.DeleteIndexDocuments(ctx, header.GetTableName())
				} else {
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
