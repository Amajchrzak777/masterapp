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
		circuitType   = flag.String("circuit", "simple", "Circuit complexity: 'simple' (R(CR)), 'medium' (R(Q(R(QR)))), 'complex' (multi-stage)")
		spectraCount  = flag.Int("spectra", 5, "Number of spectra to generate for direct EIS mode")
		impedanceCSV  = flag.String("impedance-csv", "", "Path to impedance CSV file (Frequency_Hz,Z_real,Z_imag,Spectrum_Number)")
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

	// Check if using impedance CSV file input
	if *impedanceCSV != "" {
		log.Printf("Using impedance CSV file input: %s", *impedanceCSV)
		runImpedanceCSVMode(cfg, *outputMode, *impedanceCSV)
		return
	}

	// Check if using direct EIS generation mode
	if *useDirectEIS {
		log.Println("Using direct EIS generation (Python impedance_data.csv approach)")
		runDirectEISMode(cfg, *outputMode, *circuitType, *spectraCount)
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

// getCircuitParameters returns circuit parameters based on complexity level
func getCircuitParameters(circuitType string) eisgen.CircuitParameters {
	switch circuitType {
	case "simple":
		// Simple R(CR) circuit - 3 parameters  
		return eisgen.CircuitParameters{
			Rs:         10.0,   // Solution resistance
			RctInitial: 20.0,   // Initial charge transfer resistance  
			RctGrowth:  8.0,    // Growth per spectrum
			Q:          1e-5,   // CPE coefficient
			N:          0.85,   // CPE exponent
		}
	case "medium":
		// Medium R(Q(R(QR))) circuit - 7 parameters
		// More challenging optimization with different parameter values
		return eisgen.CircuitParameters{
			Rs:         15.0,   // Higher solution resistance
			RctInitial: 50.0,   // Higher charge transfer resistance
			RctGrowth:  12.0,   // Faster degradation  
			Q:          5e-6,   // Different CPE coefficient
			N:          0.75,   // Different CPE exponent (more capacitive)
		}
	case "complex":
		// Complex multi-stage circuit - 12+ parameters
		// Very challenging optimization  
		return eisgen.CircuitParameters{
			Rs:         8.0,    // Lower solution resistance
			RctInitial: 80.0,   // High charge transfer resistance
			RctGrowth:  20.0,   // Aggressive degradation
			Q:          2e-6,   // Low CPE coefficient  
			N:          0.65,   // Low CPE exponent (diffusion-like)
		}
	default:
		// Default to simple
		return eisgen.CircuitParameters{
			Rs:         10.0,
			RctInitial: 20.0,
			RctGrowth:  8.0,
			Q:          1e-5,
			N:          0.85,
		}
	}
}

// runDirectEISMode runs the direct EIS generation mode (like Python code)
func runDirectEISMode(cfg *config.Config, outputMode, circuitType string, spectraCount int) {
	log.Println("Starting Direct EIS generation mode")
	log.Printf("Circuit complexity: %s", circuitType)
	log.Printf("Generating %d spectra", spectraCount)
	
	// Create EIS generator with parameters based on circuit complexity
	eisGenerator := eisgen.NewEISGenerator()
	params := getCircuitParameters(circuitType)
	
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
	
	// Create output file with circuit type in name
	outputFilePath := fmt.Sprintf("generated_eis_data_%s.csv", circuitType)
	if _, err := os.Stat("/root/data"); err == nil {
		// Running in Docker container
		outputFilePath = fmt.Sprintf("/root/data/generated_eis_data_%s.csv", circuitType)
	}
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()
	
	// Write CSV header
	fmt.Fprintf(outputFile, "Z_real,Z_imag,Spectrum_Number,Frequency_Hz\n")
	log.Printf("Created output file: %s", outputFilePath)
	
	// Batch processing: generate 10 spectra per batch every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	measurementCounter := 1
	batchSize := 10
	
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
			// Generate batch of spectra
			batch := make([]signal.ImpedanceDataWithIteration, 0, batchSize)
			
			for i := 0; i < batchSize; i++ {
				currentSpectrum := eisGenerator.GetCurrentSpectrum()
				if currentSpectrum >= spectraCount {
					break // Stop at specified number of spectra
				}
				
				// Generate EIS spectrum
				impedanceData := eisGenerator.GenerateEISSpectrum(params)
				
				// Create batch item with iteration number for proper ordering
				batchItem := signal.ImpedanceDataWithIteration{
					ImpedanceData: impedanceData,
					Iteration:     currentSpectrum,
				}
				batch = append(batch, batchItem)
				
				// Always save to CSV file
				for j, z := range impedanceData.Impedance {
					fmt.Fprintf(outputFile, "%.12e,%.12e,%d,%.12e\n", 
						real(z), imag(z), currentSpectrum, impedanceData.Frequencies[j])
				}
			}
			
			if len(batch) == 0 {
				log.Printf("Generated all %d spectra, stopping...", spectraCount)
				cancel()
				return
			}
			
			outputFile.Sync() // Ensure data is written to disk
			
			log.Printf("Generated batch of %d spectra (iterations %d-%d) at %s", 
				len(batch), 
				batch[0].Iteration, 
				batch[len(batch)-1].Iteration,
				time.Now().Format("15:04:05"))
			
			// Output based on mode
			switch outputMode {
			case "http":
				// Send batch via HTTP to goimpcore
				if err := sender.SendBatchImpedanceData(batch); err != nil {
					log.Printf("Error sending batch impedance data: %v", err)
				}
				
			case "console":
				// Save individual measurements to JSON files
				for _, item := range batch {
					eisMeasurement := make(signal.EISMeasurement, len(item.ImpedanceData.Impedance))
					for j, z := range item.ImpedanceData.Impedance {
						eisMeasurement[j] = signal.ImpedancePoint{
							Frequency: item.ImpedanceData.Frequencies[j],
							Real:      real(z),
							Imag:      imag(z),
						}
					}
					printEISMeasurement(eisMeasurement, "json")
				}
				
			case "csv":
				// Already saved above
			}
			
			measurementCounter += len(batch)
			
			// Check if we've generated all spectra
			if eisGenerator.GetCurrentSpectrum() >= 100 {
				log.Println("Generated all 100 spectra, stopping...")
				cancel()
				return
			}
		}
	}
}

