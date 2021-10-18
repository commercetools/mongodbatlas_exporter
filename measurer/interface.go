package measurer

import (
	"mongodbatlas_exporter/model"

	"github.com/prometheus/client_golang/prometheus"
)

type PromMetric struct {
	Type     prometheus.ValueType
	Desc     *prometheus.Desc
	Metadata *model.MeasurementMetadata
}

//ErrorLabels consumes prometheus.Labels and adds more labels to the map.
//Perhaps this is a chainable pattern we can reuse on other types to have a
//consistent interface for working with labels.
//The prometheus API is fairly inconcsistent where many APIs require a slice of
//label names or label values.
//ErrorLabels original need was to combine labels from a Measurer and use them
//with a prometheus.CounterVec to select a particular counter.
func (x *PromMetric) ErrorLabels(extraLabels prometheus.Labels) prometheus.Labels {
	result := prometheus.Labels{
		"atlas_metric": x.Metadata.Name,
	}

	for key, value := range extraLabels {
		result[key] = value
	}

	return result
}

//Measurer interface allows DiskMeasurements and ProcessMeasurements
//to share common methods in this package and external packages.
type Measurer interface {
	//Getter for the measurement map
	GetMeasurements() map[model.MeasurementID]*model.Measurement
	//Getter for the measurement metadata map
	GetMetaData() map[model.MeasurementID]*model.MeasurementMetadata
	//Returns slice of prometheus variable label values such as a disks partition name.
	PromVariableLabelValues() []string
	//Returns slice of variable prometheus label names so that collectors can register necessary
	//labels during describe.
	PromVariableLabelNames() []string
	//Returns a map of prometheus.Labels used where constant labels should be used.
	PromConstLabels() prometheus.Labels
	PromMetrics() []*PromMetric
}
