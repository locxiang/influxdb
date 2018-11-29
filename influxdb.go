package influxdb

import (
	"errors"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"time"
)

//TIMER 发送频率，如果规定时间还没达到Max数量就发送
const TIMER = 1000 * time.Millisecond

type Inbound struct {
	Self       client.Client
	points     []*client.Point
	t          *time.Timer
	dbName     string
	precision  string
	pushMax    int
	close      chan struct{}
	closeOK    chan struct{} //用来确认关闭完成
	chanPoints chan *client.Point
	errorFun   func(err error)
	debug      bool
}

func (inbound *Inbound) Log(format string, args ...interface{}) {
	if inbound.debug {
		fmt.Printf(format+"\r\n", args...)
	}
}

func (inbound *Inbound) run() {
	for {
		select {
		case <-inbound.close:
			inbound.Log("influxdb[%s]关闭前最后一次发送数据", inbound.dbName)
			inbound.lastSendDB()
			inbound.Log("关闭influxdb:%s", inbound.dbName)
			inbound.closeOK <- struct{}{}
			return
		case <-inbound.t.C:
			if len(inbound.points) > 0 {
				inbound.Log("计时器执行 添加%d条", len(inbound.points))
				err := inbound.sendDb()
				if err != nil {
					inbound.errorFun(err)
				}
			} else {
				inbound.t.Reset(TIMER)
			}
		case msg := <-inbound.chanPoints:
			inbound.points = append(inbound.points, msg)
			if len(inbound.points) >= int(inbound.pushMax) {
				err := inbound.sendDb()
				if err != nil {
					inbound.errorFun(err)
				}
			}
		}
	}
}

func (inbound *Inbound) newBatchPoints() (client.BatchPoints, error) {
	ballpoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  inbound.dbName,
		Precision: inbound.precision,
	})

	return ballpoints, err
}

//服务结束前一次发送数据
func (inbound *Inbound) lastSendDB() {

	ballpoints, err := inbound.newBatchPoints()
	if err != nil {
		inbound.errorFun(err)
		return
	}

	if len(inbound.points) == 0 {
		inbound.Log("消息池内容为空不发送")
		return
	}

	ballpoints.AddPoints(inbound.points)
	err = inbound.Self.Write(ballpoints)
	if err != nil {
		inbound.errorFun(err)
	}
	return
}

func (inbound *Inbound) sendDb() error {
	defer func() {
		inbound.t.Reset(TIMER)
	}()

	ballpoints, err := inbound.newBatchPoints()
	if err != nil {
		return err
	}

	c := make([]*client.Point, len(inbound.points))
	copy(c, inbound.points)
	inbound.points = make([]*client.Point, 0, inbound.pushMax)

	go func() {
		ballpoints.AddPoints(c)
		err = inbound.Self.Write(ballpoints)
		if err != nil {
			inbound.errorFun(err)
		}
	}()

	return nil
}

func (inbound *Inbound) stop() {
	go func() {
		inbound.close <- struct{}{}
	}()
	<-inbound.closeOK
	inbound.Self.Close()
}

//判断db是否存在
func (inbound *Inbound) existsDB() error {

	q := client.NewQuery("SHOW DATABASES", "", "")
	if response, err := inbound.Self.Query(q); err != nil {
		return err
	} else {
		if response.Error() != nil {
			return response.Error()
		}

		databases := response.Results[0].Series[0].Values
		for _, d := range databases {
			dbName, _ := d[0].(string)
			if dbName == inbound.dbName {
				return nil
			}
		}
		return errors.New("数据库不存在：" + inbound.dbName)
	}
}
