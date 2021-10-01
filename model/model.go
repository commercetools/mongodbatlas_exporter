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
	Name       string
	DataPoints []*mongodbatlas.DataPoints
	Units      UnitEnum
}

// DiskMeasurements contains all measurements of one Disk
type DiskMeasurements struct {
	ProjectID, RsName, UserAlias, PartitionName string
	Measurements                                MeasurementMap
}

// ProcessMeasurements contains all measurements of one Process
type ProcessMeasurements struct {
	ProjectID, RsName, UserAlias, Version, TypeName string
	Measurements                                    MeasurementMap
}

// Client wraps mongodbatlas.Client
type Client interface {
	GetDiskMeasurements() ([]*DiskMeasurements, ScrapeFailures, error)
	GetProcessMeasurements() ([]*ProcessMeasurements, ScrapeFailures, error)
	GetDiskMeasurementMap() (MeasurementMap, error)
	GetProcessMeasurementMap() (MeasurementMap, error)
}

// ID returns identifier of the metric
func (c Measurement) ID() MeasurementID {
	return NewMeasurementID(c.Name, string(c.Units))
}
