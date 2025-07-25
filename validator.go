package main

import (
	"fmt"
	"math"
	"math/cmplx"
)

type DefaultSignalValidator struct{}

func NewSignalValidator() SignalValidator {
	return &DefaultSignalValidator{}
}

func (v *DefaultSignalValidator) ValidateSignal(signal Signal) error {
	if len(signal.Values) == 0 {
		return NewValidationError("Values", "signal values cannot be empty")
	}

	if signal.SampleRate <= 0 {
		return NewValidationError("SampleRate", "sample rate must be greater than 0")
	}

	if signal.Timestamp.IsZero() {
		return NewValidationError("Timestamp", "timestamp cannot be zero")
	}

	for i, value := range signal.Values {
		if math.IsNaN(value) {
			return NewValidationError("Values", fmt.Sprintf("NaN value found at index %d", i))
		}
		if math.IsInf(value, 0) {
			return NewValidationError("Values", fmt.Sprintf("infinite value found at index %d", i))
		}
	}

	return nil
}

func (v *DefaultSignalValidator) ValidateComplexSignal(signal ComplexSignal) error {
	if len(signal.Values) == 0 {
		return NewValidationError("Values", "complex signal values cannot be empty")
	}

	if len(signal.Frequencies) == 0 {
		return NewValidationError("Frequencies", "frequencies cannot be empty")
	}

	if len(signal.Values) != len(signal.Frequencies) {
		return NewValidationError("Length", "values and frequencies must have the same length")
	}

	if signal.Timestamp.IsZero() {
		return NewValidationError("Timestamp", "timestamp cannot be zero")
	}

	for i, value := range signal.Values {
		if cmplx.IsNaN(value) {
			return NewValidationError("Values", fmt.Sprintf("NaN complex value found at index %d", i))
		}
		if cmplx.IsInf(value) {
			return NewValidationError("Values", fmt.Sprintf("infinite complex value found at index %d", i))
		}
	}

	for i, freq := range signal.Frequencies {
		if math.IsNaN(freq) {
			return NewValidationError("Frequencies", fmt.Sprintf("NaN frequency found at index %d", i))
		}
		if math.IsInf(freq, 0) {
			return NewValidationError("Frequencies", fmt.Sprintf("infinite frequency found at index %d", i))
		}
		// Note: Negative frequencies are allowed in full FFT spectrum
	}

	return nil
}

func (v *DefaultSignalValidator) ValidatePositiveFrequencySignal(signal ComplexSignal) error {
	if err := v.ValidateComplexSignal(signal); err != nil {
		return err
	}

	// Additional check for positive frequencies only
	for i, freq := range signal.Frequencies {
		if freq < 0 {
			return NewValidationError("Frequencies", fmt.Sprintf("negative frequency found at index %d", i))
		}
	}

	return nil
}

func (v *DefaultSignalValidator) ValidateImpedanceData(data ImpedanceData) error {
	if len(data.Impedance) == 0 {
		return NewValidationError("Impedance", "impedance values cannot be empty")
	}

	if len(data.Frequencies) == 0 {
		return NewValidationError("Frequencies", "frequencies cannot be empty")
	}

	if len(data.Impedance) != len(data.Frequencies) {
		return NewValidationError("Length", "impedance and frequencies must have the same length")
	}

	if len(data.Magnitude) > 0 && len(data.Magnitude) != len(data.Impedance) {
		return NewValidationError("Magnitude", "magnitude length must match impedance length")
	}

	if len(data.Phase) > 0 && len(data.Phase) != len(data.Impedance) {
		return NewValidationError("Phase", "phase length must match impedance length")
	}

	if data.Timestamp.IsZero() {
		return NewValidationError("Timestamp", "timestamp cannot be zero")
	}

	for i, imp := range data.Impedance {
		if cmplx.IsNaN(imp) {
			return NewValidationError("Impedance", fmt.Sprintf("NaN impedance value found at index %d", i))
		}
		if cmplx.IsInf(imp) {
			return NewValidationError("Impedance", fmt.Sprintf("infinite impedance value found at index %d", i))
		}
	}

	return nil
}

func ValidateSignalsMatch(voltageSignal, currentSignal Signal) error {
	if len(voltageSignal.Values) != len(currentSignal.Values) {
		return ErrMismatchedSignalLength
	}

	if voltageSignal.SampleRate != currentSignal.SampleRate {
		return ErrMismatchedSampleRate
	}

	timeDiff := voltageSignal.Timestamp.Sub(currentSignal.Timestamp)
	if math.Abs(timeDiff.Seconds()) > 0.1 { // Allow 100ms tolerance
		return NewValidationError("Timestamp", "voltage and current signals have significantly different timestamps")
	}

	return nil
}

func ValidateConfiguration(sampleRate float64, samplesPerSecond int, targetURL string) error {
	if sampleRate <= 0 {
		return NewValidationError("SampleRate", "sample rate must be greater than 0")
	}

	if samplesPerSecond <= 0 {
		return NewValidationError("SamplesPerSecond", "samples per second must be greater than 0")
	}

	if targetURL == "" {
		return NewValidationError("TargetURL", "target URL cannot be empty")
	}

	return nil
}