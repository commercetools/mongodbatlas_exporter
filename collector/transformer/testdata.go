package transformer

import (
	m "mongodbatlas_exporter/model"
)

var (
	exampleMeasurement = &m.Measurement{
		Name:  "DISK_PARTITION_IOPS_READ",
		Units: "SCALAR_PER_SECOND",
	}
)
