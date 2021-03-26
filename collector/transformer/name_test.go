package transformer

import (
	m "mongodbatlas_exporter/model"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameTransformer(t *testing.T) {
	assert := assert.New(t)
	exampleName := "EXAMPLE_MeasuRemenT"

	for unitName, unitTransformation := range unitsTransformationRules {
		measurement := *exampleMeasurementMetadata
		measurement.Name = exampleName
		measurement.Units = unitName
		expectedName := strings.Join([]string{strings.ToLower(exampleName), unitTransformation.nameSuffix}, "")

		promName, err := TransformName(&measurement)

		assert.NoError(err)
		assert.Equal(expectedName, promName)
	}
}

func TestNameTransformer_noName(t *testing.T) {
	assert := assert.New(t)
	exampleName := ""

	for unitName := range unitsTransformationRules {
		measurement := *exampleMeasurementMetadata
		measurement.Name = exampleName
		measurement.Units = unitName

		promName, err := TransformName(&measurement)

		assert.Error(err)
		assert.Empty(promName)
	}
}

func TestNameTransformer_noUnit(t *testing.T) {
	assert := assert.New(t)
	exampleName := "EXAMPLE_MeasuRemenT"
	var unit m.UnitEnum
	measurement := *exampleMeasurementMetadata
	measurement.Name = exampleName
	measurement.Units = unit

	promName, err := TransformName(&measurement)

	assert.Error(err)
	assert.Empty(promName)
}

func TestNameTransformer_scalar(t *testing.T) {
	assert := assert.New(t)
	exampleName := "EXAMPLE_MeasuRemenT"
	var unit m.UnitEnum = "SCALAR"
	measurement := *exampleMeasurementMetadata
	measurement.Name = exampleName
	measurement.Units = unit

	promName, err := TransformName(&measurement)

	assert.NoError(err)
	assert.Equal("example_measurement", promName)
}
