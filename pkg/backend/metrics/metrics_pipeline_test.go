package metrics

import (
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsPipeline(t *testing.T) {
	inputChan := make(chan *logs.ProcessedLog)
	outpuChan := make(chan []*MetricSample, len(processedLogsForTest)+1)

	// create a custom metric pipeline to get an overall request count grouped by host
	groupBy := []string{"host", "status"}
	customMetricPipeline := NewCustomMetricPipeline("test", "*", 10, nil, groupBy)
	metricsPipeline := NewMetricsPipeline([]*CustomMetricPipeline{customMetricPipeline})
	metricsPipeline.OutputChan = outpuChan
	metricsPipeline.From(inputChan)

	// start the pipeline
	assert.NoError(t, metricsPipeline.Start())

	// send the logs to the pipeline
	for _, log := range processedLogsForTest {
		inputChan <- log
	}
	metricsPipeline.Stop()

	// check the output
	ii := 0
	for output := range outpuChan {
		// since we only have a single custom metric, we expect to get one metric sample per output
		assert.Equal(t, 1, len(output))
		gotMetric := output[0]
		assert.Equal(t, "test", gotMetric.Name)
		assert.Equal(t, len(groupBy), len(gotMetric.Tags))
		for _, group := range groupBy {
			found := false
			gotValue := ""
			for _, tag := range gotMetric.Tags {
				if tag.Name == group {
					found = true
					gotValue = tag.Value
					break
				}
			}
			assert.True(t, found)
			wantValue := *processedLogsForTest[ii].GetAttribute(group)
			assert.Equalf(t, wantValue, gotValue, "expected %s to be %s", gotValue, wantValue)
		}
		ii++
	}

}

var processedLogsForTest = []*logs.ProcessedLog{
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
		Timestamp: 102,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 103,
		Status:    "301",
		Host:      "cHost",
	},
	{
		Timestamp: 104,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 104,
		Status:    "401",
		Host:      "bHost",
	},
	{
		Timestamp: 105,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 104,
		Status:    "404",
		Host:      "aHost",
	},
	{
		Timestamp: 106,
		Status:    "204",
		Host:      "aHost",
	},
}
