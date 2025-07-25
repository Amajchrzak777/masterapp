package config

import (
	"fmt"
	"math"
)

// Config holds the application configuration
type Config struct {
	TargetURL        string  `json:"target_url"`
	SampleRate       float64 `json:"sample_rate"`
	SamplesPerSecond int     `json:"samples_per_second"`
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		TargetURL:        "http://localhost:8080/eis-data",
		SampleRate:       1000.0,
		SamplesPerSecond: 1000,
	}
}

// Validate validates the configuration parameters
func (c *Config) Validate() error {
	if c.SampleRate <= 0 {
		return NewValidationError("SampleRate", "sample rate must be greater than 0")
	}

	if c.SamplesPerSecond <= 0 {
		return NewValidationError("SamplesPerSecond", "samples per second must be greater than 0")
	}

	if c.TargetURL == "" {
		return NewValidationError("TargetURL", "target URL cannot be empty")
	}

	// Check for reasonable limits
	if c.SampleRate > 1000000 { // 1MHz
		return NewValidationError("SampleRate", "sample rate exceeds reasonable limit (1MHz)")
	}

	if c.SamplesPerSecond > 100000 { // 100k samples
		return NewValidationError("SamplesPerSecond", "samples per second exceeds reasonable limit (100k)")
	}

	return nil
}

// ValidateSignalsMatch validates that two signals are compatible for processing
func ValidateSignalsMatch(voltageLength, currentLength int, voltageSampleRate, currentSampleRate float64) error {
	if voltageLength != currentLength {
		return ErrMismatchedSignalLength
	}

	if voltageSampleRate != currentSampleRate {
		return ErrMismatchedSampleRate
	}

	// Check for reasonable time difference tolerance (this would need actual timestamps in real implementation)
	return nil
}

// ValidateSignalData validates basic signal data properties
func ValidateSignalData(values []float64, sampleRate float64) error {
	if len(values) == 0 {
		return ErrInvalidSignalLength
	}

	if sampleRate <= 0 {
		return ErrInvalidSampleRate
	}

	// Check for invalid values
	for i, value := range values {
		if math.IsNaN(value) {
			return NewValidationError("Values", fmt.Sprintf("NaN value found at index %d", i))
		}
		if math.IsInf(value, 0) {
			return NewValidationError("Values", fmt.Sprintf("infinite value found at index %d", i))
		}
	}

	return nil
}

// ValidateFrequencies validates frequency array data
func ValidateFrequencies(frequencies []float64, allowNegative bool) error {
	if len(frequencies) == 0 {
		return ErrEmptyFrequencies
	}

	for i, freq := range frequencies {
		if math.IsNaN(freq) {
			return NewValidationError("Frequencies", fmt.Sprintf("NaN frequency found at index %d", i))
		}
		if math.IsInf(freq, 0) {
			return NewValidationError("Frequencies", fmt.Sprintf("infinite frequency found at index %d", i))
		}
		if !allowNegative && freq < 0 {
			return NewValidationError("Frequencies", fmt.Sprintf("negative frequency found at index %d", i))
		}
	}

	return nil
}