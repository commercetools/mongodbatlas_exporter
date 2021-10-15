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

type Base struct {
	Measurements                                         map[model.MeasurementID]*model.Measurement
	Metadata                                             map[model.MeasurementID]*model.MeasurementMetadata
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

//PromLabels as with many other Process methods
//version and type are excluded here as they are often not necessary
//for identifying a particular instance and increase cardinality.
func (b *Base) PromConstLabels() prometheus.Labels {
	return prometheus.Labels{
		"project_id": b.ProjectID,
		"rs_name":    b.RsName,
		"user_alias": b.UserAlias,
		"id":         b.ID,
	}
}

//buildPromMetrics returns a function that allows Process and Disk measurers to inject their specific labels
//and finally the collector to inject its namespace and prefix.
func (b *Base) BuildPromMetrics(namespace, collectorPrefix string) error {
	b.promMetrics = make([]*PromMetric, len(b.Metadata))

	i := 0
	for _, metadata := range b.Metadata {
		metric, err := metadataToMetric(metadata, namespace, collectorPrefix, DEFAULT_HELP, b.PromVariableLabelNames(), b.PromConstLabels())
		if err != nil {
			return err
		}
		b.promMetrics[i] = metric
		i++
	}
	return nil
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
