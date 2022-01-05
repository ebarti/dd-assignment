package logs

import (
	"strconv"
)

// LogMeasure represents a measure that can be extracted from a pipeline.ProcessedLog
type LogMeasure struct {
	name string
}

// NewLogMeasure creates a new LogMeasure
// param measure is a string representing any attribute in pipeline.ProcessedLog.
// If it is an attribute within pipeline.ProcessedLog.Attributes, just use the
// attribute path as the measure name. For example: "http.path.section"
func NewLogMeasure(measure *string) *LogMeasure {
	if measure == nil {
		return nil
	}
	return &LogMeasure{name: *measure}
}

// Measure resolves the dimension for the given log. It returns an int pointer as the
// dimension might not exist for the given log
func (d *LogMeasure) Measure(log *ProcessedLog) *int64 {
	val := log.GetAttribute(d.name)
	if val == nil {
		return nil
	}
	parsedVal, err := strconv.ParseInt(*val, 10, 64)
	if err != nil {
		return nil
	}

	return &parsedVal
}

// GetName returns the name of the measure
func (d *LogMeasure) GetName() string {
	return d.name
}
