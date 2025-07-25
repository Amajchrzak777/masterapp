package main

import (
	"math"
	"testing"
	"time"
)

func TestDefaultSignalValidator_ValidateSignal(t *testing.T) {
	validator := NewSignalValidator()

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
			name: "negative sample rate",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, 2.0},
				SampleRate: -100.0,
			},
			wantErr: true,
			errType: "SampleRate",
		},
		{
			name: "zero timestamp",
			signal: Signal{
				Timestamp:  time.Time{},
				Values:     []float64{1.0, 2.0},
				SampleRate: 1000.0,
			},
			wantErr: true,
			errType: "Timestamp",
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
		{
			name: "infinite value",
			signal: Signal{
				Timestamp:  time.Now(),
				Values:     []float64{1.0, math.Inf(1), 3.0},
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
				if validationErr, ok := err.(ValidationError); ok {
					if validationErr.Field != tt.errType {
						t.Errorf("Expected error field %s, got %s", tt.errType, validationErr.Field)
					}
				}
			}
		})
	}
}

func TestDefaultSignalValidator_ValidateComplexSignal(t *testing.T) {
	validator := NewSignalValidator()

	tests := []struct {
		name    string
		signal  ComplexSignal
		wantErr bool
		errType string
	}{
		{
			name: "valid complex signal",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1)},
				Frequencies: []float64{0, 100},
			},
			wantErr: false,
		},
		{
			name: "empty values",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{},
				Frequencies: []float64{0, 100},
			},
			wantErr: true,
			errType: "Values",
		},
		{
			name: "empty frequencies",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1)},
				Frequencies: []float64{},
			},
			wantErr: true,
			errType: "Frequencies",
		},
		{
			name: "mismatched lengths",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1)},
				Frequencies: []float64{0},
			},
			wantErr: true,
			errType: "Length",
		},
		{
			name: "NaN complex value",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(math.NaN(), 1)},
				Frequencies: []float64{0, 100},
			},
			wantErr: true,
			errType: "Values",
		},
		{
			name: "negative frequency (allowed in full FFT spectrum)",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1)},
				Frequencies: []float64{0, -100},
			},
			wantErr: false, // Negative frequencies are now allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateComplexSignal(tt.signal)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateComplexSignal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil {
				if validationErr, ok := err.(ValidationError); ok {
					if validationErr.Field != tt.errType {
						t.Errorf("Expected error field %s, got %s", tt.errType, validationErr.Field)
					}
				}
			}
		})
	}
}

func TestDefaultSignalValidator_ValidatePositiveFrequencySignal(t *testing.T) {
	validator := NewSignalValidator()

	tests := []struct {
		name    string
		signal  ComplexSignal
		wantErr bool
		errType string
	}{
		{
			name: "valid positive frequency signal",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1)},
				Frequencies: []float64{0, 100},
			},
			wantErr: false,
		},
		{
			name: "negative frequency (should be rejected)",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{complex(1, 0), complex(2, 1)},
				Frequencies: []float64{0, -100},
			},
			wantErr: true,
			errType: "Frequencies",
		},
		{
			name: "empty signal",
			signal: ComplexSignal{
				Timestamp:   time.Now(),
				Values:      []complex128{},
				Frequencies: []float64{},
			},
			wantErr: true,
			errType: "Values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePositiveFrequencySignal(tt.signal)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePositiveFrequencySignal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil {
				if validationErr, ok := err.(ValidationError); ok {
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
			expectedErrType: ErrMismatchedSignalLength,
		},
		{
			name: "mismatched sample rate",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0},
				SampleRate: 1000.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2, 0.3},
				SampleRate: 2000.0,
			},
			wantErr:         true,
			expectedErrType: ErrMismatchedSampleRate,
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

func TestImpedanceData_CalculateMagnitudePhase(t *testing.T) {
	impedanceData := ImpedanceData{
		Timestamp: time.Now(),
		Impedance: []complex128{
			complex(3, 4),   // magnitude = 5, phase = arctan(4/3)
			complex(1, 0),   // magnitude = 1, phase = 0
			complex(0, 1),   // magnitude = 1, phase = π/2
			complex(-1, 0),  // magnitude = 1, phase = π
		},
		Frequencies: []float64{0, 100, 200, 300},
	}

	magnitude, phase := impedanceData.CalculateMagnitudePhase()

	expectedMagnitudes := []float64{5.0, 1.0, 1.0, 1.0}
	expectedPhases := []float64{
		math.Atan2(4, 3),
		0,
		math.Pi / 2,
		math.Pi,
	}

	if len(magnitude) != len(expectedMagnitudes) {
		t.Errorf("Expected %d magnitudes, got %d", len(expectedMagnitudes), len(magnitude))
	}

	if len(phase) != len(expectedPhases) {
		t.Errorf("Expected %d phases, got %d", len(expectedPhases), len(phase))
	}

	for i, expected := range expectedMagnitudes {
		if math.Abs(magnitude[i]-expected) > 1e-10 {
			t.Errorf("Magnitude[%d]: expected %f, got %f", i, expected, magnitude[i])
		}
	}

	for i, expected := range expectedPhases {
		if math.Abs(phase[i]-expected) > 1e-10 {
			t.Errorf("Phase[%d]: expected %f, got %f", i, expected, phase[i])
		}
	}
}