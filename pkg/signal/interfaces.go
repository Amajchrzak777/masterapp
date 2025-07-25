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

// DataLoader provides capabilities for loading signals from files
type DataLoader interface {
	LoadSignalFromCSV(filename string, sampleRate float64) ([]Signal, error)
	LoadVoltageAndCurrentFromCSV(voltageFile, currentFile string, sampleRate float64) ([]Signal, []Signal, error)
}