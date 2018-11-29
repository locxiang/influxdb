package main

import (
	"github.com/locxiang/influxdb"
	"log"
	"time"
)

func main() {
	config := influxdb.Config{
		Addr:      "127.0.0.1:8086",
		Username:  "root",
		Password:  "root",
		DBName:    "dbname1",
		Precision: "ns",
		PushMax:   5000, //5000是性能最好的
		Error: func(err error) {
			log.Println(err)
		},
		Debug: false,
	}

	err := influxdb.Init(config)
	if err != nil {
		log.Fatal(err)
	}
	influxdb.SetDefaultDB(config.DBName)

	table, tags, fields := getData()
	//写入到默认数据
	influxdb.DefaultAddPoint(table, tags, fields)

	influxdb.Stop()

}

func getData() (table string, tags map[string]string, fields map[string]interface{}) {
	table = "test_default222"

	tags = map[string]string{
		"test": "example",
	}
	fields = map[string]interface{}{
		"unix": time.Now().Unix(),
	}
	return
}
