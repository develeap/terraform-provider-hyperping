# syntax=docker/dockerfile:1
# Multi-stage build for a minimal, non-root hyperping-exporter image.

FROM golang:1.26-alpine AS builder
WORKDIR /build
# Cache module downloads separately from source builds.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /hyperping-exporter ./cmd/hyperping-exporter/

# Use distroless static for a minimal attack surface (no shell, no libc).
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /hyperping-exporter /hyperping-exporter
EXPOSE 9312
USER nonroot:nonroot
ENTRYPOINT ["/hyperping-exporter"]
