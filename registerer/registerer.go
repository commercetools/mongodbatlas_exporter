package registerer

import "github.com/prometheus/client_golang/prometheus"

type Registerer interface {
	//Observe starts the Registerer observation loop to track resources and collectors.
	Observe()
	//Collectors returns the map of collectors?
	Collectors() map[string]prometheus.Collector
}
