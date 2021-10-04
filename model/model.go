package model

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

// ScrapeFailures shows number of failed Measurements scapes
type ScrapeFailures int

// Client wraps mongodbatlas.Client
type Client interface {
	GetDiskMeasurements() ([]*DiskMeasurements, ScrapeFailures, error)
	GetProcessMeasurements() ([]*ProcessMeasurements, ScrapeFailures, error)
	GetDiskMeasurementMap() (MeasurementMap, error)
	GetProcessMeasurementMap() (MeasurementMap, error)
}
