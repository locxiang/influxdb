package influxdb

import (
	"errors"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"time"
)

//  AddPoint 通过AddPoint函数来解耦，降低外界依赖
func AddPoint(dbName, tableName string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	inc, ok := db[dbName]
	if !ok {
		return fmt.Errorf("数据库不存在【%s】，请初始化", dbName)
	}

	pt, err := client.NewPoint(tableName, tags, fields, t...)
	if err != nil {
		return err
	}

	select {
	case <-time.After(100 * time.Millisecond):
		return errors.New("AddPoint 写入失败，超时")
	case inc.chanPoints <- pt:
	}

	return nil
}

//提交到默认数据库
func DefaultAddPoint(tableName string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	if defaultDB == "" {
		return errors.New("默认数据库不存在")
	}
	return AddPoint(defaultDB, tableName, tags, fields, t...)
}
