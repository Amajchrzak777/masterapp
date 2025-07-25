package main

import (
	"context"
	"log"
	"math"
	"math/rand"
	"time"
)

type DefaultDataReceiver struct {
	voltageChannel   chan Signal
	currentChannel   chan Signal
	sampleRate       float64
	samplesPerSecond int
	validator        SignalValidator
	generator        SignalGenerator
	running          bool
}

func NewDataReceiver(sampleRate float64, samplesPerSecond int) DataReceiver {
	return &DefaultDataReceiver{
		voltageChannel:   make(chan Signal, 10),
		currentChannel:   make(chan Signal, 10),
		sampleRate:       sampleRate,
		samplesPerSecond: samplesPerSecond,
		validator:        NewSignalValidator(),
		generator:        NewSignalGenerator(),
		running:          false,
	}
}

func (dr *DefaultDataReceiver) StartReceiving(ctx context.Context) error {
	if err := ValidateConfiguration(dr.sampleRate, dr.samplesPerSecond, "dummy"); err != nil {
		return NewProcessingError("configuration validation", err)
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

func (dr *DefaultDataReceiver) GetVoltageChannel() <-chan Signal {
	return dr.voltageChannel
}

func (dr *DefaultDataReceiver) GetCurrentChannel() <-chan Signal {
	return dr.currentChannel
}

func (dr *DefaultDataReceiver) Stop() error {
	dr.running = false
	close(dr.voltageChannel)
	close(dr.currentChannel)
	return nil
}

type DefaultSignalGenerator struct{}

func NewSignalGenerator() SignalGenerator {
	return &DefaultSignalGenerator{}
}

func (sg *DefaultSignalGenerator) GenerateVoltageSignal(sampleRate float64, samplesPerSecond int) (Signal, error) {
	if sampleRate <= 0 {
		return Signal{}, ErrInvalidSampleRate
	}
	
	if samplesPerSecond <= 0 {
		return Signal{}, NewValidationError("SamplesPerSecond", "samples per second must be greater than 0")
	}

	values := make([]float64, samplesPerSecond)
	now := time.Now()
	
	for i := 0; i < samplesPerSecond; i++ {
		t := float64(i) / sampleRate
		// Generate a more realistic voltage signal with sine wave + noise
		values[i] = 1.0 + 0.5*math.Sin(2*math.Pi*10*t) + 0.1*rand.Float64()
	}

	return Signal{
		Timestamp:  now,
		Values:     values,
		SampleRate: sampleRate,
	}, nil
}

func (sg *DefaultSignalGenerator) GenerateCurrentSignal(sampleRate float64, samplesPerSecond int) (Signal, error) {
	if sampleRate <= 0 {
		return Signal{}, ErrInvalidSampleRate
	}
	
	if samplesPerSecond <= 0 {
		return Signal{}, NewValidationError("SamplesPerSecond", "samples per second must be greater than 0")
	}

	values := make([]float64, samplesPerSecond)
	now := time.Now()
	
	for i := 0; i < samplesPerSecond; i++ {
		t := float64(i) / sampleRate
		// Generate a corresponding current signal with phase shift + noise
		values[i] = 0.1 + 0.05*math.Sin(2*math.Pi*10*t + math.Pi/4) + 0.01*rand.Float64()
	}

	return Signal{
		Timestamp:  now,
		Values:     values,
		SampleRate: sampleRate,
	}, nil
}