// runImpedanceCSVMode reads impedance data from CSV file and sends it to target
func runImpedanceCSVMode(cfg *config.Config, outputMode, csvPath string) {
	log.Println("Starting Impedance CSV mode")
	log.Printf("Reading impedance data from: %s", csvPath)
	
	// Create data loader
	dataLoader := signal.NewDataLoader()
	csvLoader, ok := dataLoader.(*signal.CSVDataLoader)
	if !ok {
		log.Fatalf("Failed to create CSV data loader")
	}
	
	// Load impedance data from CSV
	impedanceData, err := csvLoader.LoadImpedanceFromCSV(csvPath)
	if err != nil {
		log.Fatalf("Failed to load impedance data: %v", err)
	}
	
	log.Printf("Loaded %d spectra from CSV file", len(impedanceData))
	
	// Create network sender
	sender := network.NewSender(cfg.TargetURL)
	
	// Wait a bit for goimpcore to be ready (in Docker environment)
	log.Println("Waiting 5 seconds for target server to be ready...")
	time.Sleep(5 * time.Second)
	
	// Output based on mode
	switch outputMode {
	case "http":
		// Send all spectra as a single batch to goimpcore
		log.Printf("Sending %d spectra as batch to: %s", len(impedanceData), cfg.TargetURL)
		
		if err := sender.SendBatchImpedanceData(impedanceData); err != nil {
			log.Printf("Error sending batch impedance data: %v", err)
		} else {
			log.Printf("Successfully sent batch of %d spectra", len(impedanceData))
		}
		
	case "console":
		// Save individual measurements to JSON files
		log.Printf("Saving %d spectra to JSON files", len(impedanceData))
		
		for _, item := range impedanceData {
			eisMeasurement := make(signal.EISMeasurement, len(item.ImpedanceData.Impedance))
			for j, z := range item.ImpedanceData.Impedance {
				eisMeasurement[j] = signal.ImpedancePoint{
					Frequency: item.ImpedanceData.Frequencies[j],
					Real:      real(z),
					Imag:      imag(z),
				}
			}
			printEISMeasurement(eisMeasurement, "json")
		}
		
	case "csv":
		// Save each spectrum as separate CSV file
		log.Printf("Saving %d spectra to CSV files", len(impedanceData))
		
		for _, item := range impedanceData {
			eisMeasurement := make(signal.EISMeasurement, len(item.ImpedanceData.Impedance))
			for j, z := range item.ImpedanceData.Impedance {
				eisMeasurement[j] = signal.ImpedancePoint{
					Frequency: item.ImpedanceData.Frequencies[j],
					Real:      real(z),
					Imag:      imag(z),
				}
			}
			printEISMeasurement(eisMeasurement, "csv")
		}
	}
	
	log.Println("Impedance CSV processing completed")
}
