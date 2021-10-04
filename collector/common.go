package collector

import (
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
	measurements                                                    []*m.Measurement
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

	for _, measurement := range measurementMap {
		collector.RegisterAtlasMetric(measurement)
	}

	return &collector, nil
}

func (collector *basicCollector) RegisterAtlasMetric(measurement *m.Measurement) {
	//append to what will be the basiccollector's list of metrics.
	collector.measurements = append(collector.measurements, measurement)
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
func (c *basicCollector) reportMeasurement(ch chan<- prometheus.Metric, measurementMap m.MeasurementMap, measurement *m.Measurement, extraLabels ...string) error {
	_, ok := measurementMap[measurement.ID()]
	if !ok {
		c.measurementTransformationFailures.Inc()
		measurementMap.RegisterMeasurement(measurement)
	}
	value, err := measurement.PromVal()
	if err != nil {
		c.measurementTransformationFailures.Inc()
		return fmt.Errorf("metric %s value transformation failure: %s", measurement.Name, err)
	}

	desc, err := measurement.PromDesc(c.namespace, c.prefix, c.defaultLabels)

	if err != nil {
		c.measurementTransformationFailures.Inc()
		return fmt.Errorf("metric %s description transformation failure: %s", measurement.Name, err)
	}

	ch <- prometheus.MustNewConstMetric(
		desc,
		measurement.PromType(),
		value,
		extraLabels...,
	)
	return nil
}
