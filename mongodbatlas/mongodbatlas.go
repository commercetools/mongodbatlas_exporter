package mongodbatlas

import (
	"context"
	"errors"
	m "mongodbatlas_exporter/model"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mongodb-forks/digest"
	"go.mongodb.org/atlas/mongodbatlas"
)

var opts = &mongodbatlas.ProcessMeasurementListOptions{
	Granularity: "PT1M",
	Period:      "PT2M",
}

// AtlasClient implements mongodbatlas.Client
type AtlasClient struct {
	mongodbatlasClient *mongodbatlas.Client
	projectID          string
	atlasClusters      []string
	logger             log.Logger
}

// NewClient returns wrapper around mongodbatlas.Client, which implements necessary functionality
func NewClient(logger log.Logger, publicKey string, privateKey string, projectID string, atlasClusters []string) (*AtlasClient, error) {
	t := digest.NewTransport(publicKey, privateKey)
	tc, err := t.Client()
	if err != nil {
		level.Error(logger).Log("msg", "failed to auth", "err", err)
		return nil, errors.New("can't create mongodbatlas client, failed to auth, please check credentials")
	}
	mongodbatlasClient := mongodbatlas.NewClient(tc)
	level.Debug(logger).Log("msg", "mongodbatlas client was successfully created")

	return &AtlasClient{
		mongodbatlasClient: mongodbatlasClient,
		projectID:          projectID,
		atlasClusters:      atlasClusters,
		logger:             logger,
	}, nil
}

func (c *AtlasClient) listProcesses() ([]*mongodbatlas.Process, error) {
	processes, _, err := c.mongodbatlasClient.Processes.List(context.Background(), c.projectID, nil)
	if err != nil {
		msg := "failed to list processes of the project"
		level.Error(c.logger).Log("msg", msg, "project", c.projectID, "err", err)
		return nil, errors.New(msg)
	}
	if len(c.atlasClusters) == 0 {
		return processes, nil
	}
	filteredProceses := make([]*mongodbatlas.Process, 0, len(processes))
	for _, clusterName := range c.atlasClusters {
		for _, process := range processes {
			if strings.HasPrefix(process.UserAlias, clusterName) {
				filteredProceses = append(filteredProceses, process)
			}
		}
	}
	return filteredProceses, nil
}

func (c *AtlasClient) listDisks(host string, port int) ([]*mongodbatlas.ProcessDisk, error) {
	disks, _, err := c.mongodbatlasClient.ProcessDisks.List(context.Background(), c.projectID, host, port, nil)
	if err != nil {
		return nil, err
	}
	return disks.Results, nil
}

func (c *AtlasClient) listProcessDiskMeasurements(host string, port int, diskName string) (*mongodbatlas.ProcessDiskMeasurements, error) {
	measurements, _, err := c.mongodbatlasClient.ProcessDiskMeasurements.List(context.Background(), c.projectID, host, port, diskName, opts)
	if err != nil {
		return nil, err
	}
	return measurements, nil
}

func (c *AtlasClient) listProcessMeasurements(host string, port int) (*mongodbatlas.ProcessMeasurements, error) {
	measurements, _, err := c.mongodbatlasClient.ProcessMeasurements.List(context.Background(), c.projectID, host, port, opts)
	if err != nil {
		return nil, err
	}
	return measurements, nil
}

// GetDiskMeasurements returns measurements for all disks of a Project
func (c *AtlasClient) GetDiskMeasurements() ([]*m.DiskMeasurements, m.ScrapeFailures, error) {
	scrapeFailures := m.ScrapeFailures(0)

	processes, err := c.listProcesses()
	if err != nil {
		return nil, 0, err
	}

	result := make([]*m.DiskMeasurements, 0, len(processes))
	for _, process := range processes {
		disks, err := c.listDisks(process.Hostname, process.Port)
		if err != nil {
			level.Error(c.logger).Log("msg", "failed to list disks of the process", "process", process.Hostname, "port", process.Port, "err", err)
			return nil, 0, err
		}
		for _, disk := range disks {
			measurements, err := c.listProcessDiskMeasurements(process.Hostname, process.Port, disk.PartitionName)
			if err != nil {
				scrapeFailures++
				level.Warn(c.logger).Log("msg", "failed to scrape measurements for the disk, skipping", "disk", disk.PartitionName, "err", err)
				continue
			}
			diskMeasurements := make(m.MeasurementMap, len(measurements.Measurements))
			for _, measurement := range measurements.Measurements {
				diskMeasurements.RegisterAtlasMeasurement(measurement)
			}
			result = append(result, &m.DiskMeasurements{
				Measurements:  diskMeasurements,
				ProjectID:     process.GroupID,
				RsName:        process.ReplicaSetName,
				UserAlias:     process.UserAlias + ":" + strconv.Itoa(process.Port),
				PartitionName: disk.PartitionName,
			})
		}
	}

	return result, scrapeFailures, err
}

