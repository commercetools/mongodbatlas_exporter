package collector

import (
	"mongodbatlas_exporter/measurer"
	m "mongodbatlas_exporter/model"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

const testPrefix = "stats"

type MockClient struct {
	givenDisksMeasurements     []*measurer.Disk
	givenProcessesMeasurements []*measurer.Process
}

func convertMetrics(metrics []prometheus.Metric) map[string]string {
	result := make(map[string]string, len(metrics))
	for _, metric := range metrics {
		desc := metric.Desc().String()
		dtoMetric := dto.Metric{}
		metric.Write(&dtoMetric)
		result[desc] = dtoMetric.String()
	}
	return result
}

//TestDesc ensures that the basic collector's Describe function properly registers prometheus
//metrics.
//It is important to test here that the Prometheus Descriptions have the correct descriptions
//and uniquely identifying Constant Labels.
//If Variable labels are added to metrics in the future (such as HTTP Status Code) that should
//be tested for as well.
//However, since those fields are private the best way to test is using a deep equality function
//from a testing framework.
func TestDesc(t *testing.T) {
	assert := assert.New(t)
	mock := &MockClient{}
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	collector, err := newBasicCollector(logger, mock, getGivenMeasurementMetadata(), &measurer.Disk{}, testPrefix)
	assert.NotNil(collector)
	assert.NoError(err)
	descCh := make(chan *prometheus.Desc, 99)
	defer close(descCh)
	collector.Describe(descCh)
	var resultingDescs []*prometheus.Desc

	for len(descCh) > 0 {
		resultingDescs = append(resultingDescs, <-descCh)
	}
	expectedDescs := getExpectedDescs()
	assert.Equal(len(expectedDescs), len(resultingDescs))
	//Since all fields are private this gives the most actionable info.
	//Can be difficult to read/understand output.
	assert.ElementsMatch(expectedDescs, resultingDescs)
}

//getExpectedDescs is a mocking function to return the expected list of descriptions for a basic collector.
//In this instance the basicCollector is for disks since we are using the constant labels from model.DiskMeasurements
func getExpectedDescs() []*prometheus.Desc {
	fqNames := []string{
		prometheus.BuildFQName(namespace, testPrefix, "up"),
		prometheus.BuildFQName(namespace, testPrefix, "scrapes_total"),
		prometheus.BuildFQName(namespace, testPrefix, "scrape_failures_total"),
		prometheus.BuildFQName(namespace, testPrefix, "disk_partition_iops_read_ratio"),
	}
	help := []string{
		upHelp,
		totalScrapesHelp,
		scrapeFailuresHelp,
		"Original measurements.name: 'DISK_PARTITION_IOPS_READ'. " + defaultHelp,
	}

	result := make([]*prometheus.Desc, len(fqNames))

	for i := range fqNames {
		//Build the description and add the constant labels. Constant labels are used to uniquely identify a measurement.
		//Whereas variable lables such as HTTP Status Codes provide more context.
		result[i] = prometheus.NewDesc(fqNames[i], help[i], nil, (&measurer.Disk{}).PromConstLabels())
	}

	return result
}

func getGivenMeasurementMetadata() map[m.MeasurementID]*m.MeasurementMetadata {
	return map[m.MeasurementID]*m.MeasurementMetadata{
		m.NewMeasurementID("DISK_PARTITION_IOPS_READ", "SCALAR_PER_SECOND"): {
			Name:  "DISK_PARTITION_IOPS_READ",
			Units: "SCALAR_PER_SECOND",
		},
	}
}
