# Build Stage
FROM golang:alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and compile statically
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go

# Production Runtime Stage
FROM alpine:3.21

# Install runtime utilities (curl for healthchecks, ca-certificates for TLS)
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root system user and persistent data directory
RUN addgroup -S -g 10001 loadpulse && \
    adduser -S -u 10001 -G loadpulse -h /home/loadpulse loadpulse && \
    mkdir -p /data && \
    chown -R loadpulse:loadpulse /data

# Copy compiled binary from builder
COPY --from=builder /app/server /usr/local/bin/server

# Expose API and WebSocket port
EXPOSE 8080

# Environment defaults
ENV PORT=8080 \
    ENV=production \
    DB_PATH=/data/loadpulse.db

# Mount point for persistent SQLite database
VOLUME ["/data"]

USER loadpulse

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["server"]
