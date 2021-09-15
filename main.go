package main

import (
	"fmt"
	"mongodbatlas_exporter/collector"
	"mongodbatlas_exporter/mongodbatlas"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	name       = "mongodbatlas_exporter"
	appVersion = "0.0.1"
)

var (
	listenAddress           = kingpin.Flag("listen-address", "The address to listen on for HTTP requests.").Default(":9905").Envar("LISTEN_ADDRESS").String()
	atlasPublicKey          = kingpin.Flag("atlas.public-key", "Atlas API public key").Envar("ATLAS_PUBLIC_KEY").String()
	atlasPrivateKey         = kingpin.Flag("atlas.private-key", "Atlas API private key").Envar("ATLAS_PRIVATE_KEY").String()
	atlasProjectID          = kingpin.Flag("atlas.project-id", "Atlas project id (group id) to scrape metrics from").Envar("ATLAS_PROJECT_ID").String()
	atlasClusters           = kingpin.Flag("atlas.cluster", "Atlas cluster name to scrape metrics from. Can be defined multiple times. If not defined all clusters in the project will be scraped").Strings()
	logLevel                = kingpin.Flag("log-level", "Printed logs level.").Default("info").Enum("error", "warn", "info", "debug")
	collectorsIntervalRetry = kingpin.Flag("collectors-retry", "Prometheus collector initialization retry interval, in Minutes").Default("5").Int()

	up = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mongodbatlas_up",
		Help: "Was the last communication with MongoDB Atlas API successful and Project is not empty.",
	})
)

func main() {
	kingpin.Version(appVersion)
	kingpin.Parse()

	logger, err := createLogger(*logLevel)
	if err != nil {
		fmt.Printf("failed to create logger with error: %v", err)
		os.Exit(1)
	}

	versionMetric := version.NewCollector(name)
	prometheus.MustRegister(versionMetric)

	client, err := mongodbatlas.NewClient(logger, *atlasPublicKey, *atlasPrivateKey, *atlasProjectID, *atlasClusters)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create MongoDB Atlas client", "err", err)
		os.Exit(1)
	}

	go collectorInitRetry(logger, func() error {
		disksCollector, err := collector.NewDisks(logger, client)
		if err != nil {
			up.Set(0)
			return fmt.Errorf("failed to create Disks collector, err: %w", err)
		} else {
			prometheus.MustRegister(disksCollector)
		}
		return nil
	})

	go collectorInitRetry(logger, func() error {
		processesCollector, err := collector.NewProcesses(logger, client)
		if err != nil {
			up.Set(0)
			return fmt.Errorf("failed to create Processes collector, err: %w", err)
		} else {
			prometheus.MustRegister(processesCollector)
		}
		return nil
	})

	up.Set(1)

	http.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "failed to start the http server", "err", err)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "successfully started http server", "address", listenAddress)
}

func collectorInitRetry(logger log.Logger, f func() error) {
	for {
		err := f()
		if err == nil {
			return
		}
		level.Warn(logger).Log("msg", "retrying initialize collector after error", "err", err)
		time.Sleep(time.Duration(*collectorsIntervalRetry) * time.Minute)
	}
}
