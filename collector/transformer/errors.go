package transformer

import "errors"

var (
	//ErrNoData represents the very common case where Atlas returns
	//a metric with no datapoints. For example any FTS* metric will
	//return no datapoints every time if a cluster does not use "full text search".
	ErrNoData = errors.New("no datapoints are available")
)
