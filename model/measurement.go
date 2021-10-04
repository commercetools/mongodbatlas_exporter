package model

import (
	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/atlas/mongodbatlas"
)

const (
	defaultHelp   = "Please see MongoDB Atlas documentation for details about the measurement"
	nameDelimiter = ""
)

// Measurement contains unit and mulpiple dataPoints of one measurement
type Measurement struct {
	Name       string
	DataPoints []*mongodbatlas.DataPoints
	Units      UnitEnum
}

// DiskMeasurements contains all measurements of one Disk
type DiskMeasurements struct {
	ProjectID, RsName, UserAlias, PartitionName string
	Measurements                                MeasurementMap
}

// ProcessMeasurements contains all measurements of one Process
type ProcessMeasurements struct {
	ProjectID, RsName, UserAlias, Version, TypeName string
	Measurements                                    MeasurementMap
}

// MeasurementID consists of Measurement.Name and Measurement.Units
type MeasurementID string

// ID returns identifier of the metric
func (c Measurement) ID() MeasurementID {
	return MeasurementID(c.Name + "_" + string(c.Units))
}

//PromType returns the prometheus type for the measurement.
//Currently all Atlas metrics are Gauge metrics from what we know.
func (m *Measurement) PromType() prometheus.ValueType {
	return prometheus.GaugeValue
}

//promname comes from transform, defaulthelp I think is just a constant.
func (m *Measurement) PromDesc(namespace, prefix string, variableLabels []string) (*prometheus.Desc, error) {
	promName, err := m.PromName()

	if err != nil {
		return nil, err
	}
	promFQDN := prometheus.BuildFQName(namespace, prefix, promName)
	return prometheus.NewDesc(
		promFQDN,
		"Original measurements.name: '"+m.Name+"'. "+defaultHelp,
		variableLabels, nil,
	), nil
}

// TransformName transforms Measurement into string for Prometheus metric name
func (m *Measurement) PromName() (string, error) {
	emptyName := len(m.Name) < 1
	unit, knownUnit := unitsTransformationRules[m.Units]

	if !emptyName && knownUnit {
		lowercaseName := strings.ToLower(m.Name)
		return strings.Join([]string{lowercaseName, unit.nameSuffix}, nameDelimiter), nil
	}

	var msg string
	if emptyName {
		msg += "Can't transform name '" + m.Name + "', it seems to be invalid. "
	}

	if !knownUnit {
		msg += "Can't find suffix for unit '" + string(m.Units) + "', the unit type seems to be unknown."
	}

	return "", errors.New(msg)
}
