# Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build server binary
RUN go build -o server cmd/server/main.go

# Build client binary
RUN go build -o client cmd/client/main.go

# Final stage with Debian slim
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS connections if needed
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/server /app/server
COPY --from=builder /app/client /app/client

# Expose port for server
EXPOSE 8080

# Default command to run server
CMD ["/app/server"]
