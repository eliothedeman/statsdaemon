package console

import (
	"log"
	"time"

	"github.com/eliothedeman/statsdaemon/backend"
	"github.com/eliothedeman/statsdaemon/metric"
)

func init() {
	backend.LoadBackend("console", &Console{})
}

type Console struct {
}

func (c *Console) Submit(all []metric.Metric, deadline time.Time) error {
	for i := range all {
		log.Printf("console - metric: %s value: %v", all[i].Name(), all[i].Value())
	}
	return nil
}

func (c *Console) Init(i interface{}) error {
	return nil
}

func (c *Console) ConfigStruct() interface{} {
	return &struct{}{}
}
