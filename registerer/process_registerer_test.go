package registerer

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
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
	g := gomega.NewGomegaWithT(t)

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

	expectedProcessesMap := make(map[string]*mongodbatlas.Process)

	for i := range expectedProcesses {
		p := expectedProcesses[i]
		expectedProcessesMap[p.ID+p.TypeName] = p
	}

	timeout := 10 * time.Second
	registeredCount := func() int {
		return len(reg.collectors)
	}

	g.Eventually(registeredCount).Should(gomega.Equal(len(expectedProcessesMap)))
	g.Eventually(assertCollectorMapInSync(g, expectedProcessesMap, reg.collectors)).Should(gomega.Succeed())

	//remove b
	client.processes = expectedProcesses[0:1]
	b := expectedProcesses[1]
	delete(expectedProcessesMap, b.ID+b.TypeName)

	g.Eventually(registeredCount, timeout, 1*time.Millisecond).Should(gomega.Equal(len(expectedProcessesMap)))
	g.Eventually(assertCollectorMapInSync(g, expectedProcessesMap, reg.collectors)).Should(gomega.Succeed())

	//re-add b
	client.processes = expectedProcesses
	expectedProcessesMap[b.ID+b.TypeName] = b

	g.Eventually(registeredCount, timeout, 1*time.Millisecond).Should(gomega.Equal(len(expectedProcessesMap)))
	g.Eventually(assertCollectorMapInSync(g, expectedProcessesMap, reg.collectors)).Should(gomega.Succeed())

	//simulate re-election
	expectedProcesses[0].TypeName = "SECONDARY"
	expectedProcesses[1].TypeName = "PRIMARY"

	expectedProcessesMap = make(map[string]*mongodbatlas.Process, len(expectedProcesses))
	for i := range expectedProcesses {
		p := expectedProcesses[i]
		expectedProcessesMap[p.ID+p.TypeName] = p
	}
	g.Eventually(registeredCount, timeout, 1*time.Millisecond).Should(gomega.Equal(len(expectedProcessesMap)))
	g.Eventually(assertCollectorMapInSync(g, expectedProcessesMap, reg.collectors)).Should(gomega.Succeed())
}

func assertCollectorMapInSync(g *gomega.GomegaWithT, expected map[string]*mongodbatlas.Process, collectors map[string]prometheus.Collector) func() error {
	return func() error {
		//all the keys in reg collectors should be expected.
		for key := range collectors {
			_, ok := expected[key]
			if !ok {
				return fmt.Errorf("key %s not expected for registerer", key)
			}
		}

		for key := range expected {
			_, ok := collectors[key]
			if !ok {
				return fmt.Errorf("reg.collectors is missing key %s", key)
			}
		}
		return nil
	}

}
