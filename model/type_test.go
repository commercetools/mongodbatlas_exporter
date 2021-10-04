package model

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestTypeTransformation(t *testing.T) {
	assert := assert.New(t)

	promType := exampleMeasurement.PromType()

	assert.Equal(prometheus.GaugeValue, promType)
}
