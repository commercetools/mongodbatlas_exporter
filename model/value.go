package model

import (
	"errors"
	"math"
	"sort"
	"time"

	"go.mongodb.org/atlas/mongodbatlas"
)

const timestampFormat = "2006-01-02T15:04:05Z"

var ErrNoDatapoints = errors.New("no datapoints are available")

func sortDataPoints(dataPoints *[]*mongodbatlas.DataPoints) {
	sort.Slice(*dataPoints, func(i, j int) bool {
		t1, _ := time.Parse(timestampFormat, (*dataPoints)[i].Timestamp)
		t2, _ := time.Parse(timestampFormat, (*dataPoints)[j].Timestamp)
		v1Ptr, v2Ptr := (*dataPoints)[i].Value, (*dataPoints)[j].Value
		if t1 != t2 {
			return t1.Before(t2)
		}

		if v1Ptr == nil {
			return true
		}

		if v2Ptr != nil {
			return *v1Ptr < *v2Ptr
		}

		return false
	})
}

func containsValidDataPoints(datapoints []*mongodbatlas.DataPoints) error {
	emptyDataPoints := len(datapoints) < 1

	if emptyDataPoints {
		return ErrNoDatapoints
	}
	for _, dataPoint := range datapoints {
		_, err := time.Parse(timestampFormat, dataPoint.Timestamp)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertValue(value float64, unit UnitEnum) float64 {
	multiplier := unitsTransformationRules[unit].valueMultiplier
	return value * multiplier
}

// TransformValue transforms Measurements into float64 for Prometheus metric value
func (m *Measurement) PromVal() (float64, error) {
	dataPoints := m.DataPoints
	unit := m.Units
	err := containsValidDataPoints(dataPoints)
	if err != nil {
		return math.NaN(), err
	}
	sortDataPoints(&dataPoints)

	for i := len(dataPoints) - 1; i >= 0; i-- {
		value := dataPoints[i].Value
		if value != nil {
			return convertValue(float64(*value), unit), nil
		}
	}

	return float64(0), nil
}
