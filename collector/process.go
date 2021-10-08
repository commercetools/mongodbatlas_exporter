package collector

import (
	"mongodbatlas_exporter/model"
	a "mongodbatlas_exporter/mongodbatlas"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

// Process information struct
type Process struct {
	*basicCollector
	info     prometheus.Gauge
	measurer model.ProcessMeasurements
}

func NewProcessCollector(logger log.Logger, client a.Client, p *mongodbatlas.Process) (*Process, error) {
	metadata, httpErr := client.GetProcessMeasurementsMetadata(p)
	if httpErr != nil {
		return nil, httpErr
	}
	processMetadata := model.ProcessMeasurements{
		ProjectID: p.GroupID,
		RsName:    p.ReplicaSetName,
		UserAlias: p.UserAlias,
		Version:   p.Version,
		TypeName:  p.TypeName,
		Hostname:  p.Hostname,
		Port:      p.Port,
		ID:        p.ID,
	}

	basicCollector, err := newBasicCollector(logger, client, metadata, &processMetadata, processesPrefix)

	if err != nil {
		return nil, err
	}

	process := &Process{
		basicCollector: basicCollector,
		info: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        prometheus.BuildFQName(namespace, processesPrefix, "info"),
				Help:        infoHelp,
				ConstLabels: processMetadata.PromConstLabels(),
			}),
		measurer: processMetadata,
	}

	return process, nil
}

func (c *Process) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
		ch <- c.scrapeFailures
	}()

	processMeasurements, err := c.client.GetProcessMeasurements(c.measurer)

	if err != nil {
		x := err.Error()
		level.Debug(c.logger).Log("msg", "scrape failure", "err", err, "x", x)
		c.scrapeFailures.Inc()
		c.up.Set(0)
	}
	c.up.Set(1)

	c.measurer.Measurements = processMeasurements

	for _, metric := range c.metrics {
		err = c.report(&c.measurer, metric, ch)
		if err != nil {
			level.Debug(c.logger).Log("msg", "skipping metric", "metric", metric.Desc,
				"err", err)
		}
	}

	c.info.Set(1)
	ch <- c.info
}

// Describe implements prometheus.Collector.
func (c *Process) Describe(ch chan<- *prometheus.Desc) {
	c.basicCollector.Describe(ch)
	c.info.Describe(ch)
}
