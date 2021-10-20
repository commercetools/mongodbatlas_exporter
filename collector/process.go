package collector

import (
	"mongodbatlas_exporter/measurer"
	a "mongodbatlas_exporter/mongodbatlas"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

const (
	processesPrefix = "processes_stats"
	disksPrefix     = "disks_stats"
	infoHelp        = "Process info metric"
)

// Process information struct
type Process struct {
	*basicCollector
	info     prometheus.Gauge
	measurer measurer.Process
}

func NewProcessCollector(logger log.Logger, client a.Client, p *mongodbatlas.Process) (*Process, error) {

	processMeasurer := measurer.ProcessFromMongodbAtlasProcess(p)

	//Spice the Measurer with the list of disks.
	disks, httpErr := client.ListDisks(p)

	if httpErr != nil {
		return nil, httpErr
	}

	//MONGOS nodes have disks that do not report data so
	//we skip registering them.
	//https://github.com/commercetools/mongodbatlas_exporter/issues/15
	if p.TypeName != a.TYPE_MONGOS {
		processMeasurer.Disks = make([]*measurer.Disk, len(disks))
		for i := range disks {
			disk := measurer.DiskFromMongodbAtlasProcessDisk(p, disks[i])
			diskMetadata, err := client.GetDiskMeasurementsMetadata(processMeasurer, disk)
			if err != nil {
				//TODO: Definitely needs to be a metric
				level.Warn(logger).Log("msg", "could not get disk metadata", "disk", disk.PartitionName, "process", p.ID, "group", p.GroupID)
				continue
			}

			//assign metadata
			disk.Metadata = diskMetadata

			//build list of prometheus metrics
			err = measurer.BuildPromMetrics(disk, namespace, disksPrefix)

			if err != nil {
				level.Warn(logger).Log("msg", "could not build disk prom metrics", "disk", disk.PartitionName, "process", p.ID, "group", p.GroupID)
				continue
			}

			//if everything succeeds add the disk to the list.
			processMeasurer.Disks[i] = disk
		}
	}

	//get the metadata for the measurer.
	//this should be part of the measurer.
	httpErr = client.GetProcessMeasurementsMetadata(processMeasurer)
	if httpErr != nil {
		return nil, httpErr
	}

	err := measurer.BuildPromMetrics(processMeasurer, namespace, processesPrefix)

	if err != nil {
		return nil, err
	}

	for i := range processMeasurer.Disks {
		err = measurer.BuildPromMetrics(processMeasurer.Disks[i], namespace, disksPrefix)

		if err != nil {
			return nil, err
		}
	}

	basicCollector, err := newBasicCollector(logger, client, processMeasurer, processesPrefix)

	if err != nil {
		return nil, err
	}

	process := &Process{
		basicCollector: basicCollector,
		info: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        prometheus.BuildFQName(namespace, processesPrefix, "info"),
				Help:        infoHelp,
				ConstLabels: processMeasurer.PromInfoConstLabels(),
			}),
		measurer: *processMeasurer,
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

	for _, metric := range c.measurer.PromMetrics() {
		err = c.report(&c.measurer, metric, ch)
		if err != nil {
			level.Debug(c.logger).Log("msg", "skipping metric", "metric", metric.Desc,
				"err", err)
		}
	}

	c.info.Set(1)
	ch <- c.info

	for _, disk := range c.measurer.Disks {
		err := c.client.GetDiskMeasurements(&c.measurer, disk)

		if err != nil {
			level.Debug(c.logger).Log("msg", "skipping disk", "disk", disk.PartitionName, "host", disk.ID,
				"err", err)
		}
		for _, metric := range disk.PromMetrics() {
			err = c.report(disk, metric, ch)
			if err != nil {
				level.Debug(c.logger).Log("msg", "skipping metric", "metric", metric.Desc,
					"err", err)
			}
		}
	}
}

// Describe implements prometheus.Collector.
func (c *Process) Describe(ch chan<- *prometheus.Desc) {
	c.basicCollector.Describe(ch)

	//add the disk metrics
	for _, d := range c.measurer.Disks {
		for _, metric := range d.PromMetrics() {
			ch <- metric.Desc
		}
	}
	c.info.Describe(ch)
}
