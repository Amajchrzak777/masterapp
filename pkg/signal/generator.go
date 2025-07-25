package signal

import (
	"math"
	"math/rand"
	"time"

	"github.com/adam/masterapp/pkg/config"
)

// DefaultGenerator implements signal generation for testing and simulation
type DefaultGenerator struct{}

// NewGenerator creates a new signal generator
func NewGenerator() Generator {
	return &DefaultGenerator{}
}

// GenerateVoltageSignal generates a realistic voltage signal with sine wave and noise
func (sg *DefaultGenerator) GenerateVoltageSignal(sampleRate float64, samplesPerSecond int) (Signal, error) {
	if sampleRate <= 0 {
		return Signal{}, config.ErrInvalidSampleRate
	}
	
	if samplesPerSecond <= 0 {
		return Signal{}, config.NewValidationError("SamplesPerSecond", "samples per second must be greater than 0")
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

// GenerateCurrentSignal generates a corresponding current signal with phase shift and noise
func (sg *DefaultGenerator) GenerateCurrentSignal(sampleRate float64, samplesPerSecond int) (Signal, error) {
	if sampleRate <= 0 {
		return Signal{}, config.ErrInvalidSampleRate
	}
	
	if samplesPerSecond <= 0 {
		return Signal{}, config.NewValidationError("SamplesPerSecond", "samples per second must be greater than 0")
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