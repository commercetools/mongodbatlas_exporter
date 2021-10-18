package collector

import (
	"mongodbatlas_exporter/measurer"
	"mongodbatlas_exporter/model"
	"os"
	"reflect"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

const (
	testPrefix = "stats"
)

var commontestExpectedDescs = [][]string{
	{prometheus.BuildFQName(namespace, testPrefix, "up"), upHelp},
	{prometheus.BuildFQName(namespace, testPrefix, "scrapes_total"), totalScrapesHelp},
	{prometheus.BuildFQName(namespace, testPrefix, "scrape_failures_total"), scrapeFailuresHelp},
	{prometheus.BuildFQName(namespace, testPrefix, "disk_partition_iops_read_ratio"), "Original measurements.name: 'DISK_PARTITION_IOPS_READ'. " + measurer.DEFAULT_HELP},
}

//getExpectedDescs is a mocking function to return the expected list of descriptions for a basic collector.
//In this instance the basicCollector is for disks since we are using the constant labels from model.DiskMeasurements
func getExpectedDescs(measurer measurer.Measurer, expected [][]string) []*prometheus.Desc {
	result := make([]*prometheus.Desc, len(expected))

	for i := range expected {
		//Build the description and add the constant labels. Constant labels are used to uniquely identify a measurement.
		//Whereas variable lables such as HTTP Status Codes provide more context.
		theDesc := expected[i]
		fqName := theDesc[0]
		help := theDesc[1]
		result[i] = prometheus.NewDesc(fqName, help, nil, measurer.PromConstLabels())
	}

	return result
}

type MockClient struct {
	givenDisksMeasurements     map[model.MeasurementID]*model.Measurement
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

//converts descs into a map of the [fqname]map[constLabels]labelValue
func descsToLabelMaps(descs []*prometheus.Desc) map[string]map[string]string {
	result := make(map[string]map[string]string, len(descs))
	for i := range descs {
		value := reflect.ValueOf(*descs[i])
		fqName := value.FieldByName("fqName").String()
		//[]*dto.LabelPair
		constLabelPairs := value.FieldByName("constLabelPairs")

		if !constLabelPairs.IsValid() {
			return result
		}

		// for an fqName
		if _, ok := result[fqName]; !ok {
			result[fqName] = make(map[string]string, constLabelPairs.Len())
			for i := 0; i < constLabelPairs.Len(); i++ {
				//iterate over each constLabel pair
				//*dto.LabelPair
				name := constLabelPairs.Index(i).Elem().FieldByName("Name").Elem().String()
				value := constLabelPairs.Index(i).Elem().FieldByName("Value").Elem().String()
				result[fqName][name] = value
			}
		}
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
func TestDescribe(t *testing.T) {
	assert := assert.New(t)
	mock := &MockClient{}
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	processMeasurer := measurer.Process{
		Base: measurer.Base{
			Metadata: getExpectedDiskMeasurementMetadata(),
		},
		Disks: []*measurer.Disk{
			{
				Base: measurer.Base{
					Metadata: getExpectedDiskMeasurementMetadata(),
				},
			},
		},
	}

	collector, err := newBasicCollector(logger, mock, &processMeasurer, testPrefix)
	assert.NotNil(collector)
	assert.NoError(err)
	descCh := make(chan *prometheus.Desc, 99)
	defer close(descCh)
	collector.Describe(descCh)
	var resultingDescs []*prometheus.Desc

	for len(descCh) > 0 {
		resultingDescs = append(resultingDescs, <-descCh)
	}
	expectedDescsMap := descsToLabelMaps(getExpectedDescs(&processMeasurer, commontestExpectedDescs))
	resultingDescsMap := descsToLabelMaps(resultingDescs)

	//It is very important to check that we are maintaining
	//consistency with the available constant labels and the
	//available variable labels. This is checked here.
	//for each (fqname, labels)
	for fqname, expectedLabels := range expectedDescsMap {
		//check that the resultingDesc had those labels
		if actualLabels, ok := resultingDescsMap[fqname]; ok {
			for label := range expectedLabels {
				//assert that the FQNAME has the same labels and values.
				assert.Equal(expectedLabels[label], actualLabels[label])
			}
		} else {
			t.Fatalf("actual missing fqname %s", fqname)
		}
	}

	assert.Equal(len(expectedDescsMap), len(resultingDescsMap))
}
