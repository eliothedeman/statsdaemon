package metric

import "sync/atomic"

type Metric interface {
	Value() interface{}
	Update(n interface{}, rate float32)
}

type Counter struct {
	val int64
}

func (c *Counter) Value() interface{} {
	return c.val
}

func (c *Counter) Update(i interface{}, rate float32) {
	n := i.(int64)
	atomic.AddInt64(&c.val, int64(float64(n)*float64(1/rate)))
}

type Timer struct {
	val []uint64
}

func (t *Timer) Value() interface{} {
	return t.val
}

func (t *Timer) Update(i interface{}, rate float32) {
	t.val = append(t.val, i.(uint64))
}

type Gauge struct {
	val float64
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
	val []string
}

func (s *Set) Value() interface{} {
	return s.val
}

func (s *Set) Update(i interface{}, rate float32) {
	s.val = append(s.val, i.(string))
}
