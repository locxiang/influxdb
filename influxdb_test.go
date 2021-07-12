package influxdb_test

import (
	"github.com/locxiang/influxdb"
	"testing"
	"time"
)

func init() {
	influxdb.Init("http://192.168.1.212:8086", "XibT7Wr4QFqLWsndWMoShlZhjQlbdvpWnCisqgwcc2-YnYY5b3C0PZY_t_5p5MNz7W2KmdGkQFDTPQak0RVy3g==", "ZKJL", "data-center")

}

func TestNewPointJob(t *testing.T) {
	influxdb.WriteJob("test", "微信附近人", "alksdjflkasdjf", "completed", 30, time.Now())

	time.Sleep(5 * time.Second)
}
