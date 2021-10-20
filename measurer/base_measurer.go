package measurer

import (
	"fmt"
	"mongodbatlas_exporter/collector/transformer"
	"mongodbatlas_exporter/model"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	DEFAULT_HELP = "Please see MongoDB Atlas documentation for details about the measurement"
)

//Base helps collect most functionality for uniquely identifying resources.
type Base struct {
	Measurements map[model.MeasurementID]*model.Measurement
	Metadata     map[model.MeasurementID]*model.MeasurementMetadata
	//These fields help uniquely identify processes and disks.
	ProjectID, RsName, UserAlias, TypeName, Hostname, ID string
	promMetrics                                          []*PromMetric
}

func (b *Base) GetMetaData() map[model.MeasurementID]*model.MeasurementMetadata {
	return b.Metadata
}

func (b *Base) GetMeasurements() map[model.MeasurementID]*model.Measurement {
	return b.Measurements
}

func (b *Base) PromMetrics() []*PromMetric {
	return b.promMetrics
}

//LabelValues does not return the type and version as it would lead
//to too much cardinality.
func (b *Base) PromVariableLabelValues() []string {
	return []string{}
}

//LabelNames does not return the type and version as it would lead
//to too much cardinality. Metrics that need these extra fields should
//access them directly.
func (b *Base) PromVariableLabelNames() []string {
	return []string{}
}

func (b *Base) setPromMetrics(metrics []*PromMetric) {
	b.promMetrics = metrics
}

func (b *Base) PromInfoConstLabels() prometheus.Labels {
	return b.PromConstLabels()
}

//PromLabels as with many other Process methods
//version and type are excluded here as they are often not necessary
//for identifying a particular instance and increase cardinality.
func (b *Base) PromConstLabels() prometheus.Labels {
	return prometheus.Labels{
		"project_id": b.ProjectID,
		"rs_name":    b.RsName,
		"user_alias": b.UserAlias,
	}
}

//metadataToMetric transforms the measurement metadata we received from Atlas into a
//prometheus compatible metric description.
func metadataToMetric(metadata *model.MeasurementMetadata, namespace, collectorPrefix, defaultHelp string, variableLabels []string, constLabels prometheus.Labels) (*PromMetric, error) {
	promName, err := transformer.TransformName(metadata)
	if err != nil {
		msg := "can't transform measurement Name (%s) into metric name"
		return nil, fmt.Errorf(msg, metadata.Name)
	}
	promType, err := transformer.TransformType(metadata)
	if err != nil {
		msg := "can't transform measurement Units (%s) into prometheus.ValueType"
		return nil, fmt.Errorf(msg, metadata.Units)
	}

	metric := PromMetric{
		Type: promType,
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, collectorPrefix, promName),
			"Original measurements.name: '"+metadata.Name+"'. "+defaultHelp,
			variableLabels, constLabels,
		),
		Metadata: metadata,
	}

	return &metric, nil
}
