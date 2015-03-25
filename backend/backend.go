package backend

import (
	"time"

	"github.com/eliothedeman/statsdaemon/metric"
)

var (
	Backends = make(map[string]Backend)
)

func LoadBackend(name string, b Backend) {
	Backends[name] = b
}

type Backend interface {
	Submit([]metric.Metric, time.Time) error
	Init(interface{}) error
	ConfigStruct() interface{}
}
