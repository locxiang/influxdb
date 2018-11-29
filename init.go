package influxdb

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/lexkong/log"
	"time"
)

//默认数据库
var defaultDB string

//多数据库模式
var db map[string]*Inbound

func init() {
	db = make(map[string]*Inbound)
}

func SetDefaultDB(dbName string) error {
	var ok bool
	_, ok = db[dbName]
	if !ok {
		return fmt.Errorf("influxdb不存在：%s", dbName)
	}

	defaultDB = dbName
	return nil
}

//启动数据连接服务
func Init(addr, username, password, dbName, precision string, pushMax uint) {

	db[dbName] = &Inbound{
		Self:       openDB(addr, username, password),
		points:     make([]*client.Point, 0, pushMax),
		t:          time.NewTimer(TIMER),
		dbName:     dbName,
		precision:  precision,
		pushMax:    int(pushMax),
		close:      make(chan struct{}),
		closeOK:    make(chan struct{}),
		chanPoints: make(chan *client.Point),
	}

	go db[dbName].Run()

}

//关闭数据连接服务,dbNames为空关闭所有
func Stop(dbNames ...string) {
	if len(dbNames) == 0 {
		for _, inc := range db {
			inc.Stop()
		}
	} else {
		for _, dbName := range dbNames {
			inc, ok := db[dbName]
			if ok {
				inc.Stop()
			} else {
				log.Errorf(nil, "不存在数据库：%s", dbName)
			}
		}
	}
}

//init db
func openDB(addr, username, password string) client.Client {
	db, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s", addr),
		Username: username,
		Password: password,
		Timeout:  time.Duration(2 * time.Second),
	})
	if err != nil {
		panic(err)
		log.Error("influxdb new http client error: %s", err)
	}

	return db
}
