package main

import (
	"fmt"
	"mongodbatlas_exporter/mongodbatlas"
	"mongodbatlas_exporter/registerer"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	name = "mongodbatlas_exporter"
)

var (
	listenAddress   = kingpin.Flag("listen-address", "The address to listen on for HTTP requests.").Default(":9905").Envar("LISTEN_ADDRESS").String()
	atlasPublicKey  = kingpin.Flag("atlas.public-key", "Atlas API public key").Envar("ATLAS_PUBLIC_KEY").String()
	atlasPrivateKey = kingpin.Flag("atlas.private-key", "Atlas API private key").Envar("ATLAS_PRIVATE_KEY").String()
	atlasProjectID  = kingpin.Flag("atlas.project-id", "Atlas project id (group id) to scrape metrics from").Envar("ATLAS_PROJECT_ID").String()
	atlasClusters   = kingpin.Flag("atlas.cluster", "Atlas cluster name to scrape metrics from. Can be defined multiple times. If not defined all clusters in the project will be scraped").Strings()
	logLevel        = kingpin.Flag("log-level", "Printed logs level.").Default("info").Enum("error", "warn", "info", "debug")
	up              = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mongodbatlas_up",
		Help: "Was the last communication with MongoDB Atlas API successful and Project is not empty.",
	})
)

func main() {
	kingpin.Version(version.Print(name))
	kingpin.Parse()

	logger, err := createLogger(*logLevel)
	if err != nil {
		fmt.Printf("failed to create logger with error: %v", err)
		os.Exit(1)
	}

	prometheus.MustRegister(version.NewCollector(name))

	client, err := mongodbatlas.NewClient(logger, *atlasPublicKey, *atlasPrivateKey, *atlasProjectID, *atlasClusters)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create MongoDB Atlas client", "err", err)
		os.Exit(1)
	}

	processRegister := registerer.NewProcessRegisterer(logger, client, time.Minute)

	go processRegister.Observe()

	up.Set(1)

	http.Handle("/metrics", promhttp.Handler())
	//health
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "health is ok"); err != nil {
			level.Error(logger).Log("msg", "failed to start the http server", "err", err)

		}
	})
	//ready
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "readyness is good"); err != nil {
			level.Error(logger).Log("msg", "failed to start the http server", "err", err)
		}
	})
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "failed to start the http server", "err", err)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "successfully started http server", "address", listenAddress)
}
