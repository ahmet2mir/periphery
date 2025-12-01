# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o periphery .

# Final stage
FROM scratch

# Copy certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /build/periphery /periphery

# Create non-root user
USER 65534:65534

# Set entrypoint
ENTRYPOINT ["/periphery"]

# Default command
CMD ["--config", "/etc/periphery/config.yaml"]

# Labels
LABEL org.opencontainers.image.title="Periphery" \
      org.opencontainers.image.description="BGP anycast service with Kubernetes-inspired health probes" \
      org.opencontainers.image.url="https://github.com/ahmet2mir/periphery" \
      org.opencontainers.image.source="https://github.com/ahmet2mir/periphery" \
      org.opencontainers.image.licenses="Apache-2.0"
