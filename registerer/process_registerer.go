package registerer

import (
	"mongodbatlas_exporter/collector"
	a "mongodbatlas_exporter/mongodbatlas"
	"strconv"
	"time"

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
	collectors map[string]prometheus.Collector
	ticker     *time.Ticker
	client     a.Client
	logger     log.Logger
}

func NewProcessRegisterer(logger log.Logger, c a.Client) *ProcessRegisterer {
	return &ProcessRegisterer{
		client:     c,
		logger:     logger,
		ticker:     time.NewTicker(time.Minute),
		collectors: make(map[string]prometheus.Collector),
	}
}

func (r *ProcessRegisterer) Observe() {
	r.ticker = time.NewTicker(time.Minute)

	//Register on first call.
	r.registerAtlasProcesses()
	//Keep the register up to date.
	for range r.ticker.C {
		r.registerAtlasProcesses()
	}
}

func (r *ProcessRegisterer) registerAtlasProcesses() {
	processes, err := r.client.ListProcesses()

	if err != nil {
		metadataScrapeErrors.With(prometheus.Labels{"status": strconv.FormatInt(int64(err.StatusCode), 10)}).Inc()
	}

	for _, process := range processes {
		collector, err := collector.NewProcessCollector(r.logger, r.client, process)

		if err != nil {
			level.Debug(r.logger).Log("msg", "failed collector instantation", "err", err)
		}
		//the way to check for no longer existing hashes is to make a map[ID+TypeName]
		//out of the current list and set difference it to this map.

		collectorKey := process.ID + process.TypeName
		if _, ok := r.collectors[collectorKey]; !ok {
			r.collectors[process.ID+process.TypeName] = collector
			prometheus.MustRegister(collector)
		}
	}
}
