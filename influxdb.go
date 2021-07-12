package influxdb

import (
	"crypto/tls"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

var writeAPI api.WriteAPI
var client influxdb2.Client

func Init(url, token, org, bucket string) error {
	// Create HTTP client
	httpClient := &http.Client{
		Timeout: time.Second * time.Duration(60),
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	// Create a new client using an InfluxDB server base URL and an authentication token
	client = influxdb2.NewClientWithOptions(url, token, influxdb2.DefaultOptions().SetBatchSize(1000).SetHTTPClient(httpClient))
	// Use blocking write client for writes to desired bucket
	writeAPI = client.WriteAPI(org, bucket)
	errorsCh := writeAPI.Errors()
	// Create go proc for reading and logging errors
	go func() {
		for err := range errorsCh {
			log.Errorf("write error: %s", err.Error())
		}
	}()

	return nil
}

func Close() {
	// Force all unwritten data to be sent
	writeAPI.Flush()
	// Ensures background processes finishes
	client.Close()
}

func WriteJob(project, jType, jobid string, status string, elapsedTime float64, ctime time.Time) {
	p := influxdb2.NewPoint(
		"jobs",
		map[string]string{
			"project": project,
			"type":    jType,
			"status":  status,
		},
		map[string]interface{}{
			"jobid":   jobid,
			"elapsed": elapsedTime,
		},
		ctime)
	// write asynchronously
	writeAPI.WritePoint(p)
}

func Write(point *write.Point) {

	writeAPI.WritePoint(point)
}
