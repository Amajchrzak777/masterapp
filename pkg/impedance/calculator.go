package impedance

import (
	"fmt"
	"math/cmplx"

	"github.com/adam/masterapp/pkg/config"
	"github.com/adam/masterapp/pkg/fft"
	"github.com/adam/masterapp/pkg/signal"
)

// DefaultCalculator implements impedance calculations for EIS measurements
type DefaultCalculator struct {
	fftProcessor fft.Processor
	validator    signal.Validator
}

// NewCalculator creates a new impedance calculator
func NewCalculator() Calculator {
	return &DefaultCalculator{
		fftProcessor: fft.NewProcessor(),
		validator:    signal.NewValidator(),
	}
}

// ValidateSignals validates that voltage and current signals are compatible
func (ic *DefaultCalculator) ValidateSignals(voltageSignal, currentSignal signal.Signal) error {
	if err := ic.validator.ValidateSignal(voltageSignal); err != nil {
		return config.NewValidationError("VoltageSignal", err.Error())
	}

	if err := ic.validator.ValidateSignal(currentSignal); err != nil {
		return config.NewValidationError("CurrentSignal", err.Error())
	}

	return signal.ValidateSignalsMatch(voltageSignal, currentSignal)
}

// CalculateImpedance computes complex impedance Z(f) = U(f)/I(f) from voltage and current signals
func (ic *DefaultCalculator) CalculateImpedance(voltageSignal, currentSignal signal.Signal) (signal.ImpedanceData, error) {
	if err := ic.ValidateSignals(voltageSignal, currentSignal); err != nil {
		return signal.ImpedanceData{}, config.NewProcessingError("signal validation", err)
	}

	voltageFFT, err := ic.fftProcessor.ProcessSignal(voltageSignal)
	if err != nil {
		return signal.ImpedanceData{}, config.NewProcessingError("voltage FFT processing", err)
	}
	
	currentFFT, err := ic.fftProcessor.ProcessSignal(currentSignal)
	if err != nil {
		return signal.ImpedanceData{}, config.NewProcessingError("current FFT processing", err)
	}

	voltageFFT, err = ic.fftProcessor.GetPositiveFrequencies(voltageFFT)
	if err != nil {
		return signal.ImpedanceData{}, config.NewProcessingError("voltage positive frequencies", err)
	}
	
	currentFFT, err = ic.fftProcessor.GetPositiveFrequencies(currentFFT)
	if err != nil {
		return signal.ImpedanceData{}, config.NewProcessingError("current positive frequencies", err)
	}

	if len(voltageFFT.Values) != len(currentFFT.Values) {
		return signal.ImpedanceData{}, config.NewProcessingError("impedance calculation", config.ErrMismatchedSignalLength)
	}

	impedance := make([]complex128, len(voltageFFT.Values))
	for i := 0; i < len(voltageFFT.Values); i++ {
		currentMagnitude := cmplx.Abs(currentFFT.Values[i])
		if currentMagnitude < 1e-10 {
			impedance[i] = complex(0, 0)
		} else {
			impedance[i] = voltageFFT.Values[i] / currentFFT.Values[i]
			
			if cmplx.IsNaN(impedance[i]) || cmplx.IsInf(impedance[i]) {
				return signal.ImpedanceData{}, config.NewProcessingError("impedance calculation", 
					config.NewValidationError("Impedance", fmt.Sprintf("invalid impedance value at frequency index %d", i)))
			}
		}
	}

	impedanceData := signal.ImpedanceData{
		Timestamp:   voltageSignal.Timestamp,
		Impedance:   impedance,
		Frequencies: voltageFFT.Frequencies,
	}

	magnitude, phase := impedanceData.CalculateMagnitudePhase()
	impedanceData.Magnitude = magnitude
	impedanceData.Phase = phase

	if err := ic.validator.ValidateImpedanceData(impedanceData); err != nil {
		return signal.ImpedanceData{}, config.NewProcessingError("impedance data validation", err)
	}

	return impedanceData, nil
}

// ProcessEISMeasurement performs a complete EIS measurement including FFT and impedance calculation
func (ic *DefaultCalculator) ProcessEISMeasurement(voltageSignal, currentSignal signal.Signal) (signal.EISMeasurement, error) {
	if err := ic.ValidateSignals(voltageSignal, currentSignal); err != nil {
		return signal.EISMeasurement{}, config.NewProcessingError("signal validation", err)
	}


	impedanceData, err := ic.CalculateImpedance(voltageSignal, currentSignal)
	if err != nil {
		return signal.EISMeasurement{}, config.NewProcessingError("impedance calculation", err)
	}

	// Convert to EISMeasurement format
	measurement := make(signal.EISMeasurement, len(impedanceData.Impedance))
	for i, z := range impedanceData.Impedance {
		measurement[i] = signal.ImpedancePoint{
			Frequency: impedanceData.Frequencies[i],
			Real:      real(z),
			Imag:      imag(z),
		}
	}

	return measurement, nil
}