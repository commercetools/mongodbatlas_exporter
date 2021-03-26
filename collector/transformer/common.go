package transformer

import (
	"math"
	m "mongodbatlas_exporter/model"
)

type unitTransformationRules struct {
	valueMultiplier float64
	nameSuffix      string
}

var unitsTransformationRules = map[m.UnitEnum]unitTransformationRules{
	m.PERCENT:              {valueMultiplier: 1, nameSuffix: "_percent"},
	m.MILLISECONDS:         {valueMultiplier: 0.001, nameSuffix: "_seconds"},
	m.SECONDS:              {valueMultiplier: 1, nameSuffix: "_seconds"},
	m.BYTES:                {valueMultiplier: 1, nameSuffix: "_bytes"},
	m.KILOBYTES:            {valueMultiplier: 1024, nameSuffix: "_bytes"},
	m.MEGABYTES:            {valueMultiplier: math.Pow(1024, 2), nameSuffix: "_bytes"},
	m.GIGABYTES:            {valueMultiplier: math.Pow(1024, 3), nameSuffix: "_bytes"},
	m.BYTES_PER_SECOND:     {valueMultiplier: 1, nameSuffix: "_bytes_ratio"},
	m.MEGABYTES_PER_SECOND: {valueMultiplier: math.Pow(1024, 2), nameSuffix: "_bytes_ratio"},
	m.GIGABYTES_PER_HOUR:   {valueMultiplier: math.Pow(1024, 3), nameSuffix: "_bytes_ratio_rate1h"},
	m.SCALAR_PER_SECOND:    {valueMultiplier: 1, nameSuffix: "_ratio"},
	m.SCALAR:               {valueMultiplier: 1, nameSuffix: ""},
}
