package registerer

import (
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas/mongodbatlas"
)

//TestProcessRegisterer tests that processes have collectors
//registered to a map using their ID + TypeName.
//This allows us to ensure we are not creating duplicate metrics,
//but also identify if the TypeName for an instance changes
//due to an election.
//The Registerer should be able to ADD and REMOVE instances as
//it needs to.
func TestProcesRegisterer(t *testing.T) {
	expectedProcesses := []*mongodbatlas.Process{
		{
			GroupID:        "a",
			ID:             "hosta",
			TypeName:       "PRIMARY",
			ReplicaSetName: "a",
			UserAlias:      "a",
		},
		{
			GroupID:        "b",
			ID:             "hostb",
			TypeName:       "SECONDARY",
			ReplicaSetName: "b",
			UserAlias:      "b",
		},
	}
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	client := MockClient{
		processes: expectedProcesses,
	}

	reg := NewProcessRegisterer(logger, &client, time.Millisecond)

	go reg.Observe()

	time.Sleep(2 * time.Millisecond)
	expectedProcessesMap := make(map[string]*mongodbatlas.Process)

	for i := range expectedProcesses {
		p := expectedProcesses[i]
		expectedProcessesMap[p.ID+p.TypeName] = p
	}

	assertCollectorMapInSync(t, expectedProcessesMap, reg.collectors)

	//remove b
	client.processes = expectedProcesses[0:1]
	b := expectedProcesses[1]
	delete(expectedProcessesMap, b.ID+b.TypeName)

	time.Sleep(2 * time.Millisecond)
	assertCollectorMapInSync(t, expectedProcessesMap, reg.collectors)

}

func assertCollectorMapInSync(t *testing.T, expected map[string]*mongodbatlas.Process, collectors map[string]prometheus.Collector) {
	//all the keys in reg collectors should be expected.
	for key := range collectors {
		_, ok := expected[key]
		assert.True(t, ok, "key %s not expected for registerer", key)
	}

	for key := range expected {
		_, ok := collectors[key]
		assert.True(t, ok, "reg.collectors is missing key %s", key)
	}
}
