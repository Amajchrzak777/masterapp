package signal

// Validator provides validation capabilities for signal data
type Validator interface {
	ValidateSignal(signal Signal) error
	ValidateComplexSignal(signal ComplexSignal) error
	ValidatePositiveFrequencySignal(signal ComplexSignal) error
	ValidateImpedanceData(data ImpedanceData) error
}

// Generator provides signal generation capabilities for testing and simulation
type Generator interface {
	GenerateVoltageSignal(sampleRate float64, samplesPerSecond int) (Signal, error)
	GenerateCurrentSignal(sampleRate float64, samplesPerSecond int) (Signal, error)
}