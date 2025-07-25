package fft

import (
	"github.com/adam/masterapp/pkg/signal"
)

// Processor defines the interface for Fast Fourier Transform operations
type Processor interface {
	ProcessSignal(sig signal.Signal) (signal.ComplexSignal, error)
	GetPositiveFrequencies(complexSignal signal.ComplexSignal) (signal.ComplexSignal, error)
	ValidateSignal(sig signal.Signal) error
}