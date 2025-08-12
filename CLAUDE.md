# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the masterapp Go project - a Dynamic Electrochemical Impedance Spectroscopy (DEIS) data processor. The application receives real-time voltage U(t) and current I(t) signals every second, performs FFT analysis, calculates impedance Z(t,f) = U(f)/I(f), and sends results to a target application via JSON over HTTP.

## Development Commands

### Build and Run
```bash
go run ./cmd/masterapp                              # Run with default settings  
go run ./cmd/masterapp -target="http://localhost:9000"  # Specify target URL
go run ./cmd/masterapp -rate=2000 -samples=2000    # Custom sample rate and samples per second
go run ./cmd/masterapp -impedance-csv=combined_impedance_data.csv -output=http # Send impedance CSV to target
go run ./cmd/masterapp -direct -circuit=medium -spectra=10 -output=http      # Generate and send 10 medium-complexity spectra
go build -o masterapp ./cmd/masterapp              # Build executable
```

### Testing
```bash
go test ./pkg/...      # Run all module tests
go test -v ./pkg/...   # Run tests with verbose output
go test ./pkg/signal   # Test specific module
```

### Code Quality
```bash
go fmt ./...           # Format code
go vet ./...           # Run static analysis  
goimports -w .         # Format imports (if goimports is installed)
```

## Modular Architecture

The application follows a clean, modular architecture with proper separation of concerns:

### Project Structure
```
masterapp/
├── cmd/
│   └── masterapp/
│       └── main.go                 # Application entry point
├── pkg/                            # Public reusable packages
│   ├── signal/                     # Signal types, validation, generation
│   │   ├── types.go               # Core signal data structures
│   │   ├── interfaces.go          # Signal-related interfaces
│   │   ├── validator.go           # Signal validation logic
│   │   ├── generator.go           # Signal generation for testing
│   │   └── validator_test.go      # Validation tests
│   ├── fft/                       # Fast Fourier Transform processing
│   │   ├── interfaces.go          # FFT processor interface
│   │   ├── processor.go           # FFT implementation
│   │   └── processor_test.go      # FFT tests with known vectors
│   ├── impedance/                 # Impedance calculations
│   │   ├── interfaces.go          # Calculator interface
│   │   └── calculator.go          # Z(f) = U(f)/I(f) calculations
│   ├── network/                   # HTTP communication
│   │   ├── interfaces.go          # Network sender interface
│   │   └── sender.go              # HTTP client with health monitoring
│   ├── receiver/                  # Real-time data reception
│   │   ├── interfaces.go          # Data receiver interface
│   │   └── receiver.go            # Real-time signal processing
│   └── config/                    # Configuration and errors
│       ├── config.go              # Application configuration
│       └── errors.go              # Centralized error types
├── go.mod                         # Go module definition
├── go.sum                         # Go module checksums
├── CLAUDE.md                      # This documentation
└── .gitignore                     # Git ignore patterns
```

## Architecture Notes

### Signal Processing Pipeline
1. **Data Reception**: Receives U(t) and I(t) signals every 1 second via channels
2. **FFT Processing**: Transforms time-domain signals to frequency domain
3. **Impedance Calculation**: Computes Z(f) = U(f)/I(f) for each frequency
4. **JSON Serialization**: Formats results including magnitude and phase
5. **HTTP Transmission**: Sends data to target application via POST requests

### Alternative Input Modes
- **Impedance CSV Mode**: Reads pre-calculated impedance data from CSV files with format: Frequency_Hz,Z_real,Z_imag,Spectrum_Number
- **Direct EIS Generation**: Generates synthetic impedance spectra for various circuit complexities
- **File-based Input**: Processes voltage/current data from CSV files instead of real-time signals

### Key Components
- **Real-time Processing**: Goroutine-based concurrent signal processing
- **FFT Implementation**: Custom radix-2 FFT with DFT fallback for non-power-of-2 lengths
- **Error Handling**: Division by zero protection and signal validation
- **Graceful Shutdown**: SIGINT/SIGTERM handling with WaitGroup synchronization

