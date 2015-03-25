package metric

import (
	"sort"
	"sync/atomic"
)

const (
	MAX_TIMERS_KEPT = 60
)

type Metric interface {
	Name() string
	Value() interface{}
	Update(n interface{}, rate float32)
}

type Counter struct {
	name string
	val  int64
}

func NewCounter(name string) *Counter {
	return &Counter{
		name: name,
	}
}

func (c *Counter) Name() string {
	return c.name
}

func (c *Counter) Value() interface{} {
	return c.val
}

func (c *Counter) Update(i interface{}, rate float32) {
	n := i.(int)
	atomic.AddInt64(&c.val, int64(float64(n)*float64(1/rate)))
}

type Timer struct {
	name string
	val  chan float64
}

func NewTimer(name string) *Timer {
	return &Timer{
		name: name,
		val:  make(chan float64, MAX_TIMERS_KEPT),
	}
}

func (t *Timer) Name() string {
	return t.name
}

func chanToSlice(c chan float64) []float64 {
	arr := make([]float64, len(c))
	var tmp float64
	for i := 0; i < len(arr); i++ {
		tmp = <-c
		arr[i] = tmp
		c <- tmp
	}
	return arr
}

func (t *Timer) Value() interface{} {
	// reduce to []float64
	arr := chanToSlice(t.val)

	sort.Float64s(arr)

	return map[string]float64{
		"min":   arr[0],
		"max":   arr[len(arr)-1],
		"mean":  avg(arr),
		"count": float64(len(arr)),
	}
}

func (t *Timer) Update(i interface{}, rate float32) {
	// if we are at capasity, throw away the oldest value
	if len(t.val) == cap(t.val) {
		<-t.val
	}

	t.val <- i.(float64)
}

type Gauge struct {
	name string
	val  float64
}

func NewGauge(name string) *Gauge {
	return &Gauge{
		name: name,
	}
}

func (g *Gauge) Name() string {
	return g.name
}

type GaugeData struct {
	Relative bool
	Negative bool
	Value    float64
}

func (g *Gauge) Value() interface{} {
	return g.val
}

func (g *Gauge) Update(i interface{}, rate float32) {
	d := i.(GaugeData)
	if d.Relative {
		if d.Negative {
			g.val -= d.Value
		} else {
			g.val += d.Value
		}
	} else {
		g.val = d.Value
	}
}

type Set struct {
	name string
	val  []string
}

func NewSet(name string) *Set {
	return &Set{
		name: name,
		val:  make([]string, 0, 2),
	}
}

func (s *Set) Name() string {
	return s.name
}

func (s *Set) Value() interface{} {
	return s.val
}

func (s *Set) Update(i interface{}, rate float32) {
	s.val = append(s.val, i.(string))
}
