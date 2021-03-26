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

// DiskMeasurements contains all measurements of one Disk
type DiskMeasurements struct {
	ProjectID, RsName, UserAlias, PartitionName string
	Measurements                                map[MeasurementID]*Measurement
}

// ProcessMeasurements contains all measurements of one Process
type ProcessMeasurements struct {
	ProjectID, RsName, UserAlias string
	Measurements                 map[MeasurementID]*Measurement
}

// Client wraps mongodbatlas.Client
type Client interface {
	GetDiskMeasurements() ([]*DiskMeasurements, ScrapeFailures, error)
	GetProcessMeasurements() ([]*ProcessMeasurements, ScrapeFailures, error)
	GetDiskMeasurementsMetadata() ([]*MeasurementMetadata, error)
	GetProcessMeasurementsMetadata() ([]*MeasurementMetadata, error)
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
