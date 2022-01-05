package metrics

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"log"
	"testing"
)

func TestMetricAggregator(t *testing.T) {
	defer goleak.VerifyNone(t)
	buf := bytes.Buffer{}
	logger := log.New(&buf, "", 0)
	inputChan := make(chan []*MetricSample)
	agg := NewMetricAggregator(logger, 2)
	agg.From(inputChan)

	assert.NoError(t, agg.Start())
	for _, sample := range aMatrixOfTaggedMetricSamples {
		inputChan <- sample
	}
	agg.Stop()
	assert.Equal(t, 4, len(agg.MetricsByInterval))
	// FIRST INTERVAL t = [1641316850 to 1641316852)
	assert.Equal(t, 2, len(agg.MetricsByInterval[0]))
	assert.Contains(t, agg.MetricsByInterval[0], "a_metric")
	assert.Contains(t, agg.MetricsByInterval[0], "b_metric")

	// SECOND INTERVAL t = [1641316852 to 1641316854)
	assert.Equal(t, 2, len(agg.MetricsByInterval[1]))
	assert.Contains(t, agg.MetricsByInterval[1], "a_metric")
	assert.Contains(t, agg.MetricsByInterval[1], "b_metric")

	// THIRD INTERVAL t = [1641316854 to 1641316856)
	assert.Equal(t, 2, len(agg.MetricsByInterval[2]))
	assert.Contains(t, agg.MetricsByInterval[2], "a_metric")
	assert.Contains(t, agg.MetricsByInterval[2], "c_metric")

	// FOURTH INTERVAL t = [1641316856 to now)
	assert.Equal(t, 1, len(agg.MetricsByInterval[3]))
	assert.Contains(t, agg.MetricsByInterval[3], "c_metric")
}
