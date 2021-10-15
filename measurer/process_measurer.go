package measurer

import (
	"strconv"

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

//PromLabels as with many other Process methods
//version and type are excluded here as they are often not necessary
//for identifying a particular instance and increase cardinality.
func (p *Process) PromConstLabels() prometheus.Labels {
	labels := p.Base.PromConstLabels()
	labels["version"] = p.Version
	labels["port"] = strconv.FormatInt(int64(p.Port), 10)
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
