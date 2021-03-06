package collector

import (
	"mongodbatlas_exporter/measurer"
	"mongodbatlas_exporter/model"
	"os"
	"reflect"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

//Tuples of FQNAME and Metric help
//used by getExpectedDescs
var commontestExpectedDescs = [][]string{
	{prometheus.BuildFQName(namespace, processesPrefix, "up"), upHelp},
	{prometheus.BuildFQName(namespace, processesPrefix, "scrapes_total"), totalScrapesHelp},
	{prometheus.BuildFQName(namespace, processesPrefix, "scrape_failures_total"), scrapeFailuresHelp},
	//this one needs variable labels... anything past index 1 is a variable label.
	{prometheus.BuildFQName(namespace, processesPrefix, "measurement_transformation_failures_total"), measurementTransformationFailuresHelp, "atlas_metric", "error"},
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
		variableLabels := theDesc[2:]
		result[i] = prometheus.NewDesc(fqName, help, variableLabels, measurer.PromConstLabels())
	}

	return result
}

type MockClient struct {
	givenDisksMeasurements     map[model.MeasurementID]*model.Measurement
	givenProcessesMeasurements map[model.MeasurementID]*model.Measurement
}

type promTestMetric struct {
	FQNAME              string
	ConstLabels         prometheus.Labels
	VariableLabels      []string
	VariableLabelValues []string
	Value               float64
}

//convert metrics into a map of [fqname]map[]
func convertMetrics(metrics []prometheus.Metric) map[string]promTestMetric {
	result := make(map[string]promTestMetric, len(metrics))
	for _, metric := range metrics {
		metricValue := reflect.Indirect(reflect.ValueOf(metric))
		desc := reflect.Indirect(metricValue.FieldByName("desc"))
		fqName := desc.FieldByName("fqName").String()
		constLabelPairs := desc.FieldByName("constLabelPairs")
		labelPairs := metricValue.FieldByName("labelPairs")
		variableLabels := desc.FieldByName("variableLabels")

		//If metrics have variable labels you will have fqname collisions.
		//append the variable label values to make a unique fqname.
		//search all the label pairs
		for i := 0; i < labelPairs.Len(); i++ {
			labelPair := reflect.Indirect(labelPairs.Index(i))
			labelName := reflect.Indirect(labelPair.FieldByName("Name")).String()
			//for the variable labels
			for ii := 0; ii < variableLabels.Len(); ii++ {
				target := variableLabels.Index(ii).String()

				if labelName == target {
					fqName += reflect.Indirect(labelPair.FieldByName("Value")).String()
				}
			}

		}

		// for an fqName
		if _, ok := result[fqName]; !ok {
			valueValue := metricValue.FieldByName("val")
			testMetric := promTestMetric{
				FQNAME:      fqName,
				ConstLabels: make(prometheus.Labels, constLabelPairs.Len()),
			}
			result[fqName] = testMetric

			//TODO: There's an issue here if it's a measurement vec? Needs more looking into.
			if valueValue.IsValid() {
				testMetric.Value = valueValue.Float()
			}

			//put all label pairs into a map for easier comparison.
			for i := 0; i < constLabelPairs.Len(); i++ {
				name := constLabelPairs.Index(i).Elem().FieldByName("Name").Elem().String()
				value := constLabelPairs.Index(i).Elem().FieldByName("Value").Elem().String()
				result[fqName].ConstLabels[name] = value
			}
		}
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
func TestDescribe(t *testing.T) {
	assert := assert.New(t)
	mock := &MockClient{}
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	//For this test it is not important to populate the Process Measurer
	//Specific tests for child collectors of the basic collector should cover this.
	processMeasurer := measurer.Process{}

	collector, err := newBasicCollector(logger, mock, &processMeasurer, processesPrefix)
	assert.NotNil(collector)
	assert.NoError(err)

	testDescribe(t, collector, getExpectedDescs(&processMeasurer, commontestExpectedDescs))
}

//testDescribe is a common method for testing the descriptions returned by an exporter.
//expectedDescs are usually declared at the top of a file, but may be declared within a test
//function itself.
func testDescribe(t *testing.T, collector prometheus.Collector, expectedDescs []*prometheus.Desc) {
	descCh := make(chan *prometheus.Desc, 99)
	defer close(descCh)
	collector.Describe(descCh)
	var resultingDescs []*prometheus.Desc

	for len(descCh) > 0 {
		resultingDescs = append(resultingDescs, <-descCh)
	}

	expectedDescsMap := descsToLabelMaps(expectedDescs)
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
				assert.Equal(t, expectedLabels[label], actualLabels[label], "fqname: %s, label: %s", fqname, label)
			}
		} else {
			t.Fatalf("actual missing fqname %s", fqname)
		}
	}

	assert.Equal(t, len(expectedDescsMap), len(resultingDescsMap))
}
