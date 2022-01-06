package logs

var aProcessedLog = &ProcessedLog{
	Timestamp: 123456789,
	Status:    "200",
	Host:      "aHost",
	Service:   "aService",
	Message:   "aMessage",
	Attributes: map[string]interface{}{
		"aMeasurableAttribute": "1",
		"nested": map[string]interface{}{
			"aMeasurableAttribute": "2",
			"nested": map[string]interface{}{
				"aMeasurableAttribute": "3",
				"nested": map[string]interface{}{
					"aMeasurableAttribute": "4",
				},
			},
		},
	},
}
