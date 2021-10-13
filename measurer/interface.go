package measurer

import (
	"mongodbatlas_exporter/model"

	"github.com/prometheus/client_golang/prometheus"
)

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
}
