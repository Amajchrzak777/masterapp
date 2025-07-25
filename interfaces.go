package main

import "context"

type DataReceiver interface {
	StartReceiving(ctx context.Context) error
	GetVoltageChannel() <-chan Signal
	GetCurrentChannel() <-chan Signal
	Stop() error
}

type FFTProcessor interface {
	ProcessSignal(signal Signal) (ComplexSignal, error)
	GetPositiveFrequencies(complexSignal ComplexSignal) (ComplexSignal, error)
	ValidateSignal(signal Signal) error
}

type ImpedanceCalculator interface {
	CalculateImpedance(voltageSignal, currentSignal Signal) (ImpedanceData, error)
	ProcessEISMeasurement(voltageSignal, currentSignal Signal) (EISMeasurement, error)
	ValidateSignals(voltageSignal, currentSignal Signal) error
}

type DataSender interface {
	SendEISMeasurement(measurement EISMeasurement) error
	SendImpedanceData(impedanceData ImpedanceData) error
	FormatAsJSON(data interface{}) (string, error)
	IsHealthy() bool
}

type SignalValidator interface {
	ValidateSignal(signal Signal) error
	ValidateComplexSignal(signal ComplexSignal) error
	ValidatePositiveFrequencySignal(signal ComplexSignal) error
	ValidateImpedanceData(data ImpedanceData) error
}

type SignalGenerator interface {
	GenerateVoltageSignal(sampleRate float64, samplesPerSecond int) (Signal, error)
	GenerateCurrentSignal(sampleRate float64, samplesPerSecond int) (Signal, error)
}