package collector

import (
	transformer "mongodbatlas_exporter/collector/transformer"
	m "mongodbatlas_exporter/model"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var defaultDiskLabels = []string{"project_id", "rs_name", "user_alias", "partition_name"}

const disksPrefix = "disks_stats"

// Disks information struct
type Disks struct {
	*basicCollector
}

// NewDisks creates Disk Prometheus metrics
func NewDisks(logger log.Logger, client m.Client) (*Disks, error) {
	measurements, err := client.GetDiskMeasurementMap()
	if err != nil {
		return nil, err
	}

	basicCollector, err := newBasicCollector(logger, client, measurements, defaultDiskLabels, disksPrefix)
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
		ch <- c.measurementTransformationFailures
	}()

	disksMeasurements, failedScrapes, err := c.client.GetDiskMeasurements()
	if err != nil {
		c.up.Set(0)
	}
	c.up.Set(1)
	c.scrapeFailures.Add(float64(failedScrapes))

	for _, diskMeasurements := range disksMeasurements {
		for _, metric := range c.metrics {
			measurement, ok := diskMeasurements.Measurements[metric.Metadata.ID()]
			if !ok {
				c.measurementTransformationFailures.Inc()
				level.Warn(c.logger).Log("msg", `skipping metric because can't find matching measurement.
					It seems to be not initialized during exporter start, you should restart the exporter`,
					"metric", metric.Desc, "err", err)
				continue
			}
			value, err := transformer.TransformValue(measurement)
			if err != nil {
				c.measurementTransformationFailures.Inc()
				level.Warn(c.logger).Log("msg", "skipping metric because of value transformation failure", "metric", metric.Desc, "measurement", measurement, "err", err)
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				value,
				diskMeasurements.ProjectID, diskMeasurements.RsName, diskMeasurements.UserAlias, diskMeasurements.PartitionName,
			)
		}
	}
}
