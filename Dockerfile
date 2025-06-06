# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/projectflow cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/projectflow .

# Copy web assets
COPY --from=builder /app/web ./web

# Create data directory
RUN mkdir -p data

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV STORAGE_DIR=/app/data

# Run the application
CMD ["./projectflow"]
