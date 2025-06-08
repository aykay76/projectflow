// Package middleware provides HTTP middleware functions for metrics collection.
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/aykay76/projectflow/internal/logger"
	"github.com/aykay76/projectflow/internal/metrics"
)

// responseWriter wraps http.ResponseWriter to capture response metrics
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

// MetricsMiddleware creates middleware that collects HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Add metrics to context for handlers to use
			ctx := metrics.WithMetrics(r.Context(), m)
			r = r.WithContext(ctx)
			
			// Wrap response writer to capture metrics
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // Default status code
				size:           0,
			}
			
			// Get endpoint path for metrics (normalize dynamic paths)
			endpoint := normalizeEndpoint(r.URL.Path)
			
			// Log request start with metrics context
			requestID := GetRequestID(ctx)
			logger.InfoContext(ctx, "HTTP request started",
				"method", r.Method,
				"endpoint", endpoint,
				"request_id", requestID,
			)
			
			// Process request
			next.ServeHTTP(rw, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Record metrics
			m.RecordHTTPRequest(r.Method, endpoint, rw.statusCode, duration, rw.size)
			
			// Log request completion with metrics
			logger.InfoContext(ctx, "HTTP request completed",
				"method", r.Method,
				"endpoint", endpoint,
				"status_code", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"response_size", rw.size,
				"request_id", requestID,
			)
		})
	}
}

// normalizeEndpoint normalizes URL paths for consistent metrics labeling
// This prevents high cardinality metrics from dynamic path parameters
func normalizeEndpoint(path string) string {
	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}
	
	// Normalize common patterns
	switch {
	case path == "/health", path == "/ready", path == "/metrics":
		return path
	case strings.HasPrefix(path, "/tasks/") && len(path) > 7:
		// Normalize /tasks/{id} to /tasks/:id
		return "/tasks/:id"
	case strings.HasPrefix(path, "/projects/") && len(path) > 10:
		// Normalize /projects/{id} to /projects/:id  
		return "/projects/:id"
	case path == "/tasks":
		return path
	case path == "/projects":
		return path
	default:
		// For unknown paths, use the first two segments
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) >= 2 {
			return "/" + parts[0] + "/" + parts[1]
		} else if len(parts) == 1 {
			return "/" + parts[0]
		}
		return "/"
	}
}
