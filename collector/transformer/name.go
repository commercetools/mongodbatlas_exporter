package transformer

import (
	"errors"
	"strings"

	m "mongodbatlas_exporter/model"
)

const nameDelimiter = ""

// TransformName transforms MeasurementMetadata into string for Prometheus metric name
func TransformName(measurement *m.MeasurementMetadata) (string, error) {
	emptyName := len(measurement.Name) < 1
	unit, knownUnit := unitsTransformationRules[measurement.Units]

	if !emptyName && knownUnit {
		lowercaseName := strings.ToLower(measurement.Name)
		return strings.Join([]string{lowercaseName, unit.nameSuffix}, nameDelimiter), nil
	}

	var msg string
	if emptyName {
		msg += "Can't transform name '" + measurement.Name + "', it seems to be invalid. "
	}

	if !knownUnit {
		msg += "Can't find suffix for unit '" + string(measurement.Units) + "', the unit type seems to be unknown."
	}

	return "", errors.New(msg)
}
