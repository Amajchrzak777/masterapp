package signal

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/adam/masterapp/pkg/config"
)

// DefaultValidator implements signal validation logic
type DefaultValidator struct{}

// NewValidator creates a new signal validator
func NewValidator() Validator {
	return &DefaultValidator{}
}

// ValidateSignal validates a time-domain signal
func (v *DefaultValidator) ValidateSignal(signal Signal) error {
	if len(signal.Values) == 0 {
		return config.NewValidationError("Values", "signal values cannot be empty")
	}

	if signal.SampleRate <= 0 {
		return config.NewValidationError("SampleRate", "sample rate must be greater than 0")
	}

	if signal.Timestamp.IsZero() {
		return config.NewValidationError("Timestamp", "timestamp cannot be zero")
	}

	for i, value := range signal.Values {
		if math.IsNaN(value) {
			return config.NewValidationError("Values", fmt.Sprintf("NaN value found at index %d", i))
		}
		if math.IsInf(value, 0) {
			return config.NewValidationError("Values", fmt.Sprintf("infinite value found at index %d", i))
		}
	}

	return nil
}

// ValidateComplexSignal validates a frequency-domain signal
func (v *DefaultValidator) ValidateComplexSignal(signal ComplexSignal) error {
	if len(signal.Values) == 0 {
		return config.NewValidationError("Values", "complex signal values cannot be empty")
	}

	if len(signal.Frequencies) == 0 {
		return config.NewValidationError("Frequencies", "frequencies cannot be empty")
	}

	if len(signal.Values) != len(signal.Frequencies) {
		return config.NewValidationError("Length", "values and frequencies must have the same length")
	}

	if signal.Timestamp.IsZero() {
		return config.NewValidationError("Timestamp", "timestamp cannot be zero")
	}

	for i, value := range signal.Values {
		if cmplx.IsNaN(value) {
			return config.NewValidationError("Values", fmt.Sprintf("NaN complex value found at index %d", i))
		}
		if cmplx.IsInf(value) {
			return config.NewValidationError("Values", fmt.Sprintf("infinite complex value found at index %d", i))
		}
	}

	for i, freq := range signal.Frequencies {
		if math.IsNaN(freq) {
			return config.NewValidationError("Frequencies", fmt.Sprintf("NaN frequency found at index %d", i))
		}
		if math.IsInf(freq, 0) {
			return config.NewValidationError("Frequencies", fmt.Sprintf("infinite frequency found at index %d", i))
		}
		// Note: Negative frequencies are allowed in full FFT spectrum
	}

	return nil
}

// ValidatePositiveFrequencySignal validates a signal that should only contain positive frequencies
func (v *DefaultValidator) ValidatePositiveFrequencySignal(signal ComplexSignal) error {
	if err := v.ValidateComplexSignal(signal); err != nil {
		return err
	}

	// Additional check for positive frequencies only
	for i, freq := range signal.Frequencies {
		if freq < 0 {
			return config.NewValidationError("Frequencies", fmt.Sprintf("negative frequency found at index %d", i))
		}
	}

	return nil
}

// ValidateImpedanceData validates impedance measurement data
func (v *DefaultValidator) ValidateImpedanceData(data ImpedanceData) error {
	if len(data.Impedance) == 0 {
		return config.NewValidationError("Impedance", "impedance values cannot be empty")
	}

	if len(data.Frequencies) == 0 {
		return config.NewValidationError("Frequencies", "frequencies cannot be empty")
	}

	if len(data.Impedance) != len(data.Frequencies) {
		return config.NewValidationError("Length", "impedance and frequencies must have the same length")
	}

	if len(data.Magnitude) > 0 && len(data.Magnitude) != len(data.Impedance) {
		return config.NewValidationError("Magnitude", "magnitude length must match impedance length")
	}

	if len(data.Phase) > 0 && len(data.Phase) != len(data.Impedance) {
		return config.NewValidationError("Phase", "phase length must match impedance length")
	}

	if data.Timestamp.IsZero() {
		return config.NewValidationError("Timestamp", "timestamp cannot be zero")
	}

	for i, imp := range data.Impedance {
		if cmplx.IsNaN(imp) {
			return config.NewValidationError("Impedance", fmt.Sprintf("NaN impedance value found at index %d", i))
		}
		if cmplx.IsInf(imp) {
			return config.NewValidationError("Impedance", fmt.Sprintf("infinite impedance value found at index %d", i))
		}
	}

	return nil
}

// ValidateSignalsMatch validates that voltage and current signals are compatible
func ValidateSignalsMatch(voltageSignal, currentSignal Signal) error {
	if len(voltageSignal.Values) != len(currentSignal.Values) {
		return config.ErrMismatchedSignalLength
	}

	if voltageSignal.SampleRate != currentSignal.SampleRate {
		return config.ErrMismatchedSampleRate
	}

	timeDiff := voltageSignal.Timestamp.Sub(currentSignal.Timestamp)
	if math.Abs(timeDiff.Seconds()) > 0.1 { // Allow 100ms tolerance
		return config.NewValidationError("Timestamp", "voltage and current signals have significantly different timestamps")
	}

	return nil
}