package mongodbatlas

import (
	"context"
	"errors"
	"mongodbatlas_exporter/measurer"
	m "mongodbatlas_exporter/model"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mongodb-forks/digest"
	"go.mongodb.org/atlas/mongodbatlas"
)

const (
	TYPE_MONGOS = "SHARD_MONGOS"
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

// Client wraps mongodbatlas.Client
type Client interface {
	GetDiskMeasurements(*measurer.Process, *measurer.Disk) error
	GetProcessMeasurements(measurer.Process) (map[m.MeasurementID]*m.Measurement, error)
	GetDiskMeasurementsMetadata(*measurer.Process, *measurer.Disk) (map[m.MeasurementID]*m.MeasurementMetadata, error)
	GetProcessMeasurementsMetadata(*measurer.Process) *HTTPError
	ListProcesses() ([]*mongodbatlas.Process, *HTTPError)
	ListDisks(*mongodbatlas.Process) ([]*mongodbatlas.ProcessDisk, *HTTPError)
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

func (c *AtlasClient) ListProcesses() ([]*mongodbatlas.Process, *HTTPError) {
	processes, r, err := c.mongodbatlasClient.Processes.List(context.Background(), c.projectID, nil)
	if err != nil {
		msg := "failed to list processes of the project"
		level.Error(c.logger).Log("msg", msg, "project", c.projectID, "err", err)
		return nil, &HTTPError{
			Err:        err,
			StatusCode: r.StatusCode,
		}
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

func (c *AtlasClient) ListDisks(p *mongodbatlas.Process) ([]*mongodbatlas.ProcessDisk, *HTTPError) {
	disks, r, err := c.mongodbatlasClient.ProcessDisks.List(context.Background(), c.projectID, p.Hostname, p.Port, nil)

	if err != nil {
		return nil, &HTTPError{
			StatusCode: r.StatusCode,
			Err:        err,
		}
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

func (c *AtlasClient) listProcessMeasurements(host string, port int) (*mongodbatlas.ProcessMeasurements, *HTTPError) {
	measurements, r, err := c.mongodbatlasClient.ProcessMeasurements.List(context.Background(), c.projectID, host, port, opts)
	if err != nil {
		return nil, &HTTPError{
			Err:        err,
			StatusCode: r.StatusCode,
		}
	}
	return measurements, nil
}

// GetDiskMeasurements returns measurements for all disks of a Project
func (c *AtlasClient) GetDiskMeasurements(process *measurer.Process, disk *measurer.Disk) error {

	measurements, err := c.listProcessDiskMeasurements(process.Hostname, process.Port, disk.PartitionName)
	if err != nil {
		return err
	}

	disk.Measurements = make(map[m.MeasurementID]*m.Measurement, len(measurements.Measurements))
	for _, measurement := range measurements.Measurements {
		measurementID := m.NewMeasurementID(measurement.Name, measurement.Units)
		disk.Measurements[measurementID] = &m.Measurement{
			DataPoints: measurement.DataPoints,
			Units:      m.UnitEnum(measurement.Units),
		}
	}

	return err
}

// GetProcessMeasurements returns measurements for all processes of a Project
func (c *AtlasClient) GetProcessMeasurements(measurer measurer.Process) (map[m.MeasurementID]*m.Measurement, error) {
	measurements, err := c.listProcessMeasurements(measurer.Hostname, measurer.Port)
	if err != nil {
		return nil, err
	}
	processMeasurements := make(map[m.MeasurementID]*m.Measurement, len(measurements.Measurements))
	for _, measurement := range measurements.Measurements {
		measurementID := m.NewMeasurementID(measurement.Name, measurement.Units)
		processMeasurements[measurementID] = &m.Measurement{
			DataPoints: measurement.DataPoints,
			Units:      m.UnitEnum(measurement.Units),
		}
	}

	return processMeasurements, nil
}

// GetDiskMeasurementsMetadata returns name and unit of all available Disk measurements
func (c *AtlasClient) GetDiskMeasurementsMetadata(p *measurer.Process, d *measurer.Disk) (map[m.MeasurementID]*m.MeasurementMetadata, error) {
	result, err := c.getDiskMeasurementsForMetadata(p.Hostname, p.Port, d.PartitionName)

	if err != nil {
		return nil, err
	}

	if len(result) < 1 {
		return nil, errors.New("can't find any resource with disk measurements, please create Atlas resources first and restart the exporter")
	}
	return result, nil
}

func (c *AtlasClient) getDiskMeasurementsForMetadata(host string, port int, partitionName string) (map[m.MeasurementID]*m.MeasurementMetadata, error) {
	// At the moment of writing: 1 mongod disk expose 10 measurements
	result := make(map[m.MeasurementID]*m.MeasurementMetadata, 10)
	diskMeasurements, err := c.listProcessDiskMeasurements(host, port, partitionName)
	if err != nil {
		return nil, err
	}
	for _, measurement := range diskMeasurements.Measurements {
		metadata := &m.MeasurementMetadata{
			Name:  measurement.Name,
			Units: m.UnitEnum(measurement.Units),
		}
		result[metadata.ID()] = metadata
	}

	return result, nil
}

// GetProcessMeasurementsMetadata returns name and unit of all available Process measurements
func (c *AtlasClient) GetProcessMeasurementsMetadata(pMeasurer *measurer.Process) *HTTPError {
	// At the moment of writing: mongod process expose 96 measurements
	// measurements for mognod process and mongos process measurements contain different amount of measurements,
	// mongod contains all measurements that mongos provides and some more specific to mongod measurements
	// (like `CACHE_*`,  `DB_*`, `DOCUMENT_*`, `GLOBAL_LOCK_CURRENT_QUEUE_*`, etc)
	pMeasurer.Metadata = make(map[m.MeasurementID]*m.MeasurementMetadata, 96)

	processMeasurements, err := c.listProcessMeasurements(pMeasurer.Hostname, pMeasurer.Port)
	if err != nil {
		return err
	}

	if len(processMeasurements.Measurements) > 0 {
		for _, measurement := range processMeasurements.Measurements {
			metadata := &m.MeasurementMetadata{
				Name:  measurement.Name,
				Units: m.UnitEnum(measurement.Units),
			}
			pMeasurer.Metadata[metadata.ID()] = metadata
		}
	}

	if len(pMeasurer.Metadata) < 1 {
		return &HTTPError{
			Err: errors.New("can't find any resource with process measurements, please create Atlas resources first and restart the exporter"),
		}
	}

	return nil
}
