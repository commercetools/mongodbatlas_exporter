package measurer

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

// Process contains all measurements of one Process
type Process struct {
	Base
	Disks   []*Disk
	Version string
	Port    int
}

func (p *Process) PromInfoConstLabels() prometheus.Labels {
	labels := p.Base.PromInfoConstLabels()
	labels["version"] = p.Version
	labels["type"] = p.TypeName
	return labels
}

//FromMongodbAtlasProcess creates a measurer.Process by extracting
//the important features from a mongodbatlas.Process so we can uniquely
//identify prometheus metrics using labels.
func ProcessFromMongodbAtlasProcess(p *mongodbatlas.Process) *Process {
	return &Process{
		Base: Base{
			ProjectID: p.GroupID,
			RsName:    p.ReplicaSetName,
			UserAlias: p.UserAlias,
			TypeName:  p.TypeName,
			Hostname:  p.Hostname,
			ID:        p.ID,
		},
		Port:    p.Port,
		Version: p.Version,
	}
}
