# Build Stage
FROM golang:alpine AS builder

WORKDIR /build

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/demo-api ./main.go

# Production Runtime Stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl

RUN addgroup -S -g 10002 demoapi && \
    adduser -S -u 10002 -G demoapi demoapi

COPY --from=builder /app/demo-api /usr/local/bin/demo-api

EXPOSE 8081

ENV PORT=8081

USER demoapi

HEALTHCHECK --interval=10s --timeout=3s --start-period=3s --retries=3 \
  CMD curl -f http://localhost:8081/health || exit 1

ENTRYPOINT ["demo-api"]
