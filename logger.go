package main

import (
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func createLogger(l string) (logger log.Logger, err error) {
	var lvl level.Option
	switch l {
	case "error":
		lvl = level.AllowError()
	case "warn":
		lvl = level.AllowWarn()
	case "info":
		lvl = level.AllowInfo()
	case "debug":
		lvl = level.AllowDebug()
	default:
		return nil, fmt.Errorf("unrecognized log level: %v", l)
	}

	logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, lvl)
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)

	return logger, nil
}
