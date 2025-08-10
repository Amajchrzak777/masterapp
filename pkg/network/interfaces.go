package network

import (
	"github.com/adam/masterapp/pkg/signal"
)

// Sender defines the interface for sending data over the network
type Sender interface {
	SendEISMeasurement(measurement signal.EISMeasurement) error
	SendImpedanceData(impedanceData signal.ImpedanceData) error
	SendBatchImpedanceData(batch []signal.ImpedanceDataWithIteration) error
	FormatAsJSON(data interface{}) (string, error)
	IsHealthy() bool
}