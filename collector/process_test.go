package collector

import (
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

func (c *MockClient) ListProcesses() ([]*mongodbatlas.Process, *a.HTTPError) {
	return nil, nil
}

//TODO: this could be further refined and organized to drive table driven tests better.

//CAUTION: you need to make sure that that the API values returned from mocks_test.go are all represented here correctly.
//HELPWANTED: tie the mock API calls and the expected descriptions together better.
//Include the descriptions from the basic collector.
var processExpectedDescs = append(commontestExpectedDescs,
	//Process Specific Metric Descriptions
	[]string{prometheus.BuildFQName(namespace, processesPrefix, "info"), infoHelp},
	[]string{prometheus.BuildFQName(namespace, processesPrefix, "query_executor_scanned_ratio"), "Original measurements.name: 'QUERY_EXECUTOR_SCANNED'. " + measurer.DEFAULT_HELP},
	[]string{prometheus.BuildFQName(namespace, processesPrefix, "tickets_available_reads"), "Original measurements.name: 'TICKETS_AVAILABLE_READS'. " + measurer.DEFAULT_HELP},
)

var diskExpectedDescs = [][]string{
	//This disk metric should be attached to the sub-resource for disks on the process measurer
	{prometheus.BuildFQName(namespace, disksPrefix, "disk_partition_iops_read_ratio"), "Original measurements.name: 'DISK_PARTITION_IOPS_READ'. " + measurer.DEFAULT_HELP},
	{prometheus.BuildFQName(namespace, disksPrefix, "disk_partition_space_used_bytes"), "Original measurements.name: 'DISK_PARTITION_SPACE_USED'. " + measurer.DEFAULT_HELP},
}

var testAtlasProcess = mongodbatlas.Process{
	GroupID:        "testProjectID",
	ReplicaSetName: "testReplicaSet",
	UserAlias:      "cluster-host:27017",
	TypeName:       "REPLICA_PRIMARY",
	Version:        "4.2.13",
}
var testProcessMeasurements = measurer.ProcessFromMongodbAtlasProcess(&testAtlasProcess)

//TestProcessDescribe extends TestDescribe found in common_test.go
//The extension occurs because Process needs to describe the process
//metrics _and_ the disk metrics. The basic collector would not handle
//the disk describes.
func TestProcessDescribe(t *testing.T) {
	process := mongodbatlas.Process{
		Port: 2017,
	}
	mock := &MockClient{}
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	//make descriptions for process metrics
	processMeasurer := measurer.ProcessFromMongodbAtlasProcess(&process)
	processExpectedDescs := getExpectedDescs(processMeasurer, append(commontestExpectedDescs, processExpectedDescs...))

	//make descriptions for disk metrics
	disks, httpErr := mock.ListDisks(&process)

	assert.Nil(t, httpErr)

	diskMeasurer := measurer.DiskFromMongodbAtlasProcessDisk(&process, disks[0])

	allExpectedDecs := append(processExpectedDescs, getExpectedDescs(diskMeasurer, diskExpectedDescs)...)

	processCollector, err := NewProcessCollector(logger, mock, &process)

	assert.NoError(t, err)
	assert.NotNil(t, processCollector)

	testDescribe(t, processCollector, allExpectedDecs)
}

func TestProcessesCollector(t *testing.T) {
	assert := assert.New(t)
	value := float32(1.0499)
	mock := &MockClient{}
	mock.givenProcessesMeasurements = getGivenProcessesMeasurements(&value)
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	processCollector, err := NewProcessCollector(logger, mock, &testAtlasProcess)
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
	testProcessMeasurements.Measurements = map[m.MeasurementID]*m.Measurement{
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
	}
	return []*measurer.Process{
		testProcessMeasurements,
	}
}

type metricInput struct {
	fqName              string
	help                string
	variableLabels      []string
	variableLabelValues []string
	value               float64
}

func getExpectedProcessesMetrics(value float64) []prometheus.Metric {
	inputs := []metricInput{
		{
			fqName: prometheus.BuildFQName(namespace, processesPrefix, "query_executor_scanned_ratio"),
			help:   "Original measurements.name: 'QUERY_EXECUTOR_SCANNED'. " + measurer.DEFAULT_HELP,
			value:  value,
		},
		{
			fqName: prometheus.BuildFQName(namespace, processesPrefix, "up"),
			help:   upHelp,
			value:  1,
		},
		{
			fqName: prometheus.BuildFQName(namespace, processesPrefix, "scrapes_total"),
			help:   totalScrapesHelp,
			value:  1,
		},
		{
			fqName: prometheus.BuildFQName(namespace, processesPrefix, "scrape_failures_total"),
			help:   scrapeFailuresHelp,
			value:  3,
		},
		{
			fqName:              prometheus.BuildFQName(namespace, processesPrefix, "measurement_transformation_failures_total"),
			help:                measurementTransformationFailuresHelp,
			variableLabels:      []string{"atlas_metric", "error"},
			variableLabelValues: []string{"TICKETS_AVAILABLE_READS", "no_data"},
			value:               1,
		},
		{
			fqName: prometheus.BuildFQName(namespace, processesPrefix, "info"),
			help:   infoHelp,
			value:  1,
		},
	}

	expectedMetrics := make([]prometheus.Metric, len(inputs))

	for i := range inputs {
		desc := prometheus.NewDesc(
			inputs[i].fqName,
			inputs[i].help,
			inputs[i].variableLabels,
			testProcessMeasurements.PromConstLabels(),
		)
		expectedMetrics[i] = prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			inputs[i].value,
			inputs[i].variableLabelValues...,
		)
	}
	return expectedMetrics
}
