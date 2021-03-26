package collector

import (
	"errors"
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
	atlasScrapeFailuresHelp               = "Number of unsuccessful measurement scrapes from MongoDB Atlas API."
	defaultHelp                           = "Please see MongoDB Atlas documentation for details about the measurement"
	measurementTransformationFailuresHelp = "Number of errors during transformation of scraped MongoDB Atlas measurements into Prometheus metrics."
)

type metric struct {
	Type     prometheus.ValueType
	Desc     *prometheus.Desc
	Metadata *m.MeasurementMetadata
}

type basicCollector struct {
	client m.Client
	logger log.Logger

	up                                                                   prometheus.Gauge
	totalScrapes, atlasScrapeFailures, measurementTransformationFailures prometheus.Counter
	metrics                                                              []*metric
}

// newBasicCollector creates basicCollector
func newBasicCollector(logger log.Logger, client m.Client, measurementsMetadata []*m.MeasurementMetadata, defaultLabels []string, collectorPrefix string) (*basicCollector, error) {
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
				defaultLabels, nil,
			),
			Metadata: measurementMetadata,
		}
		metrics = append(metrics, &metric)
	}

	return &basicCollector{
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "up"),
			Help: upHelp,
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "total_scrapes"),
			Help: totalScrapesHelp,
		}),
		atlasScrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "atlas_scrape_failures"),
			Help: atlasScrapeFailuresHelp,
		}),
		measurementTransformationFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, collectorPrefix, "measurement_transformation_failures"),
			Help: measurementTransformationFailuresHelp,
		}),
		metrics: metrics,

		client: client,
		logger: logger,
	}, nil
}

// Describe implements prometheus.Collector.
func (c *basicCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.atlasScrapeFailures.Desc()
	ch <- c.measurementTransformationFailures.Desc()

	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}
