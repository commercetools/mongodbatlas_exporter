package transformer

import (
	m "mongodbatlas_exporter/model"

	"github.com/prometheus/client_golang/prometheus"
)

// TransformType transforms Measurement into prometheus.ValueType
func TransformType(measurement *m.Measurement) (prometheus.ValueType, error) {
	// According to current knowledge all mongodbatlas Measurements are Gauges.
	return prometheus.GaugeValue, nil
}
