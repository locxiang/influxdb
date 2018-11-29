package influxdb

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/lexkong/log"
	"math/rand"
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
}

func (inbound *Inbound) Run() {
	log.Infof("启动inbound连接：%s", inbound.dbName)
	defer log.Debugf("结束inbound循环：%s", inbound.dbName)

	for {
		select {
		case <-inbound.close:
			log.Debugf("influxdb[%s]关闭前最后一次发送数据", inbound.dbName)
			inbound.lastSendDB()
			log.Infof("关闭influxdb:%s", inbound.dbName)
			inbound.closeOK <- struct{}{}
			return
		case <-inbound.t.C:
			if len(inbound.points) > 0 {
				log.Debugf("计时器执行 %d", len(inbound.points))
				inbound.sendDb()
			} else {
				inbound.t.Reset(TIMER)
			}
		case msg, ok := <-inbound.chanPoints:
			if !ok {
				log.Errorf(nil, "[%s] chanPoints通道被意外关闭", inbound.dbName)
				return
			}

			inbound.points = append(inbound.points, msg)
			if len(inbound.points) >= int(inbound.pushMax) {
				log.Debugf("超过 [%d] 条执行插入数据库操作", len(inbound.points))
				inbound.sendDb()
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
		log.Errorf(err, "new batch chanPoints error")
		return
	}

	if len(inbound.points) == 0 {
		log.Debugf("消息池内容为空不发送")
		return
	}

	ballpoints.AddPoints(inbound.points)
	err = inbound.Self.Write(ballpoints)
	if err != nil {
		log.Errorf(err, "inbound[%s] Write error", inbound.dbName)
	}
	return
}

func (inbound *Inbound) sendDb() error {
	defer func() {
		inbound.t.Reset(TIMER)
	}()

	ballpoints, err := inbound.newBatchPoints()
	if err != nil {
		log.Errorf(err, "new batch chanPoints error")
		return err
	}

	c := make([]*client.Point, len(inbound.points))
	copy(c, inbound.points)
	inbound.points = make([]*client.Point, 0, inbound.pushMax)

	go func() {
		ballpoints.AddPoints(c)
		err = inbound.Self.Write(ballpoints)
		if err != nil {
			time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
			if e := inbound.Self.Write(ballpoints); e != nil {
				log.Errorf(e, "write db error")
			}
		}
	}()

	return nil
}

func (inbound *Inbound) Stop() {
	go func() {
		inbound.close <- struct{}{}
	}()
	<-inbound.closeOK
}
