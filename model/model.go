package model

import "go.mongodb.org/atlas/mongodbatlas"

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

type Measurer interface {
	GetMeasurements() map[MeasurementID]*Measurement
	ExtraLabels() []string
}

// DiskMeasurements contains all measurements of one Disk
type DiskMeasurements struct {
	Measurements                                map[MeasurementID]*Measurement
	ProjectID, RsName, UserAlias, PartitionName string
}

func (d *DiskMeasurements) GetMeasurements() map[MeasurementID]*Measurement {
	return d.Measurements
}

func (d *DiskMeasurements) ExtraLabels() []string {
	return []string{d.ProjectID, d.RsName, d.UserAlias, d.PartitionName}
}

// ProcessMeasurements contains all measurements of one Process
type ProcessMeasurements struct {
	Measurements                                    map[MeasurementID]*Measurement
	ProjectID, RsName, UserAlias, Version, TypeName string
}

func (p *ProcessMeasurements) GetMeasurements() map[MeasurementID]*Measurement {
	return p.Measurements
}

func (p *ProcessMeasurements) ExtraLabels() []string {
	return []string{p.ProjectID, p.RsName, p.UserAlias, p.Version, p.TypeName}
}

// Client wraps mongodbatlas.Client
type Client interface {
	GetDiskMeasurements() ([]*DiskMeasurements, ScrapeFailures, error)
	GetProcessMeasurements() ([]*ProcessMeasurements, ScrapeFailures, error)
	GetDiskMeasurementsMetadata() (map[MeasurementID]*MeasurementMetadata, error)
	GetProcessMeasurementsMetadata() (map[MeasurementID]*MeasurementMetadata, error)
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
