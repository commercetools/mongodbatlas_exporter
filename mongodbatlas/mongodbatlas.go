package mongodbatlas

import (
	"context"
	"errors"
	m "mongodbatlas_exporter/model"
	"strconv"

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
	logger             log.Logger
}

// NewClient returns wrapper around mongodbatlas.Client, which implements necessary functionality
func NewClient(logger log.Logger, publicKey string, privateKey string, projectID string) (*AtlasClient, error) {
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
	return processes, nil
}

func (c *AtlasClient) listDisks(host string, port int) ([]*mongodbatlas.ProcessDisk, error) {
	disks, _, err := c.mongodbatlasClient.ProcessDisks.List(context.Background(), c.projectID, host, port, nil)
	if err != nil {
		msg := "failed to list disks of the process"
		level.Error(c.logger).Log("msg", msg, "process", host, "port", port, "err", err)
		return nil, errors.New(msg)
	}
	return disks.Results, nil
}

func (c *AtlasClient) listProcessDiskMeasurements(host string, port int, diskName string) (*mongodbatlas.ProcessDiskMeasurements, error) {
	measurements, _, err := c.mongodbatlasClient.ProcessDiskMeasurements.List(context.Background(), c.projectID, host, port, diskName, opts)
	if err != nil {
		msg := "failed to list measurements of the disk"
		level.Error(c.logger).Log("msg", msg, "disk", diskName, "err", err)
		return nil, errors.New(msg)
	}
	return measurements, nil
}

func (c *AtlasClient) listProcessMeasurements(host string, port int) (*mongodbatlas.ProcessMeasurements, error) {
	measurements, _, err := c.mongodbatlasClient.ProcessMeasurements.List(context.Background(), c.projectID, host, port, opts)
	if err != nil {
		msg := "failed to list measurements of the process"
		level.Error(c.logger).Log("msg", msg, "host", host, "err", err)
		return nil, errors.New(msg)
	}
	return measurements, nil
}

// GetDiskMeasurements returns measurements for all disks of a Project
func (c *AtlasClient) GetDiskMeasurements() ([]*m.DiskMeasurements, m.ScrapeFailures, error) {
	var result []*m.DiskMeasurements
	scrapeFailures := m.ScrapeFailures(0)

	processes, err := c.listProcesses()
	if err != nil {
		return nil, 0, err
	}
	for _, process := range processes {
		disks, err := c.listDisks(process.Hostname, process.Port)
		if err != nil {
			return nil, 0, err
		}
		for _, disk := range disks {
			measurements, err := c.listProcessDiskMeasurements(process.Hostname, process.Port, disk.PartitionName)
			if err != nil {
				scrapeFailures++
				level.Error(c.logger).Log("msg", "failed to scrape measurements for the disk, skipping", "disk", disk.PartitionName, "err", err)
				continue
			}
			diskMeasurements := make(map[m.MeasurementID]*m.Measurement, len(measurements.Measurements))
			for _, measurement := range measurements.Measurements {
				measurementID := m.NewMeasurementID(measurement.Name, measurement.Units)
				diskMeasurements[measurementID] = &m.Measurement{
					DataPoints: measurement.DataPoints,
					Units:      m.UnitEnum(measurement.Units),
				}
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
	var result []*m.ProcessMeasurements
	scrapeFailures := m.ScrapeFailures(0)

	processes, err := c.listProcesses()
	if err != nil {
		return nil, 0, err
	}
	for _, process := range processes {
		measurements, err := c.listProcessMeasurements(process.Hostname, process.Port)
		if err != nil {
			scrapeFailures++
			level.Error(c.logger).Log("msg", "failed to scrape measurements for the host, skipping", "host", process.Hostname, "port", process.Port, "err", err)
			continue
		}
		processMeasurements := make(map[m.MeasurementID]*m.Measurement, len(measurements.Measurements))
		for _, measurement := range measurements.Measurements {
			measurementID := m.NewMeasurementID(measurement.Name, measurement.Units)
			processMeasurements[measurementID] = &m.Measurement{
				DataPoints: measurement.DataPoints,
				Units:      m.UnitEnum(measurement.Units),
			}
		}
		result = append(result, &m.ProcessMeasurements{
			Measurements: processMeasurements,
			ProjectID:    process.GroupID,
			RsName:       process.ReplicaSetName,
			UserAlias:    process.UserAlias + ":" + strconv.Itoa(process.Port),
		})
	}

	return result, scrapeFailures, err
}

// GetDiskMeasurementsMetadata returns name and unit of all available Disk measurements
func (c *AtlasClient) GetDiskMeasurementsMetadata() ([]*m.MeasurementMetadata, error) {
	diskMeasurements, err := c.getDiskMeasurementsForMetadata()
	if err != nil {
		msg := "can't get any disk measurements for metadata"
		level.Error(c.logger).Log("msg", msg, "err", err)
		return nil, errors.New(msg)
	}

	var result []*m.MeasurementMetadata
	for _, measurement := range diskMeasurements.Measurements {
		result = append(result, &m.MeasurementMetadata{
			Name:  measurement.Name,
			Units: m.UnitEnum(measurement.Units),
		})
	}

	return result, nil
}

func (c *AtlasClient) getDiskMeasurementsForMetadata() (*mongodbatlas.ProcessDiskMeasurements, error) {
	processes, err := c.listProcesses()
	if err != nil {
		return nil, err
	}

	for _, process := range processes {
		disks, err := c.listDisks(process.Hostname, process.Port)
		if err != nil {
			continue
		}
		if len(disks) < 1 {
			continue
		}

		for _, disk := range disks {
			diskName := disk.PartitionName
			diskMeasurements, err := c.listProcessDiskMeasurements(process.Hostname, process.Port, diskName)
			if err != nil {
				return nil, err
			}
			if len(diskMeasurements.Measurements) > 0 {
				return diskMeasurements, nil
			}
		}
	}

	return nil, errors.New("can't find any resource with disk measurements, please create Atlas resources first and restart the exporter")
}

// GetProcessMeasurementsMetadata returns name and unit of all available Process measurements
func (c *AtlasClient) GetProcessMeasurementsMetadata() ([]*m.MeasurementMetadata, error) {
	processMeasurements, err := c.getProcessMeasurementsForMetadata()
	if err != nil {
		msg := "can't get any process measurements for metadata"
		level.Error(c.logger).Log("msg", msg, "err", err)
		return nil, errors.New(msg)
	}

	var result []*m.MeasurementMetadata
	for _, measurement := range processMeasurements.Measurements {
		result = append(result, &m.MeasurementMetadata{
			Name:  measurement.Name,
			Units: m.UnitEnum(measurement.Units),
		})
	}

	return result, nil
}

func (c *AtlasClient) getProcessMeasurementsForMetadata() (*mongodbatlas.ProcessMeasurements, error) {
	processes, err := c.listProcesses()
	if err != nil {
		return nil, err
	}

	for _, process := range processes {
		processMeasurements, err := c.listProcessMeasurements(process.Hostname, process.Port)
		if err != nil {
			continue
		}
		if len(processMeasurements.Measurements) > 0 {
			return processMeasurements, nil
		}
	}

	return nil, errors.New("can't find any resource with process measurements, please create Atlas resources first and restart the exporter")
}
