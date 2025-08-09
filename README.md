# MasterApp - DEIS Data Processor

Dynamic Electrochemical Impedance Spectroscopy (DEIS) data processor that receives real-time voltage and current signals, performs FFT analysis, and sends impedance results to calculation services via HTTP.

## Quick Start

```bash
# Run with default settings (sends to calculation service on localhost:8080)
go run ./cmd/masterapp

# Run with custom target URL
go run ./cmd/masterapp -target="http://localhost:8080/eis-data"

# Run with custom sample rate and samples
go run ./cmd/masterapp -rate=2000 -samples=2000
```

## Docker

```bash
# Run entire stack with docker-compose
docker-compose up

# Run specific services
docker-compose up mockinput goimpcore webplot
```

## Build

```bash
go build -o masterapp ./cmd/masterapp
```

## Test

```bash
go test ./pkg/...
```