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
	listenAddress   = kingpin.Flag("listen-address", "Address to listen on for web interface").Default(":9905").Envar("LISTEN_ADDRESS").String()
	atlasPublicKey  = kingpin.Flag("atlas.public-key", "Address to listen on for web interface").Envar("ATLAS_PUBLIC_KEY").String()
	atlasPrivateKey = kingpin.Flag("atlas.private-key", "Address to listen on for web interface").Envar("ATLAS_PRIVATE_KEY").String()
	atlasGroupID    = kingpin.Flag("atlas.group-id", "Project ID").Envar("GROUP_ID").String()
	atlasClusters   = kingpin.Flag("atlas.cluster", "Cluster to scrape metrics from, for multiple clusters define this flag multiple times").Strings()
	logLevel        = kingpin.Flag("log-level", "Printed logs level.").Default("debug").Enum("error", "warn", "info", "debug")
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

	client, err := mongodbatlas.NewClient(logger, *atlasPublicKey, *atlasPrivateKey, *atlasGroupID, *atlasClusters)
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