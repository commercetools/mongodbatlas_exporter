package collector

import (
	m "mongodbatlas_exporter/model"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas/mongodbatlas"
)

func (c *MockClient) GetDiskMeasurements() ([]*m.DiskMeasurements, m.ScrapeFailures, error) {
	return c.givenDisksMeasurements, 3, nil
}

func (c *MockClient) GetDiskMeasurementMap() (m.MeasurementMap, error) {
	return m.MeasurementMap{
		m.NewMeasurementID("DISK_PARTITION_IOPS_READ", "SCALAR_PER_SECOND"): {
			Name:  "DISK_PARTITION_IOPS_READ",
			Units: "SCALAR_PER_SECOND",
		},
		m.NewMeasurementID("DISK_PARTITION_SPACE_USED", "BYTES"): {
			Name:  "DISK_PARTITION_SPACE_USED",
			Units: "BYTES",
		},
	}, nil
}

func TestDisksCollector(t *testing.T) {
	assert := assert.New(t)
	value := float32(3.14)
	mock := &MockClient{}
	mock.givenDisksMeasurements = getGivenMeasurements(&value)
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	disks, err := NewDisks(logger, mock)
	assert.NotNil(disks)
	assert.NoError(err)

	metricsCh := make(chan prometheus.Metric, 99)
	defer close(metricsCh)
	expectedMetrics := getExpectedDisksMetrics(float64(value))
	disks.Collect(metricsCh)
	var resultingMetrics []prometheus.Metric
	for len(metricsCh) > 0 {
		resultingMetrics = append(resultingMetrics, <-metricsCh)
	}
	assert.Equal(len(expectedMetrics), len(resultingMetrics))
	assert.Equal(convertMetrics(expectedMetrics), convertMetrics(resultingMetrics))
}

func getGivenMeasurements(value1 *float32) []*m.DiskMeasurements {
	return []*m.DiskMeasurements{
		{
			ProjectID:     "testProjectID",
			RsName:        "testReplicaSet",
			UserAlias:     "cluster-host:27017",
			PartitionName: "testPartition",
			Measurements: map[m.MeasurementID]*m.Measurement{
				"DISK_PARTITION_IOPS_READ_SCALAR_PER_SECOND": &m.Measurement{
					DataPoints: []*mongodbatlas.DataPoints{
						{
							Timestamp: "2017-08-22T20:31:12Z",
							Value:     nil,
						},
						{
							Timestamp: "2017-08-22T20:31:14Z",
							Value:     value1,
						},
					},
					Units: m.SCALAR_PER_SECOND,
				},
				"DISK_PARTITION_SPACE_USED_BYTES": &m.Measurement{
					DataPoints: []*mongodbatlas.DataPoints{},
					Units:      m.BYTES,
				},
			},
		},
	}
}

func getExpectedDisksMetrics(value float64) []prometheus.Metric {
	diskPartitionIopsReadRate := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "disks_stats", "disk_partition_iops_read_ratio"),
			"Original measurements.name: 'DISK_PARTITION_IOPS_READ'. "+defaultHelp,
			defaultDiskLabels,
			nil),
		prometheus.GaugeValue,
		value,
		"testProjectID", "testReplicaSet", "cluster-host:27017", "testPartition")
	diskUp := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, disksPrefix, "up"),
			upHelp,
			nil,
			nil),
		prometheus.GaugeValue,
		1)
	totalScrapes := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, disksPrefix, "scrapes_total"),
			totalScrapesHelp,
			nil,
			nil),
		prometheus.CounterValue,
		1)
	scrapeFailures := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, disksPrefix, "scrape_failures_total"),
			scrapeFailuresHelp,
			nil,
			nil),
		prometheus.CounterValue,
		3)
	measurementTransformationFailures := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, disksPrefix, "measurement_transformation_failures_total"),
			measurementTransformationFailuresHelp,
			nil,
			nil),
		prometheus.CounterValue,
		1)
	return []prometheus.Metric{diskPartitionIopsReadRate, diskUp, totalScrapes, scrapeFailures, measurementTransformationFailures}
}
