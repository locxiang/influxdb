package influxdb_test

import (
	"github.com/lexkong/log"
	"gitlab.singulato.com/neltharion/matrix_service/config"
	"gitlab.singulato.com/neltharion/matrix_service/influxdb"
	"testing"
	"time"
)

func init() {
	// init config
	if err := config.Init("../config.yaml"); err != nil {
		log.Errorf(err, "config init error")
		panic(err)
	}

	conf := config.Values.Db
	influxdb.Init(conf.Addr, conf.Name, conf.Password, conf.Name, conf.Push.Precision, conf.Push.Max)
}

func TestAddPoint(t *testing.T) {
	table := "test_5"

	tags := map[string]string{
		"test": "TestInit",
	}

	fields := map[string]interface{}{
		"unix": time.Now().Unix(),
	}

	err := influxdb.AddPoint(config.Values.Db.Name, table, tags, fields)
	if err != nil {
		t.Error(err)
	}
	influxdb.Stop()
}

func TestDefaultAddPoint(t *testing.T) {
	table := "test_defaultAddPoint"

	tags := map[string]string{
		"test": "TestInit",
	}

	fields := map[string]interface{}{
		"unix": time.Now().Unix(),
	}

	err := influxdb.DefaultAddPoint(table, tags, fields)
	if err != nil {
		influxdb.SetDefaultDB(config.Values.Db.Name)
	} else {
		t.Error("没设置默认，居然没错？")
	}

	if err := influxdb.DefaultAddPoint(table, tags, fields); err != nil {
		t.Error(err)
	}

	influxdb.Stop()

}

func BenchmarkAddPoint(b *testing.B) {
	table := "test_benchmark"

	tags := map[string]string{
		"test": "TestInit",
	}

	for i := 0; i < b.N; i++ {
		fields := map[string]interface{}{
			"unix": time.Now().Unix(),
		}
		influxdb.AddPoint(config.Values.Db.Name, table, tags, fields)
	}

}

func TestInbound_Stop(t *testing.T) {
	influxdb.Stop()
}
