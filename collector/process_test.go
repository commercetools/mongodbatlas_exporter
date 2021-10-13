package collector

import (
	"errors"
	"mongodbatlas_exporter/measurer"
	m "mongodbatlas_exporter/model"
	a "mongodbatlas_exporter/mongodbatlas"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas/mongodbatlas"
)

func (c *MockClient) GetProcessMeasurements(_ measurer.Process) (map[m.MeasurementID]*m.Measurement, error) {
	return make(map[m.MeasurementID]*m.Measurement), nil
}

func (c *MockClient) ListDisks(*mongodbatlas.Process) ([]*mongodbatlas.ProcessDisk, *a.HTTPError) {
	return nil, &a.HTTPError{
		StatusCode: 404,
		Err:        errors.New("not found"),
	}
}

func (c *MockClient) GetProcessMeasurementsMetadata(p *measurer.Process) *a.HTTPError {
	p.Metadata = map[m.MeasurementID]*m.MeasurementMetadata{
		m.NewMeasurementID("TICKETS_AVAILABLE_READS", "SCALAR"): {
			Name:  "TICKETS_AVAILABLE_READS",
			Units: "SCALAR",
		},
		m.NewMeasurementID("QUERY_EXECUTOR_SCANNED", "SCALAR_PER_SECOND"): {
			Name:  "QUERY_EXECUTOR_SCANNED",
			Units: "SCALAR_PER_SECOND",
		},
	}
	return nil
}

func (c *MockClient) ListProcesses() ([]*mongodbatlas.Process, *a.HTTPError) {
	return nil, nil
}

func TestProcessesCollector(t *testing.T) {
	assert := assert.New(t)
	value := float32(1.0499)
	mock := &MockClient{}
	mock.givenProcessesMeasurements = getGivenProcessesMeasurements(&value)
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	process := mongodbatlas.Process{}

	processCollector, err := NewProcessCollector(logger, mock, &process)
	assert.NotNil(processCollector)
	assert.NoError(err)

	metricsCh := make(chan prometheus.Metric, 99)
	defer close(metricsCh)
	expectedMetrics := getExpectedProcessesMetrics(float64(value))
	processCollector.Collect(metricsCh)
	var resultingMetrics []prometheus.Metric
	for len(metricsCh) > 0 {
		resultingMetrics = append(resultingMetrics, <-metricsCh)
	}
	assert.Equal(len(expectedMetrics), len(resultingMetrics))
	assert.Equal(convertMetrics(expectedMetrics), convertMetrics(resultingMetrics))
}

func getGivenProcessesMeasurements(value1 *float32) []*measurer.Process {
	return []*measurer.Process{
		{
			ProjectID: "testProjectID",
			RsName:    "testReplicaSet",
			UserAlias: "cluster-host:27017",
			Version:   "4.2.13",
			TypeName:  "REPLICA_PRIMARY",
			Measurements: map[m.MeasurementID]*m.Measurement{
				"QUERY_EXECUTOR_SCANNED_SCALAR_PER_SECOND": {
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
				"TICKETS_AVAILABLE_READS_SCALAR": {
					DataPoints: []*mongodbatlas.DataPoints{},
					Units:      m.SCALAR,
				},
			},
		},
	}
}

func getExpectedProcessesMetrics(value float64) []prometheus.Metric {
	processMeasurements := measurer.Process{
		ProjectID: "testProjectID",
		RsName:    "testReplicaSet",
		UserAlias: "cluster-host:27017",
		Version:   "4.2.13",
		TypeName:  "REPLICA_PRIMARY",
	}
	processQueryExecutorScanned := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "query_executor_scanned_ratio"),
			"Original measurements.name: 'QUERY_EXECUTOR_SCANNED'. "+defaultHelp,
			processMeasurements.PromVariableLabelNames(),
			nil),
		prometheus.GaugeValue,
		value,
		processMeasurements.PromVariableLabelValues()...)
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
			append((&measurer.Process{}).PromVariableLabelNames(), "atlas_metric", "error"),
			nil),
		prometheus.CounterValue,
		1,
		append(processMeasurements.PromVariableLabelValues(), "TICKETS_AVAILABLE_READS", "no_data")...,
	)
	processInfo := prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, processesPrefix, "info"),
			infoHelp,
			processMeasurements.PromVariableLabelNames(),
			processMeasurements.PromConstLabels()),
		prometheus.GaugeValue,
		1,
		processMeasurements.PromVariableLabelValues()...)
	return []prometheus.Metric{processQueryExecutorScanned, processUp, totalScrapes, scrapeFailures, processInfo, measurementTransformationFailures}
}
