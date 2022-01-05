package backend

import (
	"bytes"
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/common"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	lines := []string{
		"\"10.0.0.2\",\"-\",\"apache\",1549573860,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.4\",\"-\",\"apache\",1549573860,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.4\",\"-\",\"apache\",1549573860,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.2\",\"-\",\"apache\",1549573860,\"GET /api/help HTTP/1.0\",200,1234",
		"\"10.0.0.5\",\"-\",\"apache\",1549573860,\"GET /api/help HTTP/1.0\",200,1234",
	}

	buf := bytes.Buffer{}

	var msgs []*common.Message
	for _, line := range lines {
		msgs = append(msgs, &common.Message{
			Content:            []byte(line),
			Origin:             "test",
			IngestionTimestamp: time.Now().Unix(),
		})
	}
	interval := int64(2)
	threshold := int64(2)

	service := NewService(interval, GetCsvLogProcessingFunc(), []*metrics.CustomMetricPipeline{GetCsvCustomMetricsPipelines(interval)}, []*monitors.LogMonitorConfig{GetCsvLogMonitorConfig(interval, threshold)}, &buf)
	_ = service
}
