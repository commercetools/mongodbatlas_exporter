package transformer

import (
	"math"
	m "mongodbatlas_exporter/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas/mongodbatlas"
)

func TestValueTransformer_validationDataPoint(t *testing.T) {
	assert := assert.New(t)

	testCasesIsValid := map[bool][][]*mongodbatlas.DataPoints{
		true: {{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
		}},
		false: {
			{
				{
					Timestamp: "2021-20-04T16:53:06Z",
					Value:     nil,
				},
			},
			{
				{
					Timestamp: "",
					Value:     nil,
				},
			},
			{},
		},
	}
	for isValid, testCases := range testCasesIsValid {
		for _, testCase := range testCases {
			err := containsValidDataPoints(testCase)
			assert.Equal(isValid, err == nil)
		}
	}
}

func TestValueTransformer_sortDataPoints(t *testing.T) {
	assert := assert.New(t)
	valueSmaller := float32(2.72)
	valueBigger := float32(3.14)
	testCases := []map[string][]*mongodbatlas.DataPoints{
		{
			"given": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:53:06Z",
					Value:     &valueSmaller,
				},
				{
					Timestamp: "2021-03-04T16:54:06Z",
					Value:     nil,
				},
			},
			"expected": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:53:06Z",
					Value:     &valueSmaller,
				},
				{
					Timestamp: "2021-03-04T16:54:06Z",
					Value:     nil,
				},
			},
		}, // t1 < t2; float,nil => no change
		{
			"given": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
				{
					Timestamp: "2021-03-04T16:54:06Z",
					Value:     nil,
				},
			},
			"expected": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:54:06Z",
					Value:     nil,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
			},
		}, // t1 > t2; float,nil => t2,t1
		{
			"given": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueBigger,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
			},
			"expected": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueBigger,
				},
			},
		}, // t1 == t2; float1>float2 => t2,t1
		{
			"given": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     nil,
				},
			},
			"expected": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     nil,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
			},
		}, // t1 == t2; float, nil => t2, t1
		{
			"given": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     nil,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
			},
			"expected": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     nil,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
			},
		}, // t1 == t2; nil,float => no change
		{
			"given": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:05Z",
					Value:     &valueBigger,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
			},
			"expected": []*mongodbatlas.DataPoints{
				{
					Timestamp: "2021-03-04T16:55:05Z",
					Value:     &valueBigger,
				},
				{
					Timestamp: "2021-03-04T16:55:06Z",
					Value:     &valueSmaller,
				},
			},
		}, // t1 < t2; float1>float2 => no change
	}

	for _, testCase := range testCases {
		given := testCase["given"]
		expected := testCase["expected"]

		sortDataPoints(&given)

		assert.Equal(expected, given)
	}
}

func TestValueTransformer_twoDatapoints_allNotValid(t *testing.T) {
	assert := assert.New(t)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     nil,
			},
		},
		Units: m.SCALAR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(0), promValue)
}

func TestValueTransformer_twoDatapoints_latestValid(t *testing.T) {
	assert := assert.New(t)
	value2 := float32(2.10016)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value2,
			},
		},
		Units: m.GIGABYTES,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value2)*math.Pow(1024, 3), promValue)
}

func TestValueTransformer_twoDatapoints_firstValid(t *testing.T) {
	assert := assert.New(t)
	value1 := float32(2.10016)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     &value1,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     nil,
			},
		},
		Units: m.GIGABYTES_PER_HOUR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value1)*math.Pow(1024, 3), promValue)
}

func TestValueTransformer_twoDatapoints_bothValid(t *testing.T) {
	assert := assert.New(t)
	value1, value2 := float32(2.10016), float32(2.10999)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     &value1,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value2,
			},
		},
		Units: m.KILOBYTES,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value2)*1024, promValue)
}

func TestValueTransformer_twoDatapoints_bothValid_swapTimestamps(t *testing.T) {
	assert := assert.New(t)
	value1, value2 := float32(1.10016), float32(2.10999)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value2,
			},
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     &value1,
			},
		},
		Units: m.MEGABYTES,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value2)*math.Pow(1024, 2), promValue)
}

func TestValueTransformer_twoDatapoints_latestValid_swapTimestamps(t *testing.T) {
	assert := assert.New(t)
	value2 := float32(10999)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value2,
			},
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
		},
		Units: m.MILLISECONDS,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value2)*0.001, promValue)
}

func TestValueTransformer_twoDatapoints_firstValid_sameTimestamps(t *testing.T) {
	assert := assert.New(t)
	value1 := float32(2.10016)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value1,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     nil,
			},
		},
		Units: m.SCALAR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value1), promValue)
}

func TestValueTransformer_noDataPoints(t *testing.T) {
	assert := assert.New(t)
	exampleMeasurement := &m.Measurement{}
	_, err := TransformValue(exampleMeasurement)

	assert.Error(err)
}

func TestValueTransformer_oneDatapoint_valid(t *testing.T) {
	assert := assert.New(t)
	value1 := float32(2.10016)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value1,
			},
		},
		Units: m.SCALAR,
	}
	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value1), promValue)
}

func TestValueTransformer_oneDatapoint_nil(t *testing.T) {
	assert := assert.New(t)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     nil,
			},
		},
		Units: m.SCALAR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(0), promValue)
}

func TestValueTransformer_moreDataPoints_latestNotValid(t *testing.T) {
	assert := assert.New(t)
	value2 := float32(2.10999)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
			{
				Timestamp: "2021-03-04T16:53:20Z",
				Value:     &value2,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     nil,
			},
		},
		Units: m.SCALAR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value2), promValue)
}

func TestValueTransformer_moreDataPoints_lastValid(t *testing.T) {
	assert := assert.New(t)
	value3 := float32(2.10999)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
			{
				Timestamp: "2021-03-04T16:53:20Z",
				Value:     nil,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value3,
			},
		},
		Units: m.SCALAR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value3), promValue)
}

func TestValueTransformer_moreDataPoints_bothValid_duplicatedTimestamps(t *testing.T) {
	assert := assert.New(t)
	value2, value3 := float32(2.20999), float32(2.00999)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value2,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value3,
			},
		},
		Units: m.SCALAR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value2), promValue)
}

func TestValueTransformer_moreDataPoints_nilLastValid_duplicatedTimestamps(t *testing.T) {
	assert := assert.New(t)
	value3 := float32(2.10999)
	exampleMeasurement := &m.Measurement{
		DataPoints: []*mongodbatlas.DataPoints{
			{
				Timestamp: "2021-03-04T16:53:06Z",
				Value:     nil,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     &value3,
			},
			{
				Timestamp: "2021-03-04T16:54:06Z",
				Value:     nil,
			},
		},
		Units: m.SCALAR,
	}

	promValue, err := TransformValue(exampleMeasurement)

	assert.NoError(err)
	assert.Equal(float64(value3), promValue)
}
