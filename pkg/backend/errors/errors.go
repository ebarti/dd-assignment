package errors

import (
	"fmt"
)

type CouldNotComputeMetricForTagError struct {
	tagName  string
	tagValue string
}

func NewCouldNotComputeMetricForTagError(tagName string, tagValue string) CouldNotComputeMetricForTagError {
	return CouldNotComputeMetricForTagError{tagName: tagName, tagValue: tagValue}
}

func (e CouldNotComputeMetricForTagError) Error() string {
	return "could not find metric value for tag: " + e.tagName + " with value: " + e.tagValue
}

type UnsampledMetricError struct{}

func NewUnsampledMetricError() UnsampledMetricError {
	return UnsampledMetricError{}
}

func (e UnsampledMetricError) Error() string {
	return "unsampled metric"
}

type InvalidAggregationQueryError struct {
	query string
}

func NewInvalidAggregationQueryError(query string) InvalidAggregationQueryError {
	return InvalidAggregationQueryError{query: query}
}

func (e InvalidAggregationQueryError) Error() string {
	return fmt.Sprintf("invalid aggregation query: %s", e.query)
}

type InvalidCsvLogFormatError struct {
	receivedFields int
	expectedFields int
}

func NewInvalidCsvLogFormatError(receivedFields int, expectedFields int) InvalidCsvLogFormatError {
	return InvalidCsvLogFormatError{receivedFields: receivedFields, expectedFields: expectedFields}
}

func (e InvalidCsvLogFormatError) Error() string {
	return fmt.Sprintf("invalid csv log format, received %d fields, expected %d fields", e.receivedFields, e.expectedFields)
}

type UnableToParseDateError struct {
	date  string
	error error
}

func NewUnableToParseDateError(date string, error error) UnableToParseDateError {
	return UnableToParseDateError{date: date, error: error}
}

func (e UnableToParseDateError) Error() string {
	return fmt.Sprintf("unable to parse date %s. Error: %v", e.date, e.error)
}

type InvalidRequestFormatError struct {
	request string
}

func (e InvalidRequestFormatError) Error() string {
	return fmt.Sprintf("invalid request format: %s", e.request)
}

func NewInvalidRequestFormatError(request string) InvalidRequestFormatError {
	return InvalidRequestFormatError{request: request}
}