### Command Line Options
- `-target`: Target URL for sending EIS data (default: http://localhost:8080/eis-data)
- `-rate`: Sample rate in Hz (default: 1000.0)
- `-samples`: Number of samples per second (default: 1000)
- `-impedance-csv`: Path to impedance CSV file with format: Frequency_Hz,Z_real,Z_imag,Spectrum_Number
- `-file`: Use file-based voltage/current data input instead of synthetic data
- `-voltage`: Path to voltage CSV file (default: examples/data/voltage_10s.csv)
- `-current`: Path to current CSV file (default: examples/data/current_10s.csv)
- `-output`: Output mode: 'http' (send via HTTP), 'console' (save JSON files), or 'csv' (save CSV files)
- `-direct`: Use direct EIS generation instead of FFT approach
- `-circuit`: Circuit complexity for direct EIS: 'simple', 'medium', 'complex'
- `-spectra`: Number of spectra to generate for direct EIS mode (default: 5)

## Module Responsibilities

### 🔬 **signal/** - Core Signal Processing Types
- **Types**: Signal, ComplexSignal, ImpedanceData, EISMeasurement
- **Validation**: Comprehensive signal validation with edge case handling
- **Generation**: Realistic signal generation for testing and simulation
- **Interfaces**: Validator and Generator interfaces for dependency injection

### ⚡ **fft/** - Fast Fourier Transform Processing  
- **Algorithm**: Radix-2 FFT with DFT fallback for non-power-of-2 lengths
- **Validation**: Input signal validation and result verification
- **Frequency Extraction**: Positive frequency component extraction
- **Interface**: Clean Processor interface for easy testing and mocking

### 🧮 **impedance/** - Electrochemical Impedance Calculations
- **Core Function**: Z(f) = U(f)/I(f) complex impedance calculation
- **EIS Processing**: Complete electrochemical impedance spectroscopy workflow
- **Error Handling**: Division by zero protection and validation
- **Interface**: Calculator interface with signal compatibility validation

### 🌐 **network/** - HTTP Communication
- **Data Transmission**: JSON-based HTTP POST to target applications
- **Health Monitoring**: Connection health tracking and error recovery
- **Formatting**: Pretty-printed JSON formatting capabilities
- **Interface**: Sender interface with multiple data type support

### 📡 **receiver/** - Real-time Data Reception
- **Timing**: 1-second interval real-time signal processing
- **Context Management**: Graceful shutdown with context cancellation
- **Channel Management**: Buffered channels with overflow protection
- **Interface**: DataReceiver interface with lifecycle management

### ⚙️ **config/** - Configuration and Error Management
- **Configuration**: Application settings with validation
- **Error Types**: Centralized error definitions (ValidationError, ProcessingError, NetworkError)
- **Validation Utilities**: Reusable validation functions across modules
- **Constants**: Shared error constants and configuration limits

## Architecture Benefits

### ✅ **Modularity**
- Each package has a single, well-defined responsibility
- Clear boundaries prevent tight coupling between components
- Easy to understand, test, and maintain individual modules

### ✅ **Reusability** 
- Packages in `pkg/` can be imported by other Go projects
- FFT and signal processing modules are general-purpose
- Clean interfaces enable composition and extension

### ✅ **Testability**
- Interface-based design enables easy mocking
- Each module can be tested in complete isolation
- Clear dependency injection makes integration testing straightforward

### ✅ **Maintainability**
- Changes to one module don't affect others
- Clear API boundaries prevent accidental coupling
- Easy to add new features or replace implementations

## Testing Strategy

```bash
# Module-specific testing
go test ./pkg/signal                # Test signal processing
go test ./pkg/fft                   # Test FFT implementation  
go test ./pkg/impedance             # Test impedance calculations

# Comprehensive testing
go test ./pkg/...                   # All module tests
go test -v ./pkg/...               # Verbose output
go test -cover ./pkg/...           # Coverage analysis
go test -race ./pkg/...            # Race condition detection
```