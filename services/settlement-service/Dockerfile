# -------------------------
# Builder stage
# -------------------------
FROM --platform=$BUILDPLATFORM golang:1.24.5-alpine3.21 AS builder

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build static Linux binary for target platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /settlement-service .

# -------------------------
# Runtime stage
# -------------------------
FROM alpine:3.18

WORKDIR /

# Install CA certificates
RUN apk add --no-cache ca-certificates

# Copy binary from builder stage
COPY --from=builder /settlement-service /settlement-service

# Make binary executable
RUN chmod +x /settlement-service

# Expose gRPC port
EXPOSE 50053

# Non-root user
RUN adduser -D appuser
USER appuser

# Run the service
ENTRYPOINT ["/settlement-service"]
