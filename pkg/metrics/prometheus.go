package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-chassis/go-archaius"
	"github.com/prometheus/client_golang/prometheus"
)

var onceEnable sync.Once

//PrometheusExporter is a prom exporter for ggs
type PrometheusExporter struct {
	FlushInterval time.Duration
	lc            sync.RWMutex
	lg            sync.RWMutex
	ls            sync.RWMutex
	lh            sync.RWMutex
	counters      map[string]*prometheus.CounterVec
	gauges        map[string]*prometheus.GaugeVec
	summaries     map[string]*prometheus.SummaryVec
	histograms    map[string]*prometheus.HistogramVec
}

//NewPrometheusExporter create a prometheus exporter
func NewPrometheusExporter(options Options) Registry {
	if !archaius.GetBool("ggs.metrics.disableGoRuntimeMetrics", false) {
		onceEnable.Do(func() {
			EnableRunTimeMetrics()
		})
	}
	return &PrometheusExporter{
		FlushInterval: options.FlushInterval,
		lc:            sync.RWMutex{},
		lg:            sync.RWMutex{},
		ls:            sync.RWMutex{},
		lh:            sync.RWMutex{},
		summaries:     make(map[string]*prometheus.SummaryVec),
		counters:      make(map[string]*prometheus.CounterVec),
		gauges:        make(map[string]*prometheus.GaugeVec),
		histograms:    make(map[string]*prometheus.HistogramVec),
	}
}

// EnableRunTimeMetrics enable runtime metrics
func EnableRunTimeMetrics() {
	GetSystemPrometheusRegistry().MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	GetSystemPrometheusRegistry().MustRegister(prometheus.NewGoCollector())
}

//CreateGauge create collector
func (c *PrometheusExporter) CreateGauge(opts GaugeOpts) error {
	gVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: opts.Name,
		Help: opts.Help,
	}, opts.Labels)

	c.lg.Lock()
	defer c.lg.Unlock()
	if _, ok := c.gauges[opts.Name]; ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}
	c.gauges[opts.Name] = gVec
	GetSystemPrometheusRegistry().MustRegister(gVec)
	return nil
}

//GaugeSet set value
func (c *PrometheusExporter) GaugeSet(name string, val float64, labels map[string]string) error {
	c.lg.RLock()
	gVec, ok := c.gauges[name]
	c.lg.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}
	gVec.With(labels).Set(val)
	return nil
}

//CreateCounter create collector
func (c *PrometheusExporter) CreateCounter(opts CounterOpts) error {
	v := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: opts.Name,
		Help: opts.Help,
	}, opts.Labels)

	c.lc.Lock()
	defer c.lc.Unlock()
	if _, ok := c.counters[opts.Name]; ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}
	c.counters[opts.Name] = v
	GetSystemPrometheusRegistry().MustRegister(v)
	return nil
}

//CounterAdd increase value
func (c *PrometheusExporter) CounterAdd(name string, val float64, labels map[string]string) error {
	c.lc.RLock()
	v, ok := c.counters[name]
	c.lc.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}
	v.With(labels).Add(val)
	return nil
}

//CreateSummary create collector
func (c *PrometheusExporter) CreateSummary(opts SummaryOpts) error {
	v := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       opts.Name,
		Help:       opts.Help,
		Objectives: opts.Objectives,
	}, opts.Labels)

	c.ls.Lock()
	defer c.ls.Unlock()
	if _, ok := c.summaries[opts.Name]; ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}
	c.summaries[opts.Name] = v
	GetSystemPrometheusRegistry().MustRegister(v)
	return nil
}

//SummaryObserve set value
func (c *PrometheusExporter) SummaryObserve(name string, val float64, labels map[string]string) error {
	c.ls.RLock()
	v, ok := c.summaries[name]
	c.ls.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}
	v.With(labels).Observe(val)
	return nil
}

//CreateHistogram create collector
func (c *PrometheusExporter) CreateHistogram(opts HistogramOpts) error {
	v := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    opts.Name,
		Help:    opts.Help,
		Buckets: opts.Buckets,
	}, opts.Labels)

	c.lh.Lock()
	defer c.lh.Unlock()
	if _, ok := c.histograms[opts.Name]; ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}
	c.histograms[opts.Name] = v
	GetSystemPrometheusRegistry().MustRegister(v)
	return nil
}

//HistogramObserve set value
func (c *PrometheusExporter) HistogramObserve(name string, val float64, labels map[string]string) error {
	c.ls.RLock()
	v, ok := c.histograms[name]
	c.ls.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}
	v.With(labels).Observe(val)
	return nil
}
func init() {
	registries["prometheus"] = NewPrometheusExporter
}
