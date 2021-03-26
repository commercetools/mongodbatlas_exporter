package transformer

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestTypeTransformation(t *testing.T) {
	assert := assert.New(t)

	promType, err := TransformType(exampleMeasurementMetadata)

	assert.NoError(err)
	assert.Equal(prometheus.GaugeValue, promType)
}
