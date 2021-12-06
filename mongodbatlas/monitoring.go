package mongodbatlas

import "github.com/prometheus/client_golang/prometheus"

var (
	requestCounter *prometheus.CounterVec
)

//when the package initializes HTTP client metrics are set here.
func init() {
	requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mongodbatlas_http_requests_total",
		Help: "the number of Atlas HTTP API requests made since startup",
	},
		[]string{"code", "method"},
	)

	prometheus.MustRegister(requestCounter)
}
