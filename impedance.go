package main

import (
	"fmt"
	"math/cmplx"
)

type DefaultImpedanceCalculator struct {
	fftProcessor FFTProcessor
	validator    SignalValidator
}

func NewImpedanceCalculator() ImpedanceCalculator {
	return &DefaultImpedanceCalculator{
		fftProcessor: NewFFTProcessor(),
		validator:    NewSignalValidator(),
	}
}

func (ic *DefaultImpedanceCalculator) ValidateSignals(voltageSignal, currentSignal Signal) error {
	if err := ic.validator.ValidateSignal(voltageSignal); err != nil {
		return NewValidationError("VoltageSignal", err.Error())
	}

	if err := ic.validator.ValidateSignal(currentSignal); err != nil {
		return NewValidationError("CurrentSignal", err.Error())
	}

	return ValidateSignalsMatch(voltageSignal, currentSignal)
}

func (ic *DefaultImpedanceCalculator) CalculateImpedance(voltageSignal, currentSignal Signal) (ImpedanceData, error) {
	if err := ic.ValidateSignals(voltageSignal, currentSignal); err != nil {
		return ImpedanceData{}, NewProcessingError("signal validation", err)
	}

	voltageFFT, err := ic.fftProcessor.ProcessSignal(voltageSignal)
	if err != nil {
		return ImpedanceData{}, NewProcessingError("voltage FFT processing", err)
	}
	
	currentFFT, err := ic.fftProcessor.ProcessSignal(currentSignal)
	if err != nil {
		return ImpedanceData{}, NewProcessingError("current FFT processing", err)
	}

	voltageFFT, err = ic.fftProcessor.GetPositiveFrequencies(voltageFFT)
	if err != nil {
		return ImpedanceData{}, NewProcessingError("voltage positive frequencies", err)
	}
	
	currentFFT, err = ic.fftProcessor.GetPositiveFrequencies(currentFFT)
	if err != nil {
		return ImpedanceData{}, NewProcessingError("current positive frequencies", err)
	}

	if len(voltageFFT.Values) != len(currentFFT.Values) {
		return ImpedanceData{}, NewProcessingError("impedance calculation", ErrMismatchedSignalLength)
	}

	impedance := make([]complex128, len(voltageFFT.Values))
	for i := 0; i < len(voltageFFT.Values); i++ {
		currentMagnitude := cmplx.Abs(currentFFT.Values[i])
		if currentMagnitude < 1e-10 {
			impedance[i] = complex(0, 0)
		} else {
			impedance[i] = voltageFFT.Values[i] / currentFFT.Values[i]
			
			if cmplx.IsNaN(impedance[i]) || cmplx.IsInf(impedance[i]) {
				return ImpedanceData{}, NewProcessingError("impedance calculation", 
					NewValidationError("Impedance", fmt.Sprintf("invalid impedance value at frequency index %d", i)))
			}
		}
	}

	impedanceData := ImpedanceData{
		Timestamp:   voltageSignal.Timestamp,
		Impedance:   impedance,
		Frequencies: voltageFFT.Frequencies,
	}

	magnitude, phase := impedanceData.CalculateMagnitudePhase()
	impedanceData.Magnitude = magnitude
	impedanceData.Phase = phase

	if err := ic.validator.ValidateImpedanceData(impedanceData); err != nil {
		return ImpedanceData{}, NewProcessingError("impedance data validation", err)
	}

	return impedanceData, nil
}

func (ic *DefaultImpedanceCalculator) ProcessEISMeasurement(voltageSignal, currentSignal Signal) (EISMeasurement, error) {
	if err := ic.ValidateSignals(voltageSignal, currentSignal); err != nil {
		return EISMeasurement{}, NewProcessingError("signal validation", err)
	}

	impedanceData, err := ic.CalculateImpedance(voltageSignal, currentSignal)
	if err != nil {
		return EISMeasurement{}, NewProcessingError("impedance calculation", err)
	}

	voltageFFT, err := ic.fftProcessor.ProcessSignal(voltageSignal)
	if err != nil {
		return EISMeasurement{}, NewProcessingError("voltage FFT processing", err)
	}
	
	currentFFT, err := ic.fftProcessor.ProcessSignal(currentSignal)
	if err != nil {
		return EISMeasurement{}, NewProcessingError("current FFT processing", err)
	}

	voltageFFT, err = ic.fftProcessor.GetPositiveFrequencies(voltageFFT)
	if err != nil {
		return EISMeasurement{}, NewProcessingError("voltage positive frequencies", err)
	}
	
	currentFFT, err = ic.fftProcessor.GetPositiveFrequencies(currentFFT)
	if err != nil {
		return EISMeasurement{}, NewProcessingError("current positive frequencies", err)
	}

	measurement := EISMeasurement{
		Voltage:   voltageFFT,
		Current:   currentFFT,
		Impedance: impedanceData,
	}

	return measurement, nil
}