package model

import (
	"go.mongodb.org/atlas/mongodbatlas"
)

/*
* MeasurementMaps provide quick access to identify which Atlas Measurements we currently know about
* and have registered with the prometheus collector.
* When a "measurement" is registered in the map it should also be registered with the collector.
 */

type MeasurementMap map[MeasurementID]*Measurement

func (measurementMap *MeasurementMap) RegisterAtlasMeasurement(measurement *mongodbatlas.Measurements) {
	measurementID := NewMeasurementID(measurement.Name, measurement.Units)
	(*measurementMap)[measurementID] = &Measurement{
		Name:       measurement.Name,
		DataPoints: measurement.DataPoints,
		Units:      UnitEnum(measurement.Units),
	}
}

func (measurementMap *MeasurementMap) RegisterMeasurement(measurement Measurement) {
	(*measurementMap)[measurement.ID()] = &measurement
}
