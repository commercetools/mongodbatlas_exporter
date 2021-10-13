package measurer

import (
	"mongodbatlas_exporter/model"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

// DiskMeasurements contains all measurements of one Disk
type Disk struct {
	Measurements  map[model.MeasurementID]*model.Measurement
	Metadata      map[model.MeasurementID]*model.MeasurementMetadata
	PartitionName string
}

func (d *Disk) GetMeasurements() map[model.MeasurementID]*model.Measurement {
	return d.Measurements
}

func (d *Disk) GetMetaData() map[model.MeasurementID]*model.MeasurementMetadata {
	return d.Metadata
}

func (d *Disk) LabelValues() []string {
	return []string{d.PartitionName}
}

func (d *Disk) LabelNames() []string {
	return []string{"partition_name"}
}

func (d *Disk) PromLabels() prometheus.Labels {
	return prometheus.Labels{
		"partition_name": d.PartitionName,
	}
}

func (d *Disk) PromConstLabels() prometheus.Labels {
	return d.PromLabels()
}

func DiskFromMongodbAtlasProcessDisk(p *mongodbatlas.ProcessDisk) *Disk {
	return &Disk{
		PartitionName: p.PartitionName,
	}
}
