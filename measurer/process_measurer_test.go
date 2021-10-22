package measurer

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas/mongodbatlas"
)

//TestProcessInfoLabels checks that the
//extra informational labels added to the "info"
//measurement are added.
//Current informational labels include:
//version: the node version
//type: the node's current replica status.
func TestProcessInfoLabels(t *testing.T) {
	b := Process{
		Base: Base{
			ProjectID: "9u1ji2k3jlj",
			RsName:    "rs1",
			UserAlias: "cluster-rs1:27017",
			TypeName:  "REPLICA_PRIMARY",
			Hostname:  "cluster-rs1",
			ID:        "kj1lk2jklji",
		},
		Version: "4.2.17",
	}

	//if more labels are needed make sure to include
	//a reason why here.
	allowedLabels := map[string]bool{
		"project_id": true, //0.0.2
		"rs_name":    true, //0.0.2
		"user_alias": true, //0.0.2
		"version":    true, //0.0.2
		"type":       true, //0.0.2
	}

	labels := b.PromInfoConstLabels()

	for k := range allowedLabels {
		if _, ok := labels[k]; !ok {
			t.Fatalf("missing label %s", k)
		}
	}

	for k := range labels {
		if _, ok := allowedLabels[k]; !ok {
			t.Fatalf("extra label %s", k)
		}
	}
}

func TestProcessFromMongodbAtlasProcess(t *testing.T) {
	process := mongodbatlas.Process{
		GroupID:        "9uf201u9ur1",
		Hostname:       "atlas-xkljjzl.asdf.mongodb.net",
		ID:             "atlas-xkljjzl.asdf.mongodb.net:27017",
		ReplicaSetName: "atlas-xkljjzl-config-0",
		UserAlias:      "aname-config-00-0.asdf.mongodb.net",
		Port:           27017,
	}

	actual := ProcessFromMongodbAtlasProcess(&process)

	//Ensure that the measurer appends the port to the useralias
	//so that the useralias is unique.
	//MONGOS processes often share the same host as their REPLICAS.
	userAliasSplit := strings.Split(actual.UserAlias, ":")

	assert.Equal(t, process.UserAlias, userAliasSplit[0])

	port, err := strconv.Atoi(userAliasSplit[1])

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, process.Port, port)

}
