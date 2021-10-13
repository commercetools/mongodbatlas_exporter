package collector

import (
	"mongodbatlas_exporter/measurer"
	a "mongodbatlas_exporter/mongodbatlas"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const disksPrefix = "disks_stats"

// Disks information struct
type Disks struct {
	*basicCollector
}

// NewDisks creates Disk Prometheus metrics
func NewDisks(logger log.Logger, client a.Client) (*Disks, error) {
	measurementsMetadata, err := client.GetDiskMeasurementsMetadata()
	if err != nil {
		return nil, err
	}

	diskMeasurer := measurer.Disk{}

	diskMeasurer.Metadata = measurementsMetadata

	basicCollector, err := newBasicCollector(logger, client, &diskMeasurer, disksPrefix)
	if err != nil {
		return nil, err
	}

	return &Disks{basicCollector}, nil
}

// Collect implements prometheus.Collector.
func (c *Disks) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
		ch <- c.scrapeFailures
	}()

	disksMeasurements, failedScrapes, err := c.client.GetDiskMeasurements()
	if err != nil {
		c.up.Set(0)
	}
	c.up.Set(1)
	c.scrapeFailures.Add(float64(failedScrapes))

	for _, diskMeasurements := range disksMeasurements {
		for _, metric := range c.metrics {
			err = c.report(diskMeasurements, metric, ch)
			if err != nil {
				level.Debug(c.logger).Log("msg", `skipping metric`,
					"metric", metric.Desc, "err", err)
				continue
			}
		}
	}
}
