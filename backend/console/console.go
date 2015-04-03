package console

import (
	"fmt"
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
		fmt.Printf("%+v\n", all[i])
	}
	return nil
}

func (c *Console) Init(i interface{}) error {
	return nil
}

func (c *Console) ConfigStruct() interface{} {
	return &struct{}{}
}
