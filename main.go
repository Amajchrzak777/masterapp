package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	var (
		targetURL    = flag.String("target", "http://localhost:8080/eis-data", "Target URL for sending EIS data")
		sampleRate   = flag.Float64("rate", 1000.0, "Sample rate in Hz")
		samplesPerSec = flag.Int("samples", 1000, "Number of samples per second")
	)
	flag.Parse()

	// Validate configuration
	if err := ValidateConfiguration(*sampleRate, *samplesPerSec, *targetURL); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Println("Starting Dynamic Electrochemical Impedance Spectroscopy (DEIS) processor")
	log.Printf("Target URL: %s", *targetURL)
	log.Printf("Sample rate: %.1f Hz", *sampleRate)
	log.Printf("Samples per second: %d", *samplesPerSec)

	// Initialize components using dependency injection
	receiver := NewDataReceiver(*sampleRate, *samplesPerSec)
	calculator := NewImpedanceCalculator()
	sender := NewDataSender(*targetURL)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(2)

	// Start data receiver
	go func() {
		defer wg.Done()
		if err := receiver.StartReceiving(ctx); err != nil && err != context.Canceled {
			log.Printf("Data receiver error: %v", err)
		}
	}()

	// Start signal processor
	go func() {
		defer wg.Done()
		processSignals(ctx, receiver, calculator, sender)
	}()

	// Wait for shutdown signal
	<-signalChan
	log.Println("Shutdown signal received, stopping...")
	
	// Cancel context to stop all goroutines
	cancel()
	
	// Stop receiver
	if err := receiver.Stop(); err != nil {
		log.Printf("Error stopping receiver: %v", err)
	}

	wg.Wait()
	log.Println("DEIS processor stopped")
}

func processSignals(ctx context.Context, receiver DataReceiver, calculator ImpedanceCalculator, sender DataSender) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Signal processor stopping due to context cancellation")
			return
		case voltageSignal := <-receiver.GetVoltageChannel():
			select {
			case currentSignal := <-receiver.GetCurrentChannel():
				measurement, err := calculator.ProcessEISMeasurement(voltageSignal, currentSignal)
				if err != nil {
					log.Printf("Error processing EIS measurement: %v", err)
					continue
				}

				if err := sender.SendEISMeasurement(measurement); err != nil {
					log.Printf("Error sending EIS measurement: %v", err)
					
					// Check if sender is unhealthy and log warning
					if !sender.IsHealthy() {
						log.Printf("Warning: Data sender is unhealthy")
					}
				}
			default:
				log.Println("Warning: No current signal available for voltage signal")
			}
		}
	}
}
