package collector

import (
	m "mongodbatlas_exporter/model"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var defaultProcessLabels = []string{"project_id", "rs_name", "user_alias"}
var infoProcessLabels = []string{"project_id", "rs_name", "user_alias", "version", "type"}

const (
	processesPrefix = "processes_stats"
	infoHelp        = "Process info metric"
)

// Processes information struct
type Processes struct {
	*basicCollector
	info *prometheus.GaugeVec
}

// NewProcesses creates Process Prometheus metrics
func NewProcesses(logger log.Logger, client m.Client) (*Processes, error) {
	measurements, err := client.GetProcessMeasurementMap()
	if err != nil {
		return nil, err
	}

	basicCollector, err := newBasicCollector(logger, client, measurements, defaultProcessLabels, processesPrefix)
	if err != nil {
		return nil, err
	}
	processes := &Processes{
		basicCollector: basicCollector,
		info: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, processesPrefix, "info"),
				Help: infoHelp,
			},
			infoProcessLabels),
	}

	return processes, nil
}

// Collect implements prometheus.Collector.
func (c *Processes) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
		ch <- c.scrapeFailures
		ch <- c.measurementTransformationFailures
	}()

	processesMeasurements, failedScrapes, err := c.client.GetProcessMeasurements()
	if err != nil {
		c.up.Set(0)
	}
	c.up.Set(1)
	c.scrapeFailures.Add(float64(failedScrapes))

	//for every process
	for _, processMeasurements := range processesMeasurements {
		//every metric for every process
		for _, metric := range processMeasurements.Measurements {
			err := c.reportMeasurement(ch,
				metric,
				processMeasurements.ProjectID,
				processMeasurements.RsName,
				processMeasurements.UserAlias)

			if err != nil {
				level.Warn(c.logger).Log("err", err)
			}
		}

		infoGauge := c.info.WithLabelValues(
			processMeasurements.ProjectID,
			processMeasurements.RsName,
			processMeasurements.UserAlias,
			processMeasurements.Version,
			processMeasurements.TypeName)
		infoGauge.Set(1)
		ch <- infoGauge
	}
}

// Describe implements prometheus.Collector.
func (c *Processes) Describe(ch chan<- *prometheus.Desc) {
	c.basicCollector.Describe(ch)
	c.info.Describe(ch)
}
