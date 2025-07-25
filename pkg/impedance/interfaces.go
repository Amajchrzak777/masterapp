package impedance

import (
	"github.com/adam/masterapp/pkg/signal"
)

// Calculator defines the interface for impedance calculations
type Calculator interface {
	CalculateImpedance(voltageSignal, currentSignal signal.Signal) (signal.ImpedanceData, error)
	ProcessEISMeasurement(voltageSignal, currentSignal signal.Signal) (signal.EISMeasurement, error)
	ValidateSignals(voltageSignal, currentSignal signal.Signal) error
}