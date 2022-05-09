package measurer

import (
	"fmt"
	"mongodbatlas_exporter/collector/transformer"
	"mongodbatlas_exporter/model"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

const (
	DEFAULT_HELP = "Please see MongoDB Atlas documentation for details about the measurement"
)

//Base helps collect most functionality for uniquely identifying resources.
type Base struct {
	Measurements map[model.MeasurementID]*model.Measurement
	Metadata     map[model.MeasurementID]*model.MeasurementMetadata
	//These fields help uniquely identify processes and disks.
	ProjectID, RsName, TypeName, Hostname, Sharded_cluster, Project_name string
	ProcessType                                                          string
	//UserAlias is the "friendly" hostname that includes
	//a user specified prefix.
	UserAlias string
	//ID is a unique hostname and port alloted by Atlas.
	//This ID is used to retrieve an individual process
	//from the API.
	ID          string
	promMetrics []*PromMetric
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
//Typename,Sharded_cluster,projectname labels added
func (b *Base) PromConstLabels() prometheus.Labels {
	return prometheus.Labels{
		"project_id":     b.ProjectID,
		"rs_name":        b.RsName,
		"user_alias":     b.UserAlias,
		"process_state":  b.TypeName,
		"shared_cluster": b.Sharded_cluster,
		"project_name":   b.Project_name,
		"process_type":   b.ProcessType,
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

//baseFromMongodbAtlasProcess populates the base fields for both disk
//and process measurers.
func baseFromMongodbAtlasProcess(p *mongodbatlas.Process) *Base {
	//based on replicasetname get process_type
	pt := strings.Split(p.ReplicaSetName, "-")[2]
	if pt == "shard" {
		pt = "mongod"
	}
	if pt == "config" {
		pt = "configsvr"
	}
	if pt == "mongos" {
		pt = "mongos"
	}
	return &Base{
		ProjectID: p.GroupID,
		RsName:    p.ReplicaSetName,
		//We append the port to the UserAlias so that UserAlias becomes unique.
		//Often the MONGOS is hosted on the same host as the REPLICAS so only
		//the port will make it unique.
		UserAlias:       p.UserAlias + fmt.Sprintf(":%d", p.Port),
		TypeName:        strings.Split(p.TypeName, "_")[1],
		Hostname:        p.Hostname,
		ID:              p.ID,
		Sharded_cluster: strings.Join(strings.Split(p.ReplicaSetName, "-")[:2], "-"),
		Project_name:    strings.Join(strings.Split(p.UserAlias, "-")[:2], "-"),
		ProcessType:     pt,
	}
}
