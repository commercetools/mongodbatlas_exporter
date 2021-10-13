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
	//Returns slice of prometheus label values such as the real project_id or rs_name value.
	LabelValues() []string
	//Returns slice of prometheus label names so that collectors can register necessary
	//labels during describe.
	LabelNames() []string
	//Returns a map of prometheus.Labels, mostly useful with CounterVec.With to
	//select a particular counter to manipulate.
	PromLabels() prometheus.Labels
	//Returns a map of prometheus.Labels used where constant labels should be used.
	PromConstLabels() prometheus.Labels
	PromVariableLabelValues() []string
}
