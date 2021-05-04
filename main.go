package main

import (
	"fmt"
	"mongodbatlas_exporter/collector"
	"mongodbatlas_exporter/mongodbatlas"
	"net/http"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	name       = "mongodbatlas_exporter"
	appVersion = "0.0.1"
)

var (
	listenAddress   = kingpin.Flag("listen-address", "The address to listen on for HTTP requests.").Default(":9905").Envar("LISTEN_ADDRESS").String()
	atlasPublicKey  = kingpin.Flag("atlas.public-key", "Atlas API public key").Envar("ATLAS_PUBLIC_KEY").String()
	atlasPrivateKey = kingpin.Flag("atlas.private-key", "Atlas API private key").Envar("ATLAS_PRIVATE_KEY").String()
	atlasProjectID  = kingpin.Flag("atlas.project-id", "Atlas project id (group id) to scrape metrics from").Envar("ATLAS_PROJECT_ID").String()
	atlasClusters   = kingpin.Flag("atlas.cluster", "Atlas cluster name to scrape metrics from. Can be defined multiple times. If not defined all clusters in the project will be scraped").Strings()
	logLevel        = kingpin.Flag("log-level", "Printed logs level.").Default("warn").Enum("error", "warn", "info", "debug")
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

	disksCollector, err := collector.NewDisks(logger, client)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create Disks collector", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(disksCollector)

	processesCollector, err := collector.NewProcesses(logger, client)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create Processes collector", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(processesCollector)

	http.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "failed to start the server", "err", err)
		os.Exit(1)
	}
}
