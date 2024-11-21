package metrics

import (
	"fmt"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var _ Metrics = &metrics{}

// metrics is a standard implementation of the Metrics interface via prometheus.
type metrics struct {
	// logger is the logger used to log messages.
	logger logging.Logger

	// config is the configuration for the metrics.
	config *Config

	// registry is the prometheus registry used to report metrics.
	registry *prometheus.Registry

	// counterVecMap is a map of metric names to prometheus counter vectors.
	// These are used to create new counter metrics.
	counterVecMap map[string]*prometheus.CounterVec

	// summaryVecMap is a map of metric names to prometheus summary vectors.
	// These are used to create new latency metrics.
	summaryVecMap map[string]*prometheus.SummaryVec

	// gaugeVecMap is a map of metric names to prometheus gauge vectors.
	// These are used to create new gauge metrics.
	gaugeVecMap map[string]*prometheus.GaugeVec

	// A map from metricID to Metric instance. If a metric is requested but that metric
	// already exists, the existing metric will be returned instead of a new one being created.
	metricMap map[metricID]Metric

	// creationLock is a lock used to ensure that metrics are not created concurrently.
	creationLock sync.Mutex

	// server is the metrics server
	server *http.Server
}

// NewMetrics creates a new Metrics instance.
func NewMetrics(logger logging.Logger, config *Config) Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())

	logger.Infof("Starting metrics server at port %d", config.HTTPPort)
	addr := fmt.Sprintf(":%d", config.HTTPPort)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	))
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &metrics{
		logger:        logger,
		config:        config,
		registry:      reg,
		counterVecMap: make(map[string]*prometheus.CounterVec),
		summaryVecMap: make(map[string]*prometheus.SummaryVec),
		gaugeVecMap:   make(map[string]*prometheus.GaugeVec),
		metricMap:     make(map[metricID]Metric),
		server:        server,
	}
}

// metricID is a unique identifier for a metric.
type metricID struct {
	name  string
	label string
}

var legalCharactersRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// newMetricID creates a new metricID instance.
func newMetricID(name string, label string) (metricID, error) {
	if !legalCharactersRegex.MatchString(name) {
		return metricID{}, fmt.Errorf("invalid metric name: %s", name)
	}
	if label != "" && !legalCharactersRegex.MatchString(label) {
		return metricID{}, fmt.Errorf("invalid metric label: %s", label)
	}
	return metricID{
		name:  name,
		label: label,
	}, nil
}

// String returns a string representation of the metricID.
func (i *metricID) String() string {
	if i.label != "" {
		return fmt.Sprintf("%s:%s", i.name, i.label)
	}
	return i.name
}

// Start starts the metrics server.
func (m *metrics) Start() {
	go func() {
		err := m.server.ListenAndServe()
		if err != nil && !strings.Contains(err.Error(), "http: Server closed") {
			m.logger.Errorf("metrics server error: %v", err)
		}
	}()
}

// Stop stops the metrics server.
func (m *metrics) Stop() {
	err := m.server.Close()
	if err != nil {
		m.logger.Errorf("error stopping metrics server: %v", err)
	}
}

// NewLatencyMetric creates a new LatencyMetric instance.
func (m *metrics) NewLatencyMetric(name string, label string, quantiles ...*Quantile) (LatencyMetric, error) {
	m.creationLock.Lock()
	defer m.creationLock.Unlock()

	id, err := newMetricID(name, label)
	if err != nil {
		return nil, err
	}

	preExistingMetric, ok := m.metricMap[id]
	if ok {
		return preExistingMetric.(LatencyMetric), nil
	}

	if m.isBlacklisted(id) {
		metric := newLatencyMetric(name, label, nil)
		m.metricMap[id] = metric
	}

	objectives := make(map[float64]float64, len(quantiles))
	for _, q := range quantiles {
		objectives[q.Quantile] = q.Error
	}

	vec, ok := m.summaryVecMap[name]
	if !ok {
		vec = promauto.With(m.registry).NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  m.config.Namespace,
				Name:       name,
				Objectives: objectives,
			},
			[]string{"label"},
		)
		m.summaryVecMap[name] = vec
	}

	metric := newLatencyMetric(name, label, vec)
	m.metricMap[id] = metric
	return metric, nil
}

// NewCountMetric creates a new CountMetric instance.
func (m *metrics) NewCountMetric(name string, label string) (CountMetric, error) {
	m.creationLock.Lock()
	defer m.creationLock.Unlock()

	id, err := newMetricID(name, label)
	if err != nil {
		return nil, err
	}

	preExistingMetric, ok := m.metricMap[id]
	if ok {
		return preExistingMetric.(CountMetric), nil
	}

	if m.isBlacklisted(id) {
		metric := newLatencyMetric(name, label, nil)
		m.metricMap[id] = metric
	}

	vec, ok := m.counterVecMap[name]
	if !ok {
		vec = promauto.With(m.registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: m.config.Namespace,
				Name:      name,
			},
			[]string{"label"},
		)
		m.counterVecMap[name] = vec
	}

	metric := newCountMetric(name, label, vec)
	m.metricMap[id] = metric

	return metric, nil
}

// NewGaugeMetric creates a new GaugeMetric instance.
func (m *metrics) NewGaugeMetric(name string, label string) (GaugeMetric, error) {
	m.creationLock.Lock()
	defer m.creationLock.Unlock()

	id, err := newMetricID(name, label)
	if err != nil {
		return nil, err
	}

	preExistingMetric, ok := m.metricMap[id]
	if ok {
		return preExistingMetric.(GaugeMetric), nil
	}

	if m.isBlacklisted(id) {
		metric := newLatencyMetric(name, label, nil)
		m.metricMap[id] = metric
	}

	vec, ok := m.gaugeVecMap[name]
	if !ok {
		vec = promauto.With(m.registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: m.config.Namespace,
				Name:      name,
			},
			[]string{"label"},
		)
		m.gaugeVecMap[name] = vec
	}

	metric := newGaugeMetric(name, label, vec)
	m.metricMap[id] = metric

	return metric, nil
}

// isBlacklisted returns true if the metric name is blacklisted.
func (m *metrics) isBlacklisted(id metricID) bool {
	metric := id.String()

	if m.config.MetricsBlacklist != nil {
		for _, blacklisted := range m.config.MetricsBlacklist {
			if metric == blacklisted {
				return true
			}
		}
	}
	if m.config.MetricsFuzzyBlacklist != nil {
		for _, blacklisted := range m.config.MetricsFuzzyBlacklist {
			if strings.Contains(metric, blacklisted) {
				return true
			}
		}
	}
	return false
}
