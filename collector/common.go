package collector

import (
	"fmt"
	transformer "mongodbatlas_exporter/collector/transformer"
	"mongodbatlas_exporter/measurer"
	m "mongodbatlas_exporter/model"
	a "mongodbatlas_exporter/mongodbatlas"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "mongodbatlas"

	upHelp                                = "Was the last communication with MongoDB Atlas API successful."
	totalScrapesHelp                      = "Current total MongoDB Atlas scrapes."
	scrapeFailuresHelp                    = "Number of unsuccessful measurement scrapes from MongoDB Atlas API."
	defaultHelp                           = "Please see MongoDB Atlas documentation for details about the measurement"
	measurementTransformationFailuresHelp = "Number of errors during transformation of scraped MongoDB Atlas measurements into Prometheus metrics."
)

type metric struct {
	Type     prometheus.ValueType
	Desc     *prometheus.Desc
	Metadata *m.MeasurementMetadata
}

//ErrorLabels consumes prometheus.Labels and adds more labels to the map.
//Perhaps this is a chainable pattern we can reuse on other types to have a
//consistent interface for working with labels.
//The prometheus API is fairly inconcsistent where many APIs require a slice of
//label names or label values.
//ErrorLabels original need was to combine labels from a Measurer and use them
//with a prometheus.CounterVec to select a particular counter.
func (x *metric) ErrorLabels(extraLabels prometheus.Labels) prometheus.Labels {
	result := prometheus.Labels{
		"atlas_metric": x.Metadata.Name,
	}

	for key, value := range extraLabels {
		result[key] = value
	}

	return result
}

type basicCollector struct {
	client a.Client
	logger log.Logger

	up                                prometheus.Gauge
	totalScrapes, scrapeFailures      prometheus.Counter
	measurementTransformationFailures prometheus.CounterVec
	metrics                           []*metric
}

//metadataToMetric transforms the measurement metadata we received from Atlas into a
//prometheus compatible metric description.
func metadataToMetric(metadata *m.MeasurementMetadata, collectorPrefix string, variableLabels []string, constLabels prometheus.Labels) (*metric, error) {
	promName, err := transformer.TransformName(metadata)
	if err != nil {
		msg := "can't transform measurement Name (%s) into metric name"
		return nil, fmt.Errorf(msg, metadata.Name)
	}
	promType, err := transformer.TransformType(metadata)
	if err != nil {
		msg := "can't transform measurement Units (%s) into prometheus.ValueType"
		return nil, fmt.Errorf(msg, metadata.Units)
	}

	metric := metric{
		Type: promType,
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, collectorPrefix, promName),
			"Original measurements.name: '"+metadata.Name+"'. "+defaultHelp,
			variableLabels, constLabels,
		),
		Metadata: metadata,
	}

	return &metric, nil
}

// newBasicCollector creates basicCollector
func newBasicCollector(logger log.Logger, client a.Client, measurer measurer.Measurer, collectorPrefix string) (*basicCollector, error) {
	var metrics []*metric
	for _, measurementMetadata := range measurer.GetMetaData() {
		metric, err := metadataToMetric(measurementMetadata, collectorPrefix, nil, measurer.PromConstLabels())
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}
	failureLabels := []string{"atlas_metric", "error"}

	return &basicCollector{
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        prometheus.BuildFQName(namespace, collectorPrefix, "up"),
			Help:        upHelp,
			ConstLabels: measurer.PromConstLabels(),
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, collectorPrefix, "scrapes_total"),
			Help:        totalScrapesHelp,
			ConstLabels: measurer.PromConstLabels(),
		}),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, collectorPrefix, "scrape_failures_total"),
			Help:        scrapeFailuresHelp,
			ConstLabels: measurer.PromConstLabels(),
		}),
		measurementTransformationFailures: *prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, collectorPrefix, "measurement_transformation_failures_total"),
			Help:        measurementTransformationFailuresHelp,
			ConstLabels: measurer.PromConstLabels(),
		}, failureLabels),
		metrics: metrics,
		client:  client,
		logger:  logger,
	}, nil
}

// Describe implements prometheus.Collector.
func (c *basicCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.scrapeFailures.Desc()

	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}

//report may make more sense being renamed Collect since it is really the parent method
//for Collect on any sub-type of basicCollecor.
//Any Measurer and metric can use this common report function to send metrics to prometheus.
//ProcessCollector and DiskCollector have a small number of particular metrics that they must report
//themselves using their derivative implementation of Collect.
//Another nice facet of "report" is that it communicates meaning using errors rather than logs. The meaning
//can be interpreted within the program as well as by operators.
func (c *basicCollector) report(measurer measurer.Measurer, metric *metric, ch chan<- prometheus.Metric) error {
	measurement, ok := measurer.GetMeasurements()[metric.Metadata.ID()]
	baseErrorLabels := metric.ErrorLabels(prometheus.Labels{})

	//This case is different from no_data because it indicates
	//that the measurement does not exist at all.
	//This often occurs with oplog delay metrics on primaries
	//because they do not report that metric as only secondaries have
	//replication delay.
	if !ok {
		baseErrorLabels["error"] = "not_found"
		notRegistered := c.measurementTransformationFailures.With(baseErrorLabels)
		notRegistered.Inc()
		ch <- notRegistered
		return fmt.Errorf("instance has no measurement %s", metric.Metadata.Name)
	}
	value, err := transformer.TransformValue(measurement)
	//exposing different value transformation errors as metrics.
	//this is a nice example of using errors with switch statements
	if err != nil {
		switch err {
		//When a Measurement exists, but has no datapoints.
		case transformer.ErrNoData:
			baseErrorLabels["error"] = "no_data"
		default:
			baseErrorLabels["error"] = "value"
		}

		valueError := c.measurementTransformationFailures.With(baseErrorLabels)
		valueError.Inc()
		ch <- valueError
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		metric.Desc,
		metric.Type,
		value,
		measurer.PromVariableLabelValues()...,
	)
	return nil
}
