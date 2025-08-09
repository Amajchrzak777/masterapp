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
		
		// Generate multi-frequency voltage excitation based on impedance_data.csv pattern
		// This creates a broadband signal that will result in EIS-like frequency response
		signal := 0.0
		
		// Add multiple frequency components with decreasing amplitude (realistic EIS excitation)
		frequencies := []float64{1, 5, 10, 25, 50, 100, 250, 500}
		amplitudes := []float64{0.2, 0.15, 0.12, 0.1, 0.08, 0.06, 0.04, 0.02}
		
		for j, freq := range frequencies {
			if j < len(amplitudes) {
				signal += amplitudes[j] * math.Sin(2*math.Pi*freq*t)
			}
		}
		
		// Add DC component and small measurement noise
		values[i] = 1.0 + signal + 0.01*(rand.Float64()-0.5)
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
		
		// Generate current response simulating R(RC) electrochemical behavior
		// Current response has frequency-dependent amplitude and phase based on impedance_data.csv
		signal := 0.0
		
		// Same frequencies as voltage but with impedance-modified amplitude and phase
		frequencies := []float64{1, 5, 10, 25, 50, 100, 250, 500}
		voltageAmps := []float64{0.2, 0.15, 0.12, 0.1, 0.08, 0.06, 0.04, 0.02}
		
		for j, freq := range frequencies {
			if j < len(voltageAmps) {
				// Simulate R(RC) circuit response: |I| = |U|/|Z| and phase shift
				// Higher frequencies: lower impedance (~10-11 Ω), less phase shift
				// Lower frequencies: higher impedance (~30 Ω), more phase shift
				impedanceMagnitude := 10.0 + 20.0/(1.0 + freq/10.0) // Simplified R(RC) model
				phaseShift := math.Atan(freq/50.0) * 0.5 // Capacitive phase shift
				
				currentAmplitude := voltageAmps[j] / impedanceMagnitude
				signal += currentAmplitude * math.Sin(2*math.Pi*freq*t - phaseShift)
			}
		}
		
		// Add DC component and measurement noise
		values[i] = 0.05 + signal + 0.005*(rand.Float64()-0.5)
	}

	return Signal{
		Timestamp:  now,
		Values:     values,
		SampleRate: sampleRate,
	}, nil
}