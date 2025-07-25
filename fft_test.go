package main

import (
	"math"
	"math/cmplx"
	"testing"
	"time"
)

func TestDefaultFFTProcessor_ProcessSignal(t *testing.T) {
	fftProcessor := NewFFTProcessor()

	tests := []struct {
		name    string
		signal  Signal
		wantErr bool
	}{
		{
			name: "valid signal - power of 2",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 0.0, 1.0, 0.0},
				SampleRate: 4.0,
			},
			wantErr: false,
		},
		{
			name: "valid signal - non-power of 2",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 2.0, 3.0},
				SampleRate: 3.0,
			},
			wantErr: false,
		},
		{
			name: "empty signal",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{},
				SampleRate: 1000.0,
			},
			wantErr: true,
		},
		{
			name: "invalid sample rate",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 2.0},
				SampleRate: 0,
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

func TestDefaultFFTProcessor_GetPositiveFrequencies(t *testing.T) {
	fftProcessor := NewFFTProcessor()

	tests := []struct {
		name          string
		complexSignal ComplexSignal
		wantErr       bool
		expectedLen   int
	}{
		{
			name: "valid complex signal - even length",
			complexSignal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1), complex(3, 2), complex(4, 3)},
				Frequencies: []float64{0, 100, 200, 300},
			},
			wantErr:     false,
			expectedLen: 2,
		},
		{
			name: "valid complex signal - odd length",
			complexSignal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1), complex(3, 2)},
				Frequencies: []float64{0, 100, 200},
			},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name: "single value",
			complexSignal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0)},
				Frequencies: []float64{0},
			},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name: "empty signal",
			complexSignal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{},
				Frequencies: []float64{},
			},
			wantErr: true,
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

				if len(result.Frequencies) != tt.expectedLen {
					t.Errorf("Expected %d positive frequencies, got %d", tt.expectedLen, len(result.Frequencies))
				}

				if result.Timestamp != tt.complexSignal.Timestamp {
					t.Errorf("Timestamp mismatch")
				}
			}
		})
	}
}

func TestFFTRoundTrip(t *testing.T) {
	fftProcessor := NewFFTProcessor()

	signal := Signal{
		Timestamp:  time.Now(),
		Values:     []float64{1.0, 2.0, 3.0, 4.0},
		SampleRate: 4.0,
	}

	complexSignal, err := fftProcessor.ProcessSignal(signal)
	if err != nil {
		t.Fatalf("FFT processing failed: %v", err)
	}

	if len(complexSignal.Values) != len(signal.Values) {
		t.Errorf("FFT result length mismatch: expected %d, got %d", len(signal.Values), len(complexSignal.Values))
	}

	for i, freq := range complexSignal.Frequencies {
		expectedFreq := float64(i) * signal.SampleRate / float64(len(signal.Values))
		if i >= len(signal.Values)/2 {
			expectedFreq = float64(i-len(signal.Values)) * signal.SampleRate / float64(len(signal.Values))
		}
		if math.Abs(freq-expectedFreq) > 1e-10 {
			t.Errorf("Frequency[%d]: expected %f, got %f", i, expectedFreq, freq)
		}
	}
}

func TestFFTKnownValues(t *testing.T) {
	fftProcessor := NewFFTProcessor()

	signal := Signal{
		Timestamp:  time.Now(),
		Values:     []float64{1.0, 0.0, 0.0, 0.0},
		SampleRate: 4.0,
	}

	result, err := fftProcessor.ProcessSignal(signal)
	if err != nil {
		t.Fatalf("FFT processing failed: %v", err)
	}

	expectedFirst := complex(1.0, 0.0)
	if cmplx.Abs(result.Values[0]-expectedFirst) > 1e-10 {
		t.Errorf("FFT[0]: expected %v, got %v", expectedFirst, result.Values[0])
	}

	for i := 1; i < len(result.Values); i++ {
		expected := complex(1.0, 0.0)
		if cmplx.Abs(result.Values[i]-expected) > 1e-10 {
			t.Errorf("FFT[%d]: expected %v, got %v", i, expected, result.Values[i])
		}
	}
}

func TestFFTProcessorValidation(t *testing.T) {
	fftProcessor := NewFFTProcessor()

	invalidSignal := Signal{
		Timestamp:  time.Time{},
		Values:     []float64{math.NaN(), 2.0},
		SampleRate: 0,
	}

	_, err := fftProcessor.ProcessSignal(invalidSignal)
	if err == nil {
		t.Error("Expected error for invalid signal, got nil")
	}

	err = fftProcessor.ValidateSignal(invalidSignal)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}
}