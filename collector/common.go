package collector

import (
	"errors"
	"fmt"
	transformer "mongodbatlas_exporter/collector/transformer"
	m "mongodbatlas_exporter/model"

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
	client m.Client
	logger log.Logger

	up                                prometheus.Gauge
	totalScrapes, scrapeFailures      prometheus.Counter
	measurementTransformationFailures prometheus.CounterVec
	metrics                           []*metric
}

// newBasicCollector creates basicCollector
func newBasicCollector(logger log.Logger, client m.Client, measurementsMetadata map[m.MeasurementID]*m.MeasurementMetadata, measurer m.Measurer, collectorPrefix string) (*basicCollector, error) {
	var metrics []*metric
	for _, measurementMetadata := range measurementsMetadata {
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
				measurer.LabelNames(), nil,
			),
			Metadata: measurementMetadata,
		}
		metrics = append(metrics, &metric)
	}
	failureLabels := append(measurer.LabelNames(), "atlas_metric", "error")

	return &basicCollector{
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "up"),
			Help: upHelp,
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "scrapes_total"),
			Help: totalScrapesHelp,
		}),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "scrape_failures_total"),
			Help: scrapeFailuresHelp,
		}),
		measurementTransformationFailures: *prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "measurement_transformation_failures_total"),
			Help: measurementTransformationFailuresHelp,
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

func (c *basicCollector) report(measurer m.Measurer, metric *metric, ch chan<- prometheus.Metric) error {
	measurement, ok := measurer.GetMeasurements()[metric.Metadata.ID()]
	baseErrorLabels := metric.ErrorLabels(measurer.PromLabels())
	if !ok {
		baseErrorLabels["error"] = "not_registered"
		notRegistered := c.measurementTransformationFailures.With(baseErrorLabels)
		notRegistered.Inc()
		ch <- notRegistered
		return fmt.Errorf("no registered measurement for %s", metric.Metadata.Name)
	}
	value, err := transformer.TransformValue(measurement)
	if err != nil {
		switch err {
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
		measurer.LabelValues()...,
	)
	return nil
}
