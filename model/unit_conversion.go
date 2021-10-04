package model

import (
	"math"
)

type unitTransformationRules struct {
	valueMultiplier float64
	nameSuffix      string
}

var unitsTransformationRules = map[UnitEnum]unitTransformationRules{
	PERCENT:              {valueMultiplier: 1, nameSuffix: "_percent"},
	MILLISECONDS:         {valueMultiplier: 0.001, nameSuffix: "_seconds"},
	SECONDS:              {valueMultiplier: 1, nameSuffix: "_seconds"},
	BYTES:                {valueMultiplier: 1, nameSuffix: "_bytes"},
	KILOBYTES:            {valueMultiplier: 1024, nameSuffix: "_bytes"},
	MEGABYTES:            {valueMultiplier: math.Pow(1024, 2), nameSuffix: "_bytes"},
	GIGABYTES:            {valueMultiplier: math.Pow(1024, 3), nameSuffix: "_bytes"},
	BYTES_PER_SECOND:     {valueMultiplier: 1, nameSuffix: "_bytes_ratio"},
	MEGABYTES_PER_SECOND: {valueMultiplier: math.Pow(1024, 2), nameSuffix: "_bytes_ratio"},
	GIGABYTES_PER_HOUR:   {valueMultiplier: math.Pow(1024, 3), nameSuffix: "_bytes_ratio_rate1h"},
	SCALAR_PER_SECOND:    {valueMultiplier: 1, nameSuffix: "_ratio"},
	SCALAR:               {valueMultiplier: 1, nameSuffix: ""},
}
