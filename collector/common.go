package collector

import (
	"errors"
	"fmt"
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
	measurementTransformationFailuresHelp = "Number of errors during transformation of scraped MongoDB Atlas measurements into Prometheus metrics."
)

type basicCollector struct {
	client                                                          m.Client
	logger                                                          log.Logger
	namespace                                                       string
	prefix                                                          string
	defaultLabels                                                   []string
	up                                                              prometheus.Gauge
	totalScrapes, scrapeFailures, measurementTransformationFailures prometheus.Counter

	// the MeasurementMap should hold all the metrics for the particular collector.
	// since all metrics are currently reported as Gauges, this is not a stateful representation
	// of metric current values, but a known definition of the available metrics.
	measurements m.MeasurementMap
}

// newBasicCollector creates basicCollector
func newBasicCollector(logger log.Logger, client m.Client, measurementMap m.MeasurementMap, defaultLabels []string, collectorPrefix string) (*basicCollector, error) {
	collector := basicCollector{
		prefix:        collectorPrefix,
		defaultLabels: defaultLabels,
		namespace:     namespace,
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
		measurementTransformationFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "measurement_transformation_failures_total"),
			Help: measurementTransformationFailuresHelp,
		}),
		client: client,
		logger: logger,
	}

	collector.measurements = make(m.MeasurementMap, len(measurementMap))
	for _, measurement := range measurementMap {
		collector.measurements.RegisterMeasurement(measurement)
	}

	return &collector, nil
}

// Describe implements prometheus.Collector.
func (c *basicCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.scrapeFailures.Desc()
	ch <- c.measurementTransformationFailures.Desc()

	for _, measurement := range c.measurements {
		desc, err := measurement.PromDesc(c.namespace, c.prefix, c.defaultLabels)

		if err != nil {
			level.Warn(c.logger).Log("msg", "skipping metric because description transformation error", "metric", measurement.Name, "measurement", measurement, "err", err)
			continue
		}
		ch <- desc
	}
}

//reportMeasurement is used during the collector.Collect call to convert a model.Measurement into the necessary data for the prometheus report.
func (c *basicCollector) reportMeasurement(ch chan<- prometheus.Metric, measurement *m.Measurement, extraLabels ...string) error {
	_, ok := c.measurements[measurement.ID()]
	if !ok {
		c.measurementTransformationFailures.Inc()
		c.measurements.RegisterMeasurement(measurement)
	}

	desc, err := measurement.PromDesc(c.namespace, c.prefix, c.defaultLabels)

	if err != nil {
		c.measurementTransformationFailures.Inc()
		return fmt.Errorf("metric %s description transformation failure: %s", measurement.Name, err)
	}

	value, err := measurement.PromVal()
	if err != nil {
		// If there are no datapoints we do not count it as a transformation nor a scrape error.
		// It's just missing data. Not sure if this is the _best_, but I'm trying to maintain backwards compatibility.
		if errors.Is(err, m.ErrNoDatapoints) {
			return err
		}
		c.measurementTransformationFailures.Inc()
		err = fmt.Errorf("metric %s value transformation failure: %s", measurement.Name, err)
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		desc,
		measurement.PromType(),
		value,
		extraLabels...,
	)
	return err
}
