// Package metrics provides Prometheus-compatible metrics collection for the ProjectFlow application.
// It tracks HTTP request metrics including count, duration, status codes, and business metrics.
package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the metrics collectors for the application
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec
	
	// Business metrics
	tasksTotal        *prometheus.CounterVec
	tasksActive       prometheus.Gauge
	storageOperations *prometheus.CounterVec
}

// NewMetrics creates and registers all metrics with Prometheus
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP request count by method, endpoint, and status code
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		// HTTP request duration histogram
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
			},
			[]string{"method", "endpoint"},
		),

		// HTTP response size histogram
		httpResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "Size of HTTP responses in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
			},
			[]string{"method", "endpoint"},
		),

		// Business metrics - total tasks by operation type
		tasksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "projectflow_tasks_total",
				Help: "Total number of tasks processed by operation",
			},
			[]string{"operation", "status"},
		),

		// Active tasks gauge
		tasksActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "projectflow_tasks_active",
				Help: "Number of currently active tasks",
			},
		),

		// Storage operation metrics
		storageOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "projectflow_storage_operations_total",
				Help: "Total number of storage operations",
			},
			[]string{"operation", "result"},
		),
	}
}

// RecordHTTPRequest records metrics for an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration, responseSize int64) {
	statusStr := strconv.Itoa(statusCode)
	
	// Increment request counter
	m.httpRequestsTotal.WithLabelValues(method, endpoint, statusStr).Inc()
	
	// Record request duration
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
	
	// Record response size
	m.httpResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
}

// RecordTaskOperation records metrics for task operations
func (m *Metrics) RecordTaskOperation(operation, status string) {
	m.tasksTotal.WithLabelValues(operation, status).Inc()
}

// SetActiveTasks sets the current number of active tasks
func (m *Metrics) SetActiveTasks(count float64) {
	m.tasksActive.Set(count)
}

// RecordStorageOperation records metrics for storage operations
func (m *Metrics) RecordStorageOperation(operation, result string) {
	m.storageOperations.WithLabelValues(operation, result).Inc()
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const metricsKey contextKey = "metrics"

// WithMetrics adds metrics to the context
func WithMetrics(ctx context.Context, metrics *Metrics) context.Context {
	return context.WithValue(ctx, metricsKey, metrics)
}

// FromContext retrieves metrics from the context
func FromContext(ctx context.Context) (*Metrics, bool) {
	metrics, ok := ctx.Value(metricsKey).(*Metrics)
	return metrics, ok
}
