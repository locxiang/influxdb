# influxdb
> influxdb时序数据库高性能写入服务


## 功能说明

1. 支持多实例，多数据库连接
2. 支持积累一定的Point再发送，或者超过1秒发送
3. 服务结束的时候，数据积压丢失

## 安装

```
go get -v github.com/locxiang/influxdb
```

## 使用样例

```
influxdb.Init()         // 初始化数据库连接
influxdb.SetDefaultDB() // 设置默认数据库，方便单例模式的用户

influxdb.AddPoint()     // 往指定数据库写入数据
influxdb.DefaultAddPoint()  //往默认数据库写入数据

```