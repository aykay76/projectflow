package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

// RequestIDKey is the context key for request IDs
const RequestIDKey ContextKey = "request_id"

// RequestIDHeader is the HTTP header name for request IDs
const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from client)
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			// Generate a new request ID
			requestID = generateRequestID()
		}

		// Add request ID to response header
		w.Header().Set(RequestIDHeader, requestID)

		// Add request ID to request context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Log the request with request ID
		slog.InfoContext(ctx, "Request started",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", requestID,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		// Continue with the next handler
		next.ServeHTTP(w, r)

		// Log request completion
		slog.InfoContext(ctx, "Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", requestID,
		)
	})
}

// generateRequestID creates a random 16-character hex string
func generateRequestID() string {
	bytes := make([]byte, 8) // 8 bytes = 16 hex characters
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		slog.Warn("Failed to generate random request ID, using fallback", "error", err)
		return "fallback-" + hex.EncodeToString([]byte("backup"))[:8]
	}
	return hex.EncodeToString(bytes)
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
