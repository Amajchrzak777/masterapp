# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the masterapp Go project with a simple "hello world" application. The project uses Go 1.24 and consists of a single main.go file with basic structure.

## Development Commands

### Build and Run
```bash
go run main.go          # Run the application directly
go build               # Build the executable
go build -o app        # Build with custom output name
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

- `main.go` - Entry point containing the main function that prints "hello world"
- `go.mod` - Go module definition specifying Go 1.24
- `masterapp.iml` - IntelliJ IDEA module file

## Architecture Notes

This is a basic single-file Go application with no external dependencies beyond the standard library. The current structure is suitable for a simple CLI tool or as a starting point for a larger application.