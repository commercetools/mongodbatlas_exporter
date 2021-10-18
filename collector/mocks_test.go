package collector

import "mongodbatlas_exporter/model"

func getExpectedDiskMeasurementMetadata() map[model.MeasurementID]*model.MeasurementMetadata {
	return map[model.MeasurementID]*model.MeasurementMetadata{
		model.NewMeasurementID("DISK_PARTITION_IOPS_READ", "SCALAR_PER_SECOND"): {
			Name:  "DISK_PARTITION_IOPS_READ",
			Units: "SCALAR_PER_SECOND",
		},
	}
}
