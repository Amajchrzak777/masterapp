package fft

import (
	"testing"
	"time"

	"github.com/adam/masterapp/pkg/signal"
)

func TestDefaultProcessor_ProcessSignal(t *testing.T) {
	fftProcessor := NewProcessor()

	tests := []struct {
		name    string
		signal  signal.Signal
		wantErr bool
	}{
		{
			name: "valid signal - power of 2",
			signal: signal.Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 0.0, 1.0, 0.0},
				SampleRate: 4.0,
			},
			wantErr: false,
		},
		{
			name: "valid signal - non-power of 2",
			signal: signal.Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 2.0, 3.0},
				SampleRate: 3.0,
			},
			wantErr: false,
		},
		{
			name: "empty signal",
			signal: signal.Signal{
				Timestamp:  time.Now(),
				Values:     []float64{},
				SampleRate: 1000.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fftProcessor.ProcessSignal(tt.signal)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessSignal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.Values) != len(tt.signal.Values) {
					t.Errorf("Expected %d FFT values, got %d", len(tt.signal.Values), len(result.Values))
				}

				if len(result.Frequencies) != len(tt.signal.Values) {
					t.Errorf("Expected %d frequencies, got %d", len(tt.signal.Values), len(result.Frequencies))
				}

				if result.Timestamp != tt.signal.Timestamp {
					t.Errorf("Timestamp mismatch")
				}
			}
		})
	}
}

func TestDefaultProcessor_GetPositiveFrequencies(t *testing.T) {
	fftProcessor := NewProcessor()

	tests := []struct {
		name          string
		complexSignal signal.ComplexSignal
		wantErr       bool
		expectedLen   int
	}{
		{
			name: "valid complex signal - even length",
			complexSignal: signal.ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1), complex(3, 2), complex(4, 3)},
				Frequencies: []float64{0, 100, 200, 300},
			},
			wantErr:     false,
			expectedLen: 2,
		},
		{
			name: "single value",
			complexSignal: signal.ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0)},
				Frequencies: []float64{0},
			},
			wantErr:     false,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fftProcessor.GetPositiveFrequencies(tt.complexSignal)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPositiveFrequencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.Values) != tt.expectedLen {
					t.Errorf("Expected %d positive frequency values, got %d", tt.expectedLen, len(result.Values))
				}
			}
		})
	}
}