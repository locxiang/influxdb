package influxdb

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
)

func DB(dbName string) (client.Client, error) {
	inc, ok := db[dbName]
	if !ok {
		return nil, fmt.Errorf("数据库不存在【%s】，请初始化", dbName)
	}
	return inc.Self, nil
}
