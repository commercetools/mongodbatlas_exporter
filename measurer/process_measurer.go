package measurer

import (
	"mongodbatlas_exporter/model"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

// Process contains all measurements of one Process
type Process struct {
	Metadata                                                      map[model.MeasurementID]*model.MeasurementMetadata
	Measurements                                                  map[model.MeasurementID]*model.Measurement
	Disks                                                         []*Disk
	ProjectID, RsName, UserAlias, Version, TypeName, Hostname, ID string
	Port                                                          int
}

func (p *Process) GetMetaData() map[model.MeasurementID]*model.MeasurementMetadata {
	return p.Metadata
}

func (p *Process) GetMeasurements() map[model.MeasurementID]*model.Measurement {
	return p.Measurements
}

//LabelValues does not return the type and version as it would lead
//to too much cardinality.
func (p *Process) PromVariableLabelValues() []string {
	return []string{}
}

//LabelNames does not return the type and version as it would lead
//to too much cardinality. Metrics that need these extra fields should
//access them directly.
func (p *Process) PromVariableLabelNames() []string {
	return []string{}
}

//PromLabels as with many other Process methods
//version and type are excluded here as they are often not necessary
//for identifying a particular instance and increase cardinality.
func (p *Process) PromConstLabels() prometheus.Labels {
	return prometheus.Labels{
		"project_id": p.ProjectID,
		"rs_name":    p.RsName,
		"user_alias": p.UserAlias,
		"id":         p.ID,
	}
}

//FromMongodbAtlasProcess creates a measurer.Process by extracting
//the important features from a mongodbatlas.Process so we can uniquely
//identify prometheus metrics using labels.
func ProcessFromMongodbAtlasProcess(p *mongodbatlas.Process) *Process {
	return &Process{
		ProjectID: p.GroupID,
		RsName:    p.ReplicaSetName,
		UserAlias: p.UserAlias,
		Version:   p.Version,
		TypeName:  p.TypeName,
		Hostname:  p.Hostname,
		Port:      p.Port,
		ID:        p.ID,
	}
}
