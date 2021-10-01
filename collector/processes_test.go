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

func (c *MockClient) GetProcessMeasurements() ([]*m.ProcessMeasurements, m.ScrapeFailures, error) {
	return c.givenProcessesMeasurements, 3, nil
}
func (c *MockClient) GetProcessMeasurementMap() (m.MeasurementMap, error) {
	return m.MeasurementMap{
		m.NewMeasurementID("TICKETS_AVAILABLE_READS", "SCALAR"): {
			Name:  "TICKETS_AVAILABLE_READS",
			Units: "SCALAR",
		},
		m.NewMeasurementID("QUERY_EXECUTOR_SCANNED", "SCALAR_PER_SECOND"): {
			Name:  "QUERY_EXECUTOR_SCANNED",
			Units: "SCALAR_PER_SECOND",
		},
	}, nil
}

func TestProcessesCollector(t *testing.T) {
	assert := assert.New(t)
	value := float32(1.0499)
	mock := &MockClient{}
	mock.givenProcessesMeasurements = getGivenProcessesMeasurements(&value)
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	processes, err := NewProcesses(logger, mock)
	assert.NotNil(processes)
	assert.NoError(err)

	metricsCh := make(chan prometheus.Metric, 99)
	defer close(metricsCh)
	expectedMetrics := getExpectedProcessesMetrics(float64(value))
	processes.Collect(metricsCh)
	var resultingMetrics []prometheus.Metric
	for len(metricsCh) > 0 {
		resultingMetrics = append(resultingMetrics, <-metricsCh)
	}
	assert.Equal(len(expectedMetrics), len(resultingMetrics))
	assert.Equal(convertMetrics(expectedMetrics), convertMetrics(resultingMetrics))
}

func getGivenProcessesMeasurements(value1 *float32) []*m.ProcessMeasurements {
	return []*m.ProcessMeasurements{
		{
			ProjectID: "testProjectID",
			RsName:    "testReplicaSet",
			UserAlias: "cluster-host:27017",
			Version:   "4.2.13",
			TypeName:  "REPLICA_PRIMARY",
			Measurements: map[m.MeasurementID]*m.Measurement{
				"QUERY_EXECUTOR_SCANNED_SCALAR_PER_SECOND": &m.Measurement{
					DataPoints: []*mongodbatlas.DataPoints{
						{
							Timestamp: "2021-03-07T15:46:13Z",
							Value:     nil,
						},
						{
							Timestamp: "2021-03-07T15:47:13Z",
							Value:     value1,
						},
					},
					Units: m.SCALAR_PER_SECOND,
				},
				"TICKETS_AVAILABLE_READS_SCALAR": &m.Measurement{
					DataPoints: []*mongodbatlas.DataPoints{},
					Units:      m.SCALAR,
				},
			},
		},
	}
}

func getExpectedProcessesMetrics(value float64) []prometheus.Metric {
	processQueryExecutorScanned := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "query_executor_scanned_ratio"),
			"Original measurements.name: 'QUERY_EXECUTOR_SCANNED'. "+defaultHelp,
			defaultProcessLabels,
			nil),
		prometheus.GaugeValue,
		value,
		"testProjectID", "testReplicaSet", "cluster-host:27017")
	processUp := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "up"),
			upHelp,
			nil,
			nil),
		prometheus.GaugeValue,
		1)
	totalScrapes := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "scrapes_total"),
			totalScrapesHelp,
			nil,
			nil),
		prometheus.CounterValue,
		1)
	scrapeFailures := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "scrape_failures_total"),
			scrapeFailuresHelp,
			nil,
			nil),
		prometheus.CounterValue,
		3)
	measurementTransformationFailures := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "measurement_transformation_failures_total"),
			measurementTransformationFailuresHelp,
			nil,
			nil),
		prometheus.CounterValue,
		1)
	processInfo := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "info"),
			infoHelp,
			infoProcessLabels,
			nil),
		prometheus.GaugeValue,
		1,
		"testProjectID", "testReplicaSet", "cluster-host:27017", "4.2.13", "REPLICA_PRIMARY")
	return []prometheus.Metric{processQueryExecutorScanned, processUp, totalScrapes, scrapeFailures, processInfo, measurementTransformationFailures}
}
