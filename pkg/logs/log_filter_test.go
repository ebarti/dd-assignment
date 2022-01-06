package logs

import (
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogFilter_Matches(t *testing.T) {
	type fields struct {
		matchAll   bool
		status     string
		host       string
		service    string
		attributes map[string]string
	}
	tests := []struct {
		fields fields
		want   bool
	}{
		// single attribute filters
		{
			fields: fields{
				status: "200",
			},
			want: true,
		},
		{
			fields: fields{
				host: "aHost",
			},
			want: true,
		},
		{
			fields: fields{
				service: "aService",
			},
			want: true,
		},
		{
			fields: fields{
				attributes: map[string]string{
					"aMeasurableAttribute": "1",
				},
			},
			want: true,
		},
		// combinations
		{
			fields: fields{
				status: "200",
				host:   "aHost",
			},
			want: true,
		},
		{
			fields: fields{
				attributes: map[string]string{
					"aMeasurableAttribute":        "1",
					"nested.aMeasurableAttribute": "2",
				},
			},
			want: true,
		},
		{
			fields: fields{
				service: "aService",
				attributes: map[string]string{
					"aMeasurableAttribute":        "1",
					"nested.aMeasurableAttribute": "2",
				},
			},
			want: true,
		},
		// Failures
		{
			fields: fields{
				status: "1",
			},
			want: false,
		},
		{
			fields: fields{
				status: "200",
				host:   "bHost",
			},
			want: false,
		},
		{
			fields: fields{
				service: "aService",
				attributes: map[string]string{
					"aMeasurableAttribute":        "1",
					"nested.aMeasurableAttribute": "0",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		testName := "test LogFilter matches"
		a := &LogFilter{}
		if tt.fields.matchAll {
			a.matchAll = true
			testName += "  all"
		} else {
			if tt.fields.status != "" {
				a.status = &tt.fields.status
				testName += " status"
			}
			if tt.fields.host != "" {
				a.host = &tt.fields.host
				testName += " host"
			}
			if tt.fields.service != "" {
				a.service = &tt.fields.service
				testName += " service"
			}
			if tt.fields.attributes != nil && len(tt.fields.attributes) > 0 {
				a.attributes = tt.fields.attributes
				testName += " attributes"
			}
		}
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, tt.want, a.Matches(aProcessedLog))
		})
	}
}

func TestLogFilter_build(t *testing.T) {
	tests := []struct {
		query   string
		wantErr bool
	}{
		{
			query:   "*",
			wantErr: false,
		},
		{
			query:   "status:200",
			wantErr: false,
		},
		{
			query:   "service:api",
			wantErr: false,
		},
		{
			query:   "host:meow",
			wantErr: false,
		},
		{
			query:   "@http.path.section:a",
			wantErr: false,
		},
		{
			query:   "http.path.section",
			wantErr: true,
		},
		{
			query:   "@http.path.section",
			wantErr: true,
		},
		{
			query:   "@http.path.section:is:invalid",
			wantErr: true,
		},
		{
			query:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("test log filter build with query %s", tt.query), func(t *testing.T) {
			a := &LogFilter{}
			err := a.build(tt.query)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, errors.NewInvalidAggregationQueryError(tt.query))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
