package registerer

import (
	"mongodbatlas_exporter/collector"
	a "mongodbatlas_exporter/mongodbatlas"
	"strconv"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	namespace                                   = "mongodbatlas"
	subsystem                                   = "registerer"
	metadataScrapeErrors *prometheus.CounterVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "processes_metadatascrape",
	}, []string{"status"})
)

type ProcessRegisterer struct {
	collectors        map[string]prometheus.Collector
	reconcileInterval time.Duration
	client            a.Client
	logger            log.Logger
}

func NewProcessRegisterer(logger log.Logger, c a.Client, reconcileInterval time.Duration) *ProcessRegisterer {
	return &ProcessRegisterer{
		client:            c,
		logger:            logger,
		reconcileInterval: reconcileInterval,
		collectors:        make(map[string]prometheus.Collector),
	}
}

func (r *ProcessRegisterer) Observe() {
	//Keep the register up to date.
	for {
		r.registerAtlasProcesses()
		time.Sleep(r.reconcileInterval)
	}
}

func (r *ProcessRegisterer) registerAtlasProcesses() {
	processes, err := r.client.ListProcesses()

	if err != nil {
		metadataScrapeErrors.With(prometheus.Labels{"status": strconv.FormatInt(int64(err.StatusCode), 10)}).Inc()
	}

	currentCollectorKeys := make(map[string]bool, len(processes)) //tracks the existing processes for pruning.
	for _, process := range processes {
		collectorKey := process.ID + process.TypeName
		currentCollectorKeys[collectorKey] = true
	}

	//unregister excess collectors
	for key := range r.collectors {
		//if the collector is no longer needed
		if _, ok := currentCollectorKeys[key]; !ok {
			prometheus.Unregister(r.collectors[key])
			delete(r.collectors, key)
		}
	}

	for _, process := range processes {
		//the way to check for no longer existing hashes is to make a map[ID+TypeName]
		//out of the current list and set difference it to this map.
		collectorKey := process.ID + process.TypeName
		if _, ok := r.collectors[collectorKey]; !ok {
			b := backoff.NewExponentialBackOff()

			b.InitialInterval = time.Second * 5
			b.MaxElapsedTime = 1 * time.Minute

			err := backoff.Retry(func() error {
				collector, err := collector.NewProcessCollector(r.logger, r.client, process)
				if err != nil {
					return err
				}
				r.collectors[collectorKey] = collector
				prometheus.MustRegister(collector)
				return nil
			}, b)

			if err != nil {
				level.Debug(r.logger).Log("msg", "failed collector instantation", "err", err)
			}

		}
	}

}
