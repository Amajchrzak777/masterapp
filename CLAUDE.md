# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the masterapp Go project - a Dynamic Electrochemical Impedance Spectroscopy (DEIS) data processor. The application receives real-time voltage U(t) and current I(t) signals every second, performs FFT analysis, calculates impedance Z(t,f) = U(f)/I(f), and sends results to a target application via JSON over HTTP.

## Development Commands

### Build and Run
```bash
go run main.go                                    # Run with default settings
go run main.go -target="http://localhost:9000"   # Specify target URL
go run main.go -rate=2000 -samples=2000          # Custom sample rate and samples per second
go build -o masterapp                            # Build executable
```

### Testing
```bash
go test ./...          # Run all tests (when tests are added)
go test -v ./...       # Run tests with verbose output
```

### Code Quality
```bash
go fmt ./...           # Format code
go vet ./...           # Run static analysis
goimports -w .         # Format imports (if goimports is installed)
```

## Project Structure

### Core Files
- `main.go` - Entry point with signal processing orchestration and CLI flags
- `types.go` - Core data structures for signals, FFT results, and impedance data
- `interfaces.go` - Interface definitions for dependency injection and testability
- `errors.go` - Centralized error types and error handling utilities
- `validator.go` - Signal validation logic with comprehensive error checking

### Implementation Files
- `receiver.go` - Real-time data receiver with signal generation (DefaultDataReceiver)
- `fft.go` - Fast Fourier Transform implementation for signal processing (DefaultFFTProcessor)
- `impedance.go` - Impedance calculation Z(f) = U(f)/I(f) with error handling (DefaultImpedanceCalculator)
- `sender.go` - HTTP client for sending JSON data to target application (DefaultDataSender)

### Test Files
- `validator_test.go` - Unit tests for validation logic
- `fft_test.go` - Unit tests for FFT processing with known test vectors
- `impedance_test.go` - Unit tests for impedance calculations

### Configuration
- `go.mod` - Go module definition specifying Go 1.24
- `CLAUDE.md` - This documentation file
- `.gitignore` - Git ignore patterns for Go projects

## Architecture Notes

### Signal Processing Pipeline
1. **Data Reception**: Receives U(t) and I(t) signals every 1 second via channels
2. **FFT Processing**: Transforms time-domain signals to frequency domain
3. **Impedance Calculation**: Computes Z(f) = U(f)/I(f) for each frequency
4. **JSON Serialization**: Formats results including magnitude and phase
5. **HTTP Transmission**: Sends data to target application via POST requests

### Key Components
- **Real-time Processing**: Goroutine-based concurrent signal processing
- **FFT Implementation**: Custom radix-2 FFT with DFT fallback for non-power-of-2 lengths
- **Error Handling**: Division by zero protection and signal validation
- **Graceful Shutdown**: SIGINT/SIGTERM handling with WaitGroup synchronization

### Command Line Options
- `-target`: Target URL for sending EIS data (default: http://localhost:8080/eis-data)
- `-rate`: Sample rate in Hz (default: 1000.0)
- `-samples`: Number of samples per second (default: 1000)

## Interface-Based Architecture

The application follows a clean architecture pattern with dependency injection:

### Core Interfaces
- **DataReceiver**: Handles real-time signal reception with context-based cancellation
- **FFTProcessor**: Processes signals using Fast Fourier Transform with validation
- **ImpedanceCalculator**: Calculates complex impedance from voltage/current signals
- **DataSender**: Sends processed data via HTTP with health monitoring
- **SignalValidator**: Validates all signal types with comprehensive error checking
- **SignalGenerator**: Generates realistic test signals for development/testing

### Error Handling Strategy
- **Centralized Error Types**: ValidationError, ProcessingError, NetworkError
- **Error Wrapping**: Maintains error context through the processing pipeline
- **Graceful Degradation**: Invalid signals are logged and skipped, not fatal
- **Health Monitoring**: Components track their health status for diagnostics

### Testing Strategy
- **Unit Tests**: Comprehensive coverage for all components with edge cases
- **Interface Mocking**: Easy to mock dependencies for isolated testing  
- **Known Test Vectors**: FFT and impedance calculations tested against known values
- **Validation Testing**: Edge cases like NaN, infinity, and zero values covered
- **Error Path Testing**: All error conditions and error types tested

### Dependency Injection Benefits
- **Testability**: Easy to mock components for unit testing
- **Flexibility**: Can swap implementations without changing calling code
- **Maintainability**: Clear separation of concerns between components
- **Extensibility**: Easy to add new implementations (e.g., different data sources)

## Testing Commands

```bash
go test -v ./...                    # Run all tests with verbose output
go test -cover ./...                # Run tests with coverage analysis
go test -race ./...                 # Run tests with race condition detection
go test -run=TestFFT ./...          # Run only FFT-related tests
go test -bench=. ./...              # Run benchmarks (if any)
```