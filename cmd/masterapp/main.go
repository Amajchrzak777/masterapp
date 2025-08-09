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
	eisgen "github.com/adam/masterapp/pkg/impedance"
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
		useDirectEIS  = flag.Bool("direct", false, "Use direct EIS generation (like Python impedance_data.csv) instead of FFT approach")
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

	// Check if using direct EIS generation mode
	if *useDirectEIS {
		log.Println("Using direct EIS generation (Python impedance_data.csv approach)")
		runDirectEISMode(cfg, *outputMode)
		return
	}

	// Initialize data receiver based on mode (traditional FFT approach)
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
				impedanceData, err := calculator.CalculateImpedance(voltageSignal, currentSignal)
				if err != nil {
					log.Printf("Error calculating impedance: %v", err)
					continue
				}

				if outputMode == "console" {
					// Convert to EISMeasurement for file output
					measurement, err := calculator.ProcessEISMeasurement(voltageSignal, currentSignal)
					if err != nil {
						log.Printf("Error processing EIS measurement: %v", err)
						continue
					}
					printEISMeasurement(measurement, "json")
				} else if outputMode == "csv" {
					// Convert to EISMeasurement for CSV output
					measurement, err := calculator.ProcessEISMeasurement(voltageSignal, currentSignal)
					if err != nil {
						log.Printf("Error processing EIS measurement: %v", err)
						continue
					}
					printEISMeasurement(measurement, "csv")
				} else {
					// Send impedance data with voltage via HTTP
					if err := sender.SendImpedanceData(impedanceData); err != nil {
						log.Printf("Error sending impedance data: %v", err)

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

// runDirectEISMode runs the direct EIS generation mode (like Python code)
func runDirectEISMode(cfg *config.Config, outputMode string) {
	log.Println("Starting Direct EIS generation mode")
	log.Println("Replicating Python impedance_data.csv approach")
	
	// Create EIS generator with same parameters as Python code
	eisGenerator := eisgen.NewEISGenerator()
	params := eisGenerator.GetDefaultParameters()
	
	log.Printf("Circuit parameters: Rs=%.1f, Rct_initial=%.1f, Q=%.2e, n=%.2f", 
		params.Rs, params.RctInitial, params.Q, params.N)
		
	// Create network sender
	sender := network.NewSender(cfg.TargetURL)
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	signalChan := make(chan os.Signal, 1)
	ossignal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Create single output file that gets overwritten each run
	outputFilePath := "generated_eis_data.csv"
	if _, err := os.Stat("/root/data"); err == nil {
		// Running in Docker container
		outputFilePath = "/root/data/generated_eis_data.csv"
	}
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()
	
	// Write CSV header
	fmt.Fprintf(outputFile, "Z_real,Z_imag,Spectrum_Number,Frequency_Hz\n")
	log.Printf("Created output file: %s", outputFilePath)
	
	// Start data generation loop (1 second intervals like Python)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	measurementCounter := 1
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Direct EIS generator stopping due to context cancellation")
			return
			
		case <-signalChan:
			log.Println("Shutdown signal received, stopping...")
			cancel()
			return
			
		case <-ticker.C:
			// Generate EIS spectrum (exactly like Python code)
			impedanceData := eisGenerator.GenerateEISSpectrum(params)
			
			currentSpectrum := eisGenerator.GetCurrentSpectrum() - 1 // -1 because it increments after generation
			log.Printf("Generated EIS spectrum #%d at %s", 
				currentSpectrum, 
				time.Now().Format("15:04:05"))
			
			// Convert to EIS measurement format
			eisMeasurement := make(signal.EISMeasurement, len(impedanceData.Impedance))
			for i, z := range impedanceData.Impedance {
				eisMeasurement[i] = signal.ImpedancePoint{
					Frequency: impedanceData.Frequencies[i],
					Real:      real(z),
					Imag:      imag(z),
				}
			}
			
			// Always save to CSV file (in addition to other outputs)
			for i, z := range impedanceData.Impedance {
				fmt.Fprintf(outputFile, "%.12e,%.12e,%d,%.12e\n", 
					real(z), imag(z), currentSpectrum, impedanceData.Frequencies[i])
			}
			outputFile.Sync() // Ensure data is written to disk
			
			// Output based on mode
			switch outputMode {
			case "http":
				// Send via HTTP to goimpcore
				if err := sender.SendImpedanceData(impedanceData); err != nil {
					log.Printf("Error sending impedance data: %v", err)
				}
				
			case "console":
				// Save to JSON file
				printEISMeasurement(eisMeasurement, "json")
				
			case "csv":
				// Save to CSV file  
				printEISMeasurement(eisMeasurement, "csv")
			}
			
			measurementCounter++
			
			// Stop after 100 spectra
			if currentSpectrum >= 99 {
				log.Println("Generated all 100 spectra, stopping...")
				cancel()
				return
			}
		}
	}
}
