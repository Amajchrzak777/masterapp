package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	ossignal "os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/adam/masterapp/pkg/config"
	"github.com/adam/masterapp/pkg/impedance"
	"github.com/adam/masterapp/pkg/network"
	"github.com/adam/masterapp/pkg/receiver"
	"github.com/adam/masterapp/pkg/signal"
)

func main() {
	var (
		targetURL     = flag.String("target", "http://localhost:8080/eis-data", "Target URL for sending EIS data")
		sampleRate    = flag.Float64("rate", 200000.0, "Sample rate in Hz")
		samplesPerSec = flag.Int("samples", 200, "Number of samples per second")
		useFileData   = flag.Bool("file", false, "Use file-based data input instead of synthetic data")
		voltageFile   = flag.String("voltage", "examples/data/voltage_10s.csv", "Path to voltage CSV file")
		currentFile   = flag.String("current", "examples/data/current_10s.csv", "Path to current CSV file")
		outputMode    = flag.String("output", "console", "Output mode: 'http' (send via HTTP), 'console' (print JSON to files), or 'csv' (print CSV format)")
	)
	flag.Parse()

	// Create and validate configuration
	cfg := &config.Config{
		TargetURL:        *targetURL,
		SampleRate:       *sampleRate,
		SamplesPerSecond: *samplesPerSec,
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Println("Starting Dynamic Electrochemical Impedance Spectroscopy (DEIS) processor")
	log.Printf("Target URL: %s", cfg.TargetURL)
	log.Printf("Sample rate: %.1f Hz", cfg.SampleRate)
	log.Printf("Samples per second: %d", cfg.SamplesPerSecond)

	// Initialize data receiver based on mode
	var dataReceiver receiver.DataReceiver
	var err error

	if *useFileData {
		log.Printf("Using file-based data input:")
		log.Printf("  Voltage file: %s", *voltageFile)
		log.Printf("  Current file: %s", *currentFile)
		dataReceiver, err = receiver.NewFileReceiver(*voltageFile, *currentFile, cfg.SampleRate)
		if err != nil {
			log.Fatalf("Failed to create file receiver: %v", err)
		}
	} else {
		log.Println("Using synthetic data generation")
		dataReceiver = receiver.NewReceiver(cfg.SampleRate, cfg.SamplesPerSecond)
	}

	// Initialize other components
	calculator := impedance.NewCalculator()
	sender := network.NewSender(cfg.TargetURL)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	signalChan := make(chan os.Signal, 1)
	ossignal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(2)

	// Start data receiver
	go func() {
		defer wg.Done()
		if err := dataReceiver.StartReceiving(ctx); err != nil && err != context.Canceled {
			log.Printf("Data receiver error: %v", err)
		}
	}()

	// Start signal processor
	go func() {
		defer wg.Done()
		processSignals(ctx, dataReceiver, calculator, sender, *outputMode)
	}()

	// Wait for shutdown signal
	<-signalChan
	log.Println("Shutdown signal received, stopping...")

	// Cancel context to stop all goroutines
	cancel()

	// Stop receiver
	if err := dataReceiver.Stop(); err != nil {
		log.Printf("Error stopping receiver: %v", err)
	}

	wg.Wait()
	log.Println("DEIS processor stopped")
}

func processSignals(ctx context.Context, dataReceiver receiver.DataReceiver, calculator impedance.Calculator, sender network.Sender, outputMode string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Signal processor stopping due to context cancellation")
			return
		case voltageSignal := <-dataReceiver.GetVoltageChannel():
			select {
			case currentSignal := <-dataReceiver.GetCurrentChannel():
				measurement, err := calculator.ProcessEISMeasurement(voltageSignal, currentSignal)
				if err != nil {
					log.Printf("Error processing EIS measurement: %v", err)
					continue
				}

				if outputMode == "console" {
					// Output processed data to console as JSON files
					printEISMeasurement(measurement, "json")
				} else if outputMode == "csv" {
					// Output processed data as CSV
					printEISMeasurement(measurement, "csv")
				} else {
					// Send via HTTP
					if err := sender.SendEISMeasurement(measurement); err != nil {
						log.Printf("Error sending EIS measurement: %v", err)

						// Check if sender is unhealthy and log warning
						if !sender.IsHealthy() {
							log.Printf("Warning: Data sender is unhealthy")
						}
					}
				}
			default:
				log.Println("Warning: No current signal available for voltage signal")
			}
		}
	}
}

var measurementCounter int

func printEISMeasurement(measurement interface{}, format string) {
	measurementCounter++

	if format == "csv" {
		printCSVMeasurement(measurement)
		return
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Join("output", "json")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		return
	}

	// Generate filename with timestamp and counter
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("eis_measurement_%s_%03d.json", timestamp, measurementCounter)
	filePath := filepath.Join(outputDir, filename)

	// Marshal JSON with pretty formatting
	jsonData, err := json.MarshalIndent(measurement, "", "  ")
	if err != nil {
		log.Printf("Error marshaling EIS measurement: %v", err)
		return
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Printf("Error writing EIS measurement to file %s: %v", filePath, err)
		return
	}

	log.Printf("EIS measurement saved to: %s", filePath)
}

func printCSVMeasurement(measurement interface{}) {
	eisMeasurement, ok := measurement.(signal.EISMeasurement)
	if !ok {
		log.Printf("Error: Invalid measurement type for CSV output")
		return
	}

	// Create CSV output directory if it doesn't exist
	outputDir := filepath.Join("output", "csv")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating CSV output directory: %v", err)
		return
	}

	// Generate CSV filename with timestamp and counter
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("eis_measurement_%s_%03d.csv", timestamp, measurementCounter)
	filePath := filepath.Join(outputDir, filename)

	// Create CSV file
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating CSV file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	// Write CSV header
	fmt.Fprintf(file, "frequency,real,imag\n")
	
	// Write impedance data
	for _, point := range eisMeasurement {
		fmt.Fprintf(file, "%.6g,%.6f,%.6f\n", point.Frequency, point.Real, point.Imag)
	}

	log.Printf("EIS measurement CSV saved to: %s", filePath)
}
