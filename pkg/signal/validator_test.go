package signal

import (
	"math"
	"testing"
	"time"

	"github.com/adam/masterapp/pkg/config"
)

func TestDefaultValidator_ValidateSignal(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		signal  Signal
		wantErr bool
		errType string
	}{
		{
			name: "valid signal",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 2.0, 3.0},
				SampleRate: 1000.0,
			},
			wantErr: false,
		},
		{
			name: "empty values",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{},
				SampleRate: 1000.0,
			},
			wantErr: true,
			errType: "Values",
		},
		{
			name: "zero sample rate",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 2.0},
				SampleRate: 0,
			},
			wantErr: true,
			errType: "SampleRate",
		},
		{
			name: "NaN value",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, math.NaN(), 3.0},
				SampleRate: 1000.0,
			},
			wantErr: true,
			errType: "Values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSignal(tt.signal)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSignal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil {
				if validationErr, ok := err.(config.ValidationError); ok {
					if validationErr.Field != tt.errType {
						t.Errorf("Expected error field %s, got %s", tt.errType, validationErr.Field)
					}
				}
			}
		})
	}
}

func TestValidateSignalsMatch(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		voltageSignal   Signal
		currentSignal   Signal
		wantErr         bool
		expectedErrType error
	}{
		{
			name: "matching signals",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0},
				SampleRate: 1000.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2, 0.3},
				SampleRate: 1000.0,
			},
			wantErr: false,
		},
		{
			name: "mismatched length",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0},
				SampleRate: 1000.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2},
				SampleRate: 1000.0,
			},
			wantErr:         true,
			expectedErrType: config.ErrMismatchedSignalLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSignalsMatch(tt.voltageSignal, tt.currentSignal)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSignalsMatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil && tt.expectedErrType != nil {
				if err != tt.expectedErrType {
					t.Errorf("Expected error %v, got %v", tt.expectedErrType, err)
				}
			}
		})
	}
}