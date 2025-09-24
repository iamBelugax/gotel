package gotel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// MetricRegistry provides a convenient way to create and manage custom metrics.
type MetricRegistry struct {
	prefix         string
	meter          metric.Meter
	counters       map[string]metric.Int64Counter
	histograms     map[string]metric.Float64Histogram
	upDownCounters map[string]metric.Int64UpDownCounter
	gauges         map[string]metric.Int64ObservableGauge
	mu             sync.RWMutex
}

// NewMetricRegistry creates a new MetricRegistry.
func NewMetricRegistry(meter metric.Meter, prefix string) *MetricRegistry {
	return &MetricRegistry{
		meter:          meter,
		prefix:         prefix,
		counters:       make(map[string]metric.Int64Counter),
		histograms:     make(map[string]metric.Float64Histogram),
		upDownCounters: make(map[string]metric.Int64UpDownCounter),
		gauges:         make(map[string]metric.Int64ObservableGauge),
	}
}

// Counter creates or returns an existing counter metric.
func (m *MetricRegistry) Counter(
	name, description string, options ...metric.Int64CounterOption,
) (metric.Int64Counter, error) {
	metricName := m.generateMetricName(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if counter, exists := m.counters[metricName]; exists {
		return counter, nil
	}

	options = append(options, metric.WithDescription(description))
	counter, err := m.meter.Int64Counter(metricName, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create counter %s: %w", metricName, err)
	}

	m.counters[metricName] = counter
	return counter, nil
}

// Histogram creates or returns an existing histogram metric.
func (m *MetricRegistry) Histogram(name, description string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	metricName := m.generateMetricName(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if histogram, exists := m.histograms[metricName]; exists {
		return histogram, nil
	}

	options = append(options, metric.WithDescription(description))
	histogram, err := m.meter.Float64Histogram(metricName, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create histogram %s: %w", metricName, err)
	}

	m.histograms[metricName] = histogram
	return histogram, nil
}

// UpDownCounter creates or returns an existing up/down counter metric.
func (m *MetricRegistry) UpDownCounter(name, description string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	metricName := m.generateMetricName(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if upDownCounter, exists := m.upDownCounters[metricName]; exists {
		return upDownCounter, nil
	}

	options = append(options, metric.WithDescription(description))
	upDownCounter, err := m.meter.Int64UpDownCounter(metricName, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create up/down counter %s: %w", metricName, err)
	}

	m.upDownCounters[metricName] = upDownCounter
	return upDownCounter, nil
}

// Gauge creates or returns an existing observable gauge metric.
func (m *MetricRegistry) Gauge(
	name string,
	description string,
	callback func(context.Context, metric.Int64Observer) error,
	options ...metric.Int64ObservableGaugeOption,
) (metric.Int64ObservableGauge, error) {
	metricName := m.generateMetricName(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if gauge, exists := m.gauges[metricName]; exists {
		return gauge, nil
	}

	options = append(options, metric.WithDescription(description), metric.WithInt64Callback(callback))
	gauge, err := m.meter.Int64ObservableGauge(metricName, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gauge %s: %w", metricName, err)
	}

	m.gauges[metricName] = gauge
	return gauge, nil
}

func (m *MetricRegistry) generateMetricName(name string) string {
	if m.prefix == "" {
		return name
	}
	return fmt.Sprintf("%s_%s", m.prefix, name)
}

// CommonMetrics provides a set of commonly used metrics for web applications.
type CommonMetrics struct {
	registry            *MetricRegistry
	HTTPRequestsTotal   metric.Int64Counter
	HTTPRequestDuration metric.Float64Histogram
	HTTPActiveRequests  metric.Int64UpDownCounter
	DBConnectionsActive metric.Int64UpDownCounter
	DBQueriesTotal      metric.Int64Counter
	DBQueryDuration     metric.Float64Histogram
	ErrorsTotal         metric.Int64Counter
	StartTime           metric.Int64ObservableGauge
}

// NewCommonMetrics creates a new set of common metrics.
func NewCommonMetrics(registry *MetricRegistry) (*CommonMetrics, error) {
	cm := &CommonMetrics{registry: registry}
	var err error

	cm.HTTPRequestsTotal, err = registry.Counter("http_requests_total", "Total number of HTTP requests")
	if err != nil {
		return nil, err
	}

	cm.HTTPRequestDuration, err = registry.Histogram(
		"http_request_duration_seconds",
		"Duration of HTTP requests in seconds",
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
	)
	if err != nil {
		return nil, err
	}

	cm.HTTPActiveRequests, err = registry.UpDownCounter("http_active_requests", "Number of active HTTP requests")
	if err != nil {
		return nil, err
	}

	cm.DBConnectionsActive, err = registry.UpDownCounter("db_connections_active", "Number of active database connections")
	if err != nil {
		return nil, err
	}

	cm.DBQueriesTotal, err = registry.Counter("db_queries_total", "Total number of database queries")
	if err != nil {
		return nil, err
	}

	cm.DBQueryDuration, err = registry.Histogram(
		"db_query_duration_seconds",
		"Duration of database queries in seconds",
		metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1),
	)
	if err != nil {
		return nil, err
	}

	cm.ErrorsTotal, err = registry.Counter("errors_total", "Total number of errors")
	if err != nil {
		return nil, err
	}

	startTime := time.Now().Unix()
	cm.StartTime, err = registry.Gauge(
		registry.generateMetricName("start_time_seconds"),
		"Unix timestamp of when the application started",
		func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(startTime)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cm, nil
}
