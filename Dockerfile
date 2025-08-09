# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o masterapp cmd/masterapp/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/masterapp .

# Copy necessary directories
COPY --from=builder /app/examples ./examples
COPY --from=builder /app/output ./output

# Create output directories if they don't exist
RUN mkdir -p output/csv output/json


# Default command (can be overridden)
CMD ["./masterapp"]