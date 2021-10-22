package measurer

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

// DiskMeasurements contains all measurements of one Disk
type Disk struct {
	Base
	PartitionName string
}

func (d *Disk) PromConstLabels() prometheus.Labels {
	labels := d.Base.PromConstLabels()
	labels["paritition_name"] = d.PartitionName
	return labels
}

func DiskFromMongodbAtlasProcessDisk(p *mongodbatlas.Process, d *mongodbatlas.ProcessDisk) *Disk {
	return &Disk{
		Base:          *baseFromMongodbAtlasProcess(p),
		PartitionName: d.PartitionName,
	}
}
