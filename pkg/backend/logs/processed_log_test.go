package logs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessedLog_GetAttribute(t *testing.T) {
	tests := []struct {
		getAttributeName string
		want             string
		exists           bool
	}{
		{
			getAttributeName: "timestamp",
			want:             "123456789",
			exists:           true,
		},
		{
			getAttributeName: "status",
			want:             "200",
			exists:           true,
		},
		{
			getAttributeName: "host",
			want:             "aHost",
			exists:           true,
		},
		{
			getAttributeName: "service",
			want:             "aService",
			exists:           true,
		},
		{
			getAttributeName: "message",
			want:             "aMessage",
			exists:           true,
		},
		{
			getAttributeName: "aMeasurableAttribute",
			want:             "1",
			exists:           true,
		},
		{
			getAttributeName: "nested.aMeasurableAttribute",
			want:             "2",
			exists:           true,
		},
		{
			getAttributeName: "nested.nested.aMeasurableAttribute",
			want:             "3",
			exists:           true,
		},
		{
			getAttributeName: "nested.nested.nested.aMeasurableAttribute",
			want:             "4",
			exists:           true,
		},
		// Non-existent attributes:
		{
			getAttributeName: "aNonExistentAttribute",
			exists:           false,
		},
		{
			getAttributeName: "nested.aNonExistentAttribute",
			exists:           false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("get %s attribute", tt.getAttributeName), func(t *testing.T) {
			got := aProcessedLog.GetAttribute(tt.getAttributeName)
			if tt.exists {
				assert.Equal(t, tt.want, *got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestProcessedLog_HasAttributeWithValue(t *testing.T) {
	tests := []struct {
		getAttributeName  string
		getAttributeValue string
		want              bool
	}{
		{
			getAttributeName:  "timestamp",
			getAttributeValue: "123456789",
			want:              true,
		},
		{
			getAttributeName:  "timestamp",
			getAttributeValue: "23456789",
			want:              false,
		},
		{
			getAttributeName:  "status",
			getAttributeValue: "200",
			want:              true,
		},
		{
			getAttributeName:  "status",
			getAttributeValue: "201",
			want:              false,
		},
		{
			getAttributeName:  "host",
			getAttributeValue: "aHost",
			want:              true,
		},
		{
			getAttributeName:  "host",
			getAttributeValue: "bHost",
			want:              false,
		},
		{
			getAttributeName:  "service",
			getAttributeValue: "aService",
			want:              true,
		},
		{
			getAttributeName:  "service",
			getAttributeValue: "cService",
			want:              false,
		},
		{
			getAttributeName:  "message",
			getAttributeValue: "aMessage",
			want:              true,
		},
		{
			getAttributeName:  "message",
			getAttributeValue: "noMessage",
			want:              false,
		},
		{
			getAttributeName:  "aMeasurableAttribute",
			getAttributeValue: "1",
			want:              true,
		},
		{
			getAttributeName:  "aMeasurableAttribute",
			getAttributeValue: "-1",
			want:              false,
		},
		{
			getAttributeName:  "nested.nested.nested.aMeasurableAttribute",
			getAttributeValue: "4",
			want:              true,
		},
		{
			getAttributeName:  "nested.nested.nested.aMeasurableAttribute",
			getAttributeValue: "3",
			want:              false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("test %s attribute exists with value %s", tt.getAttributeName, tt.getAttributeValue), func(t *testing.T) {
			assert.Equal(t, tt.want, aProcessedLog.HasAttributeWithValue(tt.getAttributeName, tt.getAttributeValue))
		})
	}
}
