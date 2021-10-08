package collector

import (
	"errors"
	"fmt"
	transformer "mongodbatlas_exporter/collector/transformer"
	m "mongodbatlas_exporter/model"
	a "mongodbatlas_exporter/mongodbatlas"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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

// newBasicCollector creates basicCollector
func newBasicCollector(logger log.Logger, client a.Client, measurementsMetadata map[m.MeasurementID]*m.MeasurementMetadata, measurer m.Measurer, collectorPrefix string) (*basicCollector, error) {
	var metrics []*metric
	count := 0
	for _, measurementMetadata := range measurementsMetadata {
		if count == 3 {
			break
		}
		count += 1
		promName, err := transformer.TransformName(measurementMetadata)
		if err != nil {
			msg := "can't transform measurement Name into metric name"
			level.Error(logger).Log("msg", msg, "measurementMetadata", measurementMetadata, "err", err)
			return nil, errors.New(msg)
		}
		promType, err := transformer.TransformType(measurementMetadata)
		if err != nil {
			msg := "can't transform measurement Units into prometheus.ValueType"
			level.Error(logger).Log("msg", msg, "measurementMetadata", measurementMetadata, "err", err)
			return nil, errors.New(msg)
		}

		metric := metric{
			Type: promType,
			Desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, collectorPrefix, promName),
				"Original measurements.name: '"+measurementMetadata.Name+"'. "+defaultHelp,
				nil, measurer.PromConstLabels(),
			),
			Metadata: measurementMetadata,
		}

		metrics = append(metrics, &metric)
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
func (c *basicCollector) report(measurer m.Measurer, metric *metric, ch chan<- prometheus.Metric) error {
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
	)
	return nil
}
