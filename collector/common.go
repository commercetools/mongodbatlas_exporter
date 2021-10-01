package collector

import (
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
	Type        prometheus.ValueType
	Desc        *prometheus.Desc
	Measurement *m.Measurement
}

type basicCollector struct {
	client                                                          m.Client
	logger                                                          log.Logger
	prefix                                                          string
	defaultLabels                                                   []string
	up                                                              prometheus.Gauge
	totalScrapes, scrapeFailures, measurementTransformationFailures prometheus.Counter
	metrics                                                         []*metric
}

// newBasicCollector creates basicCollector
func newBasicCollector(logger log.Logger, client m.Client, measurementMap m.MeasurementMap, defaultLabels []string, collectorPrefix string) (*basicCollector, error) {
	collector := basicCollector{
		prefix:        collectorPrefix,
		defaultLabels: defaultLabels,
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
		err := RegisterAtlasMetric(measurement, &collector)

		if err != nil {
			level.Error(logger).Log("err", err)
		}
	}

	return &collector, nil
}

func RegisterAtlasMetric(measurement *m.Measurement, collector *basicCollector) error {
	// Transforms the Atlas name to a Prometheus Name
	promName, err := transformer.TransformName(measurement)
	if err != nil {
		msg := "can't transform measurement Name (%s) into metric name: %s"
		return fmt.Errorf(msg, measurement.Name, err.Error())
	}

	// Transforms the Atlas type to a Prometheus Type
	promType, err := transformer.TransformType(measurement)
	if err != nil {
		msg := "can't transform measurement (%s) Units (%s) into prometheus.ValueType: %s"
		return fmt.Errorf(msg, measurement.Name, measurement.Units, err.Error())
	}

	// Defines a prometheus metric using the name, type and description.
	metric := metric{
		Type: promType,
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, collector.prefix, promName),
			"Original measurements.name: '"+measurement.Name+"'. "+defaultHelp,
			collector.defaultLabels, nil,
		),
		Measurement: measurement,
	}
	//append to what will be the basiccollector's list of metrics.
	collector.metrics = append(collector.metrics, &metric)
	return nil
}

// Describe implements prometheus.Collector.
func (c *basicCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.scrapeFailures.Desc()
	ch <- c.measurementTransformationFailures.Desc()

	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}
