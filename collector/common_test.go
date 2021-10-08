package collector

import (
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
	givenDisksMeasurements     []*m.DiskMeasurements
	givenProcessesMeasurements []*m.ProcessMeasurements
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

func TestDesc(t *testing.T) {
	assert := assert.New(t)
	mock := &MockClient{}
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	collector, err := newBasicCollector(logger, mock, getGivenMeasurementMetadata(), &m.DiskMeasurements{}, testPrefix)
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
	assert.ElementsMatch(expectedDescs, resultingDescs)
}

func getExpectedDescs() []*prometheus.Desc {
	upDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, testPrefix, "up"),
		upHelp,
		nil,
		nil)
	totalScrapesDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, testPrefix, "scrapes_total"),
		totalScrapesHelp,
		nil,
		nil)
	scrapeFailuresDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, testPrefix, "scrape_failures_total"),
		scrapeFailuresHelp,
		nil,
		nil)
	// no measurementTransformationFailuresDesc because it uses a CounterVec
	// we would need to return ALL combinations of labels at start time to
	// maintain the "checked" exporter status.
	metricDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, testPrefix, "disk_partition_iops_read_ratio"),
		"Original measurements.name: 'DISK_PARTITION_IOPS_READ'. "+defaultHelp,
		(&m.DiskMeasurements{}).LabelNames(),
		nil)
	return []*prometheus.Desc{upDesc, totalScrapesDesc, scrapeFailuresDesc, metricDesc}
}

func getGivenMeasurementMetadata() map[m.MeasurementID]*m.MeasurementMetadata {
	return map[m.MeasurementID]*m.MeasurementMetadata{
		m.NewMeasurementID("DISK_PARTITION_IOPS_READ", "SCALAR_PER_SECOND"): {
			Name:  "DISK_PARTITION_IOPS_READ",
			Units: "SCALAR_PER_SECOND",
		},
	}
}
