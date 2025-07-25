package receiver

import (
	"context"
	"log"
	"time"

	"github.com/adam/masterapp/pkg/config"
	"github.com/adam/masterapp/pkg/signal"
)

// DefaultReceiver implements real-time signal reception with simulation
type DefaultReceiver struct {
	voltageChannel   chan signal.Signal
	currentChannel   chan signal.Signal
	sampleRate       float64
	samplesPerSecond int
	validator        signal.Validator
	generator        signal.Generator
	running          bool
}

// NewReceiver creates a new data receiver
func NewReceiver(sampleRate float64, samplesPerSecond int) DataReceiver {
	return &DefaultReceiver{
		voltageChannel:   make(chan signal.Signal, 10),
		currentChannel:   make(chan signal.Signal, 10),
		sampleRate:       sampleRate,
		samplesPerSecond: samplesPerSecond,
		validator:        signal.NewValidator(),
		generator:        signal.NewGenerator(),
		running:          false,
	}
}

// StartReceiving begins real-time data reception at 1-second intervals
func (dr *DefaultReceiver) StartReceiving(ctx context.Context) error {
	// Validate configuration
	cfg := &config.Config{
		SampleRate:       dr.sampleRate,
		SamplesPerSecond: dr.samplesPerSecond,
		TargetURL:        "dummy", // Not used in receiver validation
	}
	if err := cfg.Validate(); err != nil {
		return config.NewProcessingError("configuration validation", err)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	dr.running = true
	log.Println("Starting real-time data reception (1-second intervals)")

	for dr.running {
		select {
		case <-ctx.Done():
			dr.running = false
			return ctx.Err()
		case <-ticker.C:
			voltageSignal, err := dr.generator.GenerateVoltageSignal(dr.sampleRate, dr.samplesPerSecond)
			if err != nil {
				log.Printf("Error generating voltage signal: %v", err)
				continue
			}

			currentSignal, err := dr.generator.GenerateCurrentSignal(dr.sampleRate, dr.samplesPerSecond)
			if err != nil {
				log.Printf("Error generating current signal: %v", err)
				continue
			}

			if err := dr.validator.ValidateSignal(voltageSignal); err != nil {
				log.Printf("Invalid voltage signal: %v", err)
				continue
			}

			if err := dr.validator.ValidateSignal(currentSignal); err != nil {
				log.Printf("Invalid current signal: %v", err)
				continue
			}

			select {
			case dr.voltageChannel <- voltageSignal:
			default:
				log.Println("Warning: Voltage channel buffer full, dropping sample")
			}

			select {
			case dr.currentChannel <- currentSignal:
			default:
				log.Println("Warning: Current channel buffer full, dropping sample")
			}

			log.Printf("Received data at %v", time.Now().Format("15:04:05"))
		}
	}

	return nil
}

// GetVoltageChannel returns the channel for voltage signals
func (dr *DefaultReceiver) GetVoltageChannel() <-chan signal.Signal {
	return dr.voltageChannel
}

// GetCurrentChannel returns the channel for current signals
func (dr *DefaultReceiver) GetCurrentChannel() <-chan signal.Signal {
	return dr.currentChannel
}

// Stop gracefully stops the receiver and closes channels
func (dr *DefaultReceiver) Stop() error {
	dr.running = false
	close(dr.voltageChannel)
	close(dr.currentChannel)
	return nil
}