// GetProcessMeasurements returns measurements for all processes of a Project
func (c *AtlasClient) GetProcessMeasurements() ([]*m.ProcessMeasurements, m.ScrapeFailures, error) {
	scrapeFailures := m.ScrapeFailures(0)

	processes, err := c.listProcesses()
	if err != nil {
		return nil, 0, err
	}

	result := make([]*m.ProcessMeasurements, 0, len(processes))
	for _, process := range processes {
		measurements, err := c.listProcessMeasurements(process.Hostname, process.Port)
		if err != nil {
			scrapeFailures++
			level.Warn(c.logger).Log("msg", "failed to scrape measurements for the host, skipping", "host", process.Hostname, "port", process.Port, "err", err)
			continue
		}
		processMeasurements := make(m.MeasurementMap, len(measurements.Measurements))
		for _, measurement := range measurements.Measurements {
			processMeasurements.RegisterAtlasMeasurement(measurement)
		}
		result = append(result, &m.ProcessMeasurements{
			Measurements: processMeasurements,
			ProjectID:    process.GroupID,
			RsName:       process.ReplicaSetName,
			UserAlias:    process.UserAlias + ":" + strconv.Itoa(process.Port),
			Version:      process.Version,
			TypeName:     process.TypeName,
		})
	}

	return result, scrapeFailures, err
}

// GetDiskMeasurementsMetadata returns name and unit of all available Disk measurements
func (c *AtlasClient) GetDiskMeasurementMap() (m.MeasurementMap, error) {
	processes, err := c.listProcesses()
	if err != nil {
		return nil, err
	}
	for _, process := range processes {
		if process.TypeName != "SHARD_MONGOS" {
			result, err := c.getDiskMeasurementsForMetadata(process.Hostname, process.Port)
			if err != nil || len(result) < 1 {
				continue
			}
			return result, nil
		}
	}
	return nil, errors.New("can't find any resource with disk measurements, please create Atlas resources first and restart the exporter")
}

func (c *AtlasClient) getDiskMeasurementsForMetadata(host string, port int) (m.MeasurementMap, error) {
	disks, err := c.listDisks(host, port)
	if err != nil {
		return nil, err
	}
	// At the moment of writing: 1 mongod disk expose 10 measurements
	result := make(m.MeasurementMap, 10)
	for _, disk := range disks {
		diskName := disk.PartitionName
		diskMeasurements, err := c.listProcessDiskMeasurements(host, port, diskName)
		if err != nil {
			return nil, err
		}
		for _, measurement := range diskMeasurements.Measurements {
			result.RegisterAtlasMeasurement(measurement)
		}
	}
	return result, nil
}

// GetProcessMeasurementsMetadata returns name and unit of all available Process measurements
func (c *AtlasClient) GetProcessMeasurementMap() (m.MeasurementMap, error) {
	processes, err := c.listProcesses()
	if err != nil {
		return nil, err
	}
	// At the moment of writing: there are 5 mongod process types
	// SHARD_MONGOS, REPLICA_SECONDARY, REPLICA_PRIMARY, SHARD_CONFIG_PRIMARY, SHARD_CONFIG_SECONDARY
	processTypes := make(map[string]bool, 5)
	// At the moment of writing: mongod process expose 96 measurements
	// measurements for mognod process and mongos process measurements contain different amount of measurements,
	// mongod contains all measurements that mongos provides and some more specific to mongod measurements
	// (like `CACHE_*`,  `DB_*`, `DOCUMENT_*`, `GLOBAL_LOCK_CURRENT_QUEUE_*`, etc)
	result := make(m.MeasurementMap, 96)

	for _, process := range processes {
		if _, ok := processTypes[process.TypeName]; ok {
			continue
		}
		processMeasurements, err := c.listProcessMeasurements(process.Hostname, process.Port)
		if err != nil {
			continue
		}
		processTypes[process.TypeName] = true
		if len(processMeasurements.Measurements) > 0 {
			for _, measurement := range processMeasurements.Measurements {
				result.RegisterAtlasMeasurement(measurement)
			}
		}
	}

	if len(result) < 1 {
		return nil, errors.New("can't find any resource with process measurements, please create Atlas resources first and restart the exporter")
	}

	return result, nil
}
