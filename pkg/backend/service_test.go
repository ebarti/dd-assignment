package backend

import (
	"bytes"
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	defer goleak.VerifyNone(t)
	lines := []string{
		"\"10.0.0.2\",\"-\",\"apache\",1549573860,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.4\",\"-\",\"apache\",1549573861,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.4\",\"-\",\"apache\",1549573861,\"GET /api/user HTTP/1.0\",200,1234", // trigger
		"\"10.0.0.2\",\"-\",\"apache\",1549573862,\"GET /test/help HTTP/1.0\",200,1234",
		"\"10.0.0.5\",\"-\",\"apache\",1549573863,\"GET /test/help HTTP/1.0\",200,1234",
		"\"10.0.0.5\",\"-\",\"apache\",1549573865,\"GET /api/help HTTP/1.0\",200,1234", // recover
		"\"10.0.0.5\",\"-\",\"apache\",1549573866,\"GET /test/help HTTP/1.0\",200,1234",
		"\"10.0.0.5\",\"-\",\"apache\",1549573866,\"GET /api/help HTTP/1.0\",200,1234", // retrigger
		"\"10.0.0.5\",\"-\",\"apache\",1549573868,\"GET /api/help HTTP/1.0\",200,1234",
		"\"10.0.0.5\",\"-\",\"apache\",1549573869,\"GET /api/help HTTP/1.0\",200,1234", // recover
	}

	expectedOutputLines := []string{
		"High traffic generated an alert - hits 3, triggered at 101",
		"Recovered from high traffic at time 105",
		"High traffic generated an alert - hits 3, triggered at 106",
		"Recovered from high traffic at time 109",
	}
	_ = expectedOutputLines

	buf := bytes.Buffer{}

	var msgs []*common.Message
	for _, line := range lines {
		msgs = append(msgs, &common.Message{
			Content:            []byte(line),
			Origin:             "test",
			IngestionTimestamp: time.Now().Unix(),
		})
	}
	inputChan := make(chan *common.Message)
	interval := int64(2)
	threshold := int64(2)

	service := NewService(interval, GetCsvLogProcessingFunc(), []*metrics.CustomMetricPipeline{GetCsvCustomMetricsPipelines(interval)}, []*monitors.LogMonitorConfig{GetCsvLogMonitorConfig(interval, threshold)}, &buf)
	service.From(inputChan)
	assert.NoError(t, service.Start())
	for _, msg := range msgs {
		inputChan <- msg
	}
	service.Stop()
	fmt.Println(buf.String())
}

var logsForMonitorTest = []*logs.ProcessedLog{
	{
		Timestamp: 100,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 101,
		Status:    "201",
		Host:      "bHost",
	},
	{
		Timestamp: 101, // trigger alert
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 102,
		Status:    "301",
		Host:      "cHost",
	},
	{
		Timestamp: 103,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 105, // recover
		Status:    "401",
		Host:      "bHost",
	},
	{
		Timestamp: 106,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 106, // retrigger
		Status:    "404",
		Host:      "aHost",
	},
	{
		Timestamp: 108,
		Status:    "204",
		Host:      "aHost",
	},
	{
		Timestamp: 109, // recover
		Status:    "204",
		Host:      "aHost",
	},
}
