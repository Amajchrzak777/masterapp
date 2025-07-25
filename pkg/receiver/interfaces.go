package receiver

import (
	"context"

	"github.com/adam/masterapp/pkg/signal"
)

// DataReceiver defines the interface for real-time signal reception
type DataReceiver interface {
	StartReceiving(ctx context.Context) error
	GetVoltageChannel() <-chan signal.Signal
	GetCurrentChannel() <-chan signal.Signal
	Stop() error
}