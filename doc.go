/**
influxdb 连接器，

1. 支持多实例，多数据库连接
2. 支持积累一定的Point再发送，或者超过1秒发送
3. 服务结束的时候，数据积压丢失
 */
package influxdb
