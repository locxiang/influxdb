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
var db = make(map[string]*Inbound)

// Config 配置文件
type Config struct {
	Addr,
	Username,
	Password,
	DBName,
	Precision string
	PushMax int
	Error   func(err error)
	Debug   bool
}

// SetDefaultDB 设置默认数据库
func SetDefaultDB(dbName string) error {
	var ok bool
	_, ok = db[dbName]
	if !ok {
		return fmt.Errorf("influxdb不存在：%s", dbName)
	}

	defaultDB = dbName
	return nil
}

//Init 启动数据连接服务  返回一个退出状态
func Init(c Config) error {
	self := openDB(c.Addr, c.Username, c.Password)

	db[c.DBName] = &Inbound{
		Self:       self,
		points:     make([]*client.Point, 0, c.PushMax),
		t:          time.NewTimer(TIMER),
		dbName:     c.DBName,
		precision:  c.Precision,
		pushMax:    int(c.PushMax),
		close:      make(chan struct{}),
		closeOK:    make(chan struct{}),
		chanPoints: make(chan *client.Point),
		errorFun:   c.Error,
		debug:      c.Debug,
	}

	err := db[c.DBName].existsDB()
	if err != nil {
		return err
	}

	go db[c.DBName].run()

	return nil

}

//Stop 关闭数据连接服务,dbNames为空关闭所有
func Stop(dbNames ...string) error {
	if len(dbNames) == 0 {
		for dbName, inc := range db {
			inc.stop()
			delete(db, dbName)
		}
	} else {
		for _, dbName := range dbNames {
			inc, ok := db[dbName]
			if ok {
				inc.stop()
				delete(db, dbName)
			} else {
				return fmt.Errorf("不存在数据库：%s", dbName)
			}
		}
	}

	return nil
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
