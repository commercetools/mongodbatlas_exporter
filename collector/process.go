package collector

import (
	"mongodbatlas_exporter/measurer"
	"mongodbatlas_exporter/model"
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

	processMetadata := measurer.ProcessFromMongodbAtlasProcess(p)

	//Spice the Measurer with the list of disks.
	disks, httpErr := client.ListDisks(p)

	if httpErr != nil {
		return nil, httpErr
	}

	//MONGOS nodes have disks that do not report data so
	//we skip registering them.
	//https://github.com/commercetools/mongodbatlas_exporter/issues/15
	if p.TypeName != a.TYPE_MONGOS {
		processMetadata.Disks = make([]*measurer.Disk, len(disks))
		for i := range disks {
			processMetadata.Disks[i] = measurer.DiskFromMongodbAtlasProcessDisk(disks[i])
			diskMetadata, err := client.GetDiskMeasurementsMetadata(processMetadata, processMetadata.Disks[i])
			if err != nil {
				//TODO: Definitely needs to be a metric
				level.Warn(logger).Log("msg", "could not get disk metadata", "disk", disks[i].PartitionName, "process", p.ID, "group", p.GroupID)
				continue
			}

			processMetadata.Disks[i].Metadata = diskMetadata
		}
	}

	//get the metadata for the measurer.
	//this should be part of the measurer.
	httpErr = client.GetProcessMeasurementsMetadata(processMetadata)
	if httpErr != nil {
		return nil, httpErr
	}

	basicCollector, err := newBasicCollector(logger, client, processMetadata, processesPrefix)

	if err != nil {
		return nil, err
	}

	//MONGOS nodes have disks that do not report data so
	//we skip registering them.
	//https://github.com/commercetools/mongodbatlas_exporter/issues/15
	if len(processMetadata.Disks) > 0 {
		//here a set is used to add each measurement to the collector just once.
		//each disk has the variable label: partition_name
		everyDiskMetadataKey := make(map[model.MeasurementID]bool, len(processMetadata.Disks[0].Metadata))
		for _, disk := range processMetadata.Disks {
			for key := range disk.Metadata {
				if _, ok := everyDiskMetadataKey[key]; !ok {
					everyDiskMetadataKey[key] = true
					metadata, ok := processMetadata.Metadata[key]

					if ok {
						metric, err := metadataToMetric(metadata, disksPrefix, disk.PromVariableLabelNames(), processMetadata.PromConstLabels())
						if err != nil {
							level.Warn(logger).Log("msg", "could not transform metadata to metric", "err", err)
							continue
						}
						basicCollector.metrics = append(basicCollector.metrics, metric)
					}
				}
			}
		}
	}

	process := &Process{
		basicCollector: basicCollector,
		info: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        prometheus.BuildFQName(namespace, processesPrefix, "info"),
				Help:        infoHelp,
				ConstLabels: processMetadata.PromConstLabels(),
			}),
		measurer: *processMetadata,
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

	//this entire block is because the collector has the process metrics registered to it
	//but not the disk metrics.
	//I think the metrics should be registered to the MEASURER instead.
	for _, disk := range c.measurer.Disks {
		err := c.client.GetDiskMeasurements(&c.measurer, disk)
		if err != nil {
			x := err.Error()
			level.Debug(c.logger).Log("msg", "scrape failure", "err", err, "x", x)
			c.scrapeFailures.Inc()
			c.up.Set(0)
		}
		c.up.Set(1)
		for _, metadata := range disk.Metadata {
			metric, err := metadataToMetric(metadata, disksPrefix, disk.PromVariableLabelNames(), c.measurer.PromConstLabels())
			if err != nil {
				level.Debug(c.logger).Log("msg", "could not convert disk metadata to metric", "err", err)
			}
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
	c.info.Describe(ch)
}
