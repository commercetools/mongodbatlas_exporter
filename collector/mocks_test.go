package collector

import (
	"mongodbatlas_exporter/measurer"
	"mongodbatlas_exporter/model"
	a "mongodbatlas_exporter/mongodbatlas"

	"go.mongodb.org/atlas/mongodbatlas"
)

func (c *MockClient) GetDiskMeasurements(_ *measurer.Process, d *measurer.Disk) error {
	d.Measurements = c.givenDisksMeasurements
	return nil
}

func (c *MockClient) GetDiskMeasurementsMetadata(_ *measurer.Process, _ *measurer.Disk) (map[model.MeasurementID]*model.MeasurementMetadata, error) {
	return map[model.MeasurementID]*model.MeasurementMetadata{
		model.NewMeasurementID("DISK_PARTITION_IOPS_READ", "SCALAR_PER_SECOND"): {
			Name:  "DISK_PARTITION_IOPS_READ",
			Units: "SCALAR_PER_SECOND",
		},
		model.NewMeasurementID("DISK_PARTITION_SPACE_USED", "BYTES"): {
			Name:  "DISK_PARTITION_SPACE_USED",
			Units: "BYTES",
		},
	}, nil
}

func (c *MockClient) GetProcessMeasurements(_ measurer.Process) (map[model.MeasurementID]*model.Measurement, error) {
	return make(map[model.MeasurementID]*model.Measurement), nil
}

func (c *MockClient) ListDisks(*mongodbatlas.Process) ([]*mongodbatlas.ProcessDisk, *a.HTTPError) {
	return []*mongodbatlas.ProcessDisk{
		{PartitionName: "data"},
	}, nil
}

func (c *MockClient) GetProcessMeasurementsMetadata(p *measurer.Process) *a.HTTPError {
	p.Metadata = map[model.MeasurementID]*model.MeasurementMetadata{
		model.NewMeasurementID("TICKETS_AVAILABLE_READS", "SCALAR"): {
			Name:  "TICKETS_AVAILABLE_READS",
			Units: "SCALAR",
		},
		model.NewMeasurementID("QUERY_EXECUTOR_SCANNED", "SCALAR_PER_SECOND"): {
			Name:  "QUERY_EXECUTOR_SCANNED",
			Units: "SCALAR_PER_SECOND",
		},
	}
	return nil
}
