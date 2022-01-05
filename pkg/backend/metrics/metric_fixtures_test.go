package metrics

var host1Tag = &Tag{Name: "host", Value: "host1"}
var host2Tag = &Tag{Name: "host", Value: "host2"}
var section1Tag = &Tag{Name: "http.path.section", Value: "section1"}
var section2Tag = &Tag{Name: "http.path.section", Value: "section2"}

var aFlushTime = int64(105)
var aMetricSamples = []*MetricSample{
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 100,
	},
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 101,
	},
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 102,
	},
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 103,
	},
}

var aMetricSamplesComputed = &ComputedMetric{
	Timestamp: aFlushTime,
	Value:     4,
}

var aTaggedMetricSamples = []*MetricSample{
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 100,
		Tags:      []*Tag{host1Tag},
	},
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 101,
		Tags:      []*Tag{host1Tag},
	},
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 102,
		Tags:      []*Tag{host1Tag, section1Tag},
	},
	{
		Name:      "a_metric",
		Value:     1,
		Timestamp: 103,
		Tags:      []*Tag{host2Tag, section2Tag},
	},
}

var aTaggedMetricSamplesComputed = &ComputedMetric{
	Timestamp: aFlushTime,
	Value:     4,
	Groups: []*ComputedMetric{
		{
			Name:      "host",
			Timestamp: aFlushTime,
			Value:     4,
			Groups: []*ComputedMetric{
				{
					Name:      "host1",
					Timestamp: aFlushTime,
					Value:     3,
					Groups:    nil,
				},
				{
					Name:      "host2",
					Timestamp: aFlushTime,
					Value:     1,
					Groups:    nil,
				},
			},
		},
		{
			Name:      "http.path.section",
			Timestamp: aFlushTime,
			Value:     2,
			Groups: []*ComputedMetric{
				{
					Name:      "section1",
					Timestamp: aFlushTime,
					Value:     1,
					Groups:    nil,
				},
				{
					Name:      "section2",
					Timestamp: aFlushTime,
					Value:     1,
					Groups:    nil,
				},
			},
		},
	},
}

var aMatrixOfTaggedMetricSamples = [][]*MetricSample{
	{
		{
			Name:      "a_metric",
			Value:     1,
			Timestamp: 1641316850,
			Tags:      []*Tag{host1Tag},
		},
		{
			Name:      "b_metric",
			Value:     1,
			Timestamp: 1641316850,
			Tags:      []*Tag{section1Tag},
		},
	},
	{
		{
			Name:      "a_metric",
			Value:     1,
			Timestamp: 1641316851,
			Tags:      []*Tag{host1Tag},
		},
		{
			Name:      "b_metric",
			Value:     1,
			Timestamp: 1641316851,
			Tags:      []*Tag{section2Tag},
		},
	},
	{
		{
			Name:      "a_metric",
			Value:     1,
			Timestamp: 1641316852,
			Tags:      []*Tag{host2Tag},
		},
		{
			Name:      "b_metric",
			Value:     1,
			Timestamp: 1641316852,
			Tags:      []*Tag{section1Tag},
		},
	},
	{
		{
			Name:      "a_metric",
			Value:     1,
			Timestamp: 1641316853,
			Tags:      []*Tag{section2Tag},
		},
		{
			Name:      "b_metric",
			Value:     1,
			Timestamp: 1641316853,
		},
	},
	{
		{
			Name:      "a_metric",
			Value:     1,
			Timestamp: 1641316854,
		},
	},
	{
		{
			Name:      "a_metric",
			Value:     1,
			Timestamp: 1641316853, // logs often have lines that are out of order
			Tags:      []*Tag{section2Tag},
		},
		{
			Name:      "b_metric",
			Value:     1,
			Timestamp: 1641316853, // logs often have lines that are out of order
		},
	},
	{
		{
			Name:      "c_metric",
			Value:     1,
			Timestamp: 1641316855,
			Tags:      []*Tag{section2Tag},
		},
	},
	{
		{
			Name:      "c_metric",
			Value:     1,
			Timestamp: 1641316856,
			Tags:      []*Tag{host1Tag},
		},
	},
}
