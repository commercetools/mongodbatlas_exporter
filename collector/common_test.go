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

var (
	defaultTestLabels = []string{"label_name1", "label_name2", "label_name3"}
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

	collector, err := newBasicCollector(logger, mock, getGivenMeasurementMetadata(), defaultTestLabels, testPrefix)
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
		prometheus.BuildFQName(namespace, testPrefix, "total_scrapes"),
		totalScrapesHelp,
		nil,
		nil)
	atlasScrapeFailuresDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, testPrefix, "atlas_scrape_failures"),
		atlasScrapeFailuresHelp,
		nil,
		nil)
	measurementTransformationFailuresDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, testPrefix, "measurement_transformation_failures"),
		measurementTransformationFailuresHelp,
		nil,
		nil)
	metricDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, testPrefix, "disk_partition_iops_read_ratio"),
		"Original measurements.name: 'DISK_PARTITION_IOPS_READ'. "+defaultHelp,
		defaultTestLabels,
		nil)
	return []*prometheus.Desc{upDesc, totalScrapesDesc, atlasScrapeFailuresDesc, measurementTransformationFailuresDesc, metricDesc}
}

func getGivenMeasurementMetadata() map[m.MeasurementID]*m.MeasurementMetadata {
	return map[m.MeasurementID]*m.MeasurementMetadata{
		m.NewMeasurementID("DISK_PARTITION_IOPS_READ", "SCALAR_PER_SECOND"): {
			Name:  "DISK_PARTITION_IOPS_READ",
			Units: "SCALAR_PER_SECOND",
		},
	}
}
