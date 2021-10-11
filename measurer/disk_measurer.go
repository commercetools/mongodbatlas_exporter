package measurer

import (
	"mongodbatlas_exporter/model"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

// DiskMeasurements contains all measurements of one Disk
type Disk struct {
	Measurements                                map[model.MeasurementID]*model.Measurement
	ProjectID, RsName, UserAlias, PartitionName string
}

func (d *Disk) GetMeasurements() map[model.MeasurementID]*model.Measurement {
	return d.Measurements
}

func (d *Disk) LabelValues() []string {
	return []string{d.ProjectID, d.RsName, d.UserAlias, d.PartitionName}
}

func (d *Disk) LabelNames() []string {
	return []string{"project_id", "rs_name", "user_alias", "partition_name"}
}

func (d *Disk) PromLabels() prometheus.Labels {
	return prometheus.Labels{
		"project_id":     d.ProjectID,
		"rs_name":        d.RsName,
		"user_alias":     d.UserAlias,
		"partition_name": d.PartitionName,
	}
}

func (d *Disk) PromConstLabels() prometheus.Labels {
	return d.PromLabels()
}

func DiskFromMongodbAtlasProcess(p *mongodbatlas.Process, partitionName string) *Disk {
	return &Disk{
		ProjectID:     p.GroupID,
		RsName:        p.ReplicaSetName,
		UserAlias:     p.UserAlias,
		PartitionName: partitionName,
	}
}
