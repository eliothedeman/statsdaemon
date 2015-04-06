package metric

import (
	"sync"
	"time"
)

type Container struct {
	metrics map[string]*holder
	sync.RWMutex
	reaper ReapStrategy
}

type holder struct {
	m        Metric
	t        time.Time
	orphaned bool
}

func newHolder(m Metric) *holder {
	return &holder{
		m: m,
		t: time.Now(),
	}
}

func NewContainer(r ReapStrategy) *Container {
	return &Container{
		metrics: make(map[string]*holder),
		reaper:  r,
	}
}

type ReapStrategy func(h *holder, c *Container) bool

func KeepAll(h *holder, c *Container) bool {
	return false
}

func ExpireTime(d time.Duration) ReapStrategy {
	return func(h *holder, c *Container) bool {
		return h.t.After(time.Now().Add(d))
	}
}

func ExpireOrphans(h *holder, c *Container) bool {
	if h.m.Name() == "packets_recieved" {
		return false

	}
	return h.orphaned
}

func (c *Container) Add(key string, val Metric) {
	c.Lock()
	c.metrics[key] = newHolder(val)
	c.Unlock()
}

func (c *Container) Update(key string, val interface{}, rate float32) {
	h := c.get(key)
	if h == nil {
		return
	}
	c.Lock()
	h.m.Update(val, rate)
	h.orphaned = false
	c.Unlock()
}

func (c *Container) get(key string) *holder {
	c.RLock()
	h, ok := c.metrics[key]
	c.RUnlock()
	if !ok {
		return nil
	}
	return h
}

func (c *Container) Get(key string) Metric {
	h := c.get(key)
	if h == nil {
		return nil
	}

	return h.m
}

func (c *Container) Remove(key string) {
	delete(c.metrics, key)
}

func (c *Container) List() []Metric {
	c.Lock()
	l := make([]Metric, len(c.metrics))
	i := 0
	for _, v := range c.metrics {
		l[i] = v.m
		v.orphaned = true
		i++
	}
	c.Unlock()
	return l
}

func (c *Container) Reap() {
	c.Lock()
	for k, v := range c.metrics {
		if c.reaper(v, c) {
			c.Remove(k)
		}
	}
	c.Unlock()
}
