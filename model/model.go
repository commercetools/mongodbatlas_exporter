package model

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

// UnitEnum is a enum of supported Messurements Units
type UnitEnum string

const (
	PERCENT              UnitEnum = "PERCENT"
	MILLISECONDS         UnitEnum = "MILLISECONDS"
	SECONDS              UnitEnum = "SECONDS"
	BYTES                UnitEnum = "BYTES"
	KILOBYTES            UnitEnum = "KILOBYTES"
	MEGABYTES            UnitEnum = "MEGABYTES"
	GIGABYTES            UnitEnum = "GIGABYTES"
	BYTES_PER_SECOND     UnitEnum = "BYTES_PER_SECOND"
	MEGABYTES_PER_SECOND UnitEnum = "MEGABYTES_PER_SECOND"
	GIGABYTES_PER_HOUR   UnitEnum = "GIGABYTES_PER_HOUR"
	SCALAR_PER_SECOND    UnitEnum = "SCALAR_PER_SECOND"
	SCALAR               UnitEnum = "SCALAR"
)

// MeasurementID consists of Measurement.Name and Measurement.Units
type MeasurementID string

// NewMeasurementID creates MeasurementId from name and units
func NewMeasurementID(name, unit string) MeasurementID {
	return MeasurementID(name + "_" + unit)
}

// ScrapeFailures shows number of failed Measurements scapes
type ScrapeFailures int

// Measurement contains unit and mulpiple dataPoints of one measurement
type Measurement struct {
	DataPoints []*mongodbatlas.DataPoints
	Units      UnitEnum
}

//Measurer interface allows DiskMeasurements and ProcessMeasurements
//to share common methods in this package and external packages.
type Measurer interface {
	//Getter for the measurement map
	GetMeasurements() map[MeasurementID]*Measurement
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
}

// DiskMeasurements contains all measurements of one Disk
type DiskMeasurements struct {
	Measurements                                map[MeasurementID]*Measurement
	ProjectID, RsName, UserAlias, PartitionName string
}

func (d *DiskMeasurements) GetMeasurements() map[MeasurementID]*Measurement {
	return d.Measurements
}

func (d *DiskMeasurements) LabelValues() []string {
	return []string{d.ProjectID, d.RsName, d.UserAlias, d.PartitionName}
}

func (d *DiskMeasurements) LabelNames() []string {
	return []string{"project_id", "rs_name", "user_alias", "partition_name"}
}

func (d *DiskMeasurements) PromLabels() prometheus.Labels {
	return prometheus.Labels{
		"project_id":     d.ProjectID,
		"rs_name":        d.RsName,
		"user_alias":     d.UserAlias,
		"partition_name": d.PartitionName,
	}
}

func (d *DiskMeasurements) PromConstLabels() prometheus.Labels {
	return d.PromLabels()
}

// ProcessMeasurements contains all measurements of one Process
type ProcessMeasurements struct {
	Measurements                                                  map[MeasurementID]*Measurement
	ProjectID, RsName, UserAlias, Version, TypeName, Hostname, ID string
	Port                                                          int
}

func (p *ProcessMeasurements) GetMeasurements() map[MeasurementID]*Measurement {
	return p.Measurements
}

//LabelValues does not return the type and version as it would lead
//to too much cardinality.
func (p *ProcessMeasurements) LabelValues() []string {
	return []string{p.ProjectID, p.ID, p.RsName, p.UserAlias}
}

//LabelNames does not return the type and version as it would lead
//to too much cardinality. Metrics that need these extra fields should
//access them directly.
func (p *ProcessMeasurements) LabelNames() []string {
	return []string{"project_id", "id", "rs_name", "user_alias"}
}

//AllLabelNames is a special case for the process info measurement.
//includes version and node type.
func (p *ProcessMeasurements) AllLabelNames() []string {
	return append(p.LabelNames(), "version", "type")
}

//AllLabelValues is a special case for the process info measurement.
//includes version and node type.
func (p *ProcessMeasurements) AllLabelValues() []string {
	return append(p.LabelValues(), p.Version, p.TypeName)
}

//PromLabels as with many other ProcessMeasurements methods
//version and type are excluded here as they are often not necessary
//for identifying a particular instance and increase cardinality.
func (p *ProcessMeasurements) PromLabels() prometheus.Labels {
	return prometheus.Labels{
		"project_id": p.ProjectID,
		"rs_name":    p.RsName,
		"user_alias": p.UserAlias,
		"id":         p.ID,
	}
}

func (p *ProcessMeasurements) PromConstLabels() prometheus.Labels {
	return p.PromLabels()
}

// MeasurementMetadata contains Measurements.Name and Measurements.Unit
type MeasurementMetadata struct {
	Name  string
	Units UnitEnum
}

// ID returns identifier of the metric
func (c MeasurementMetadata) ID() MeasurementID {
	return NewMeasurementID(c.Name, string(c.Units))
}
