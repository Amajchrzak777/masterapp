package main

import (
	"math"
	"math/cmplx"
	"testing"
	"time"
)

func TestDefaultImpedanceCalculator_CalculateImpedance(t *testing.T) {
	calculator := NewImpedanceCalculator()
	now := time.Now()

	tests := []struct {
		name          string
		voltageSignal Signal
		currentSignal Signal
		wantErr       bool
	}{
		{
			name: "valid signals",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0, 4.0},
				SampleRate: 4.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2, 0.3, 0.4},
				SampleRate: 4.0,
			},
			wantErr: false,
		},
		{
			name: "mismatched signal lengths",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0, 4.0},
				SampleRate: 4.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2, 0.3},
				SampleRate: 4.0,
			},
			wantErr: true,
		},
		{
			name: "mismatched sample rates",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0, 4.0},
				SampleRate: 4.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2, 0.3, 0.4},
				SampleRate: 8.0,
			},
			wantErr: true,
		},
		{
			name: "empty signals",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{},
				SampleRate: 4.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{},
				SampleRate: 4.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.CalculateImpedance(tt.voltageSignal, tt.currentSignal)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateImpedance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.Impedance) == 0 {
					t.Error("Expected non-empty impedance values")
				}

				if len(result.Frequencies) == 0 {
					t.Error("Expected non-empty frequency values")
				}

				if len(result.Magnitude) != len(result.Impedance) {
					t.Error("Magnitude array length should match impedance array length")
				}

				if len(result.Phase) != len(result.Impedance) {
					t.Error("Phase array length should match impedance array length")
				}

				if result.Timestamp != tt.voltageSignal.Timestamp {
					t.Error("Result timestamp should match input timestamp")
				}

				for i, imp := range result.Impedance {
					if cmplx.IsNaN(imp) || cmplx.IsInf(imp) {
						t.Errorf("Invalid impedance value at index %d: %v", i, imp)
					}
				}

				for i, mag := range result.Magnitude {
					if math.IsNaN(mag) || math.IsInf(mag, 0) || mag < 0 {
						t.Errorf("Invalid magnitude value at index %d: %f", i, mag)
					}
				}

				for i, phase := range result.Phase {
					if math.IsNaN(phase) || math.IsInf(phase, 0) {
						t.Errorf("Invalid phase value at index %d: %f", i, phase)
					}
				}
			}
		})
	}
}

func TestDefaultImpedanceCalculator_ProcessEISMeasurement(t *testing.T) {
	calculator := NewImpedanceCalculator()
	now := time.Now()

	voltageSignal := Signal{
		Timestamp:  now,
		Values:     []float64{1.0, 2.0, 3.0, 4.0},
		SampleRate: 4.0,
	}

	currentSignal := Signal{
		Timestamp:  now,
		Values:     []float64{0.1, 0.2, 0.3, 0.4},
		SampleRate: 4.0,
	}

	measurement, err := calculator.ProcessEISMeasurement(voltageSignal, currentSignal)
	if err != nil {
		t.Fatalf("ProcessEISMeasurement() failed: %v", err)
	}

	if len(measurement.Voltage.Values) == 0 {
		t.Error("Expected non-empty voltage FFT values")
	}

	if len(measurement.Current.Values) == 0 {
		t.Error("Expected non-empty current FFT values")
	}

	if len(measurement.Impedance.Impedance) == 0 {
		t.Error("Expected non-empty impedance values")
	}

	if len(measurement.Voltage.Values) != len(measurement.Current.Values) {
		t.Error("Voltage and current FFT should have same length")
	}

	if len(measurement.Voltage.Frequencies) != len(measurement.Current.Frequencies) {
		t.Error("Voltage and current frequencies should have same length")
	}

	if measurement.Voltage.Timestamp != voltageSignal.Timestamp {
		t.Error("Voltage timestamp mismatch")
	}

	if measurement.Current.Timestamp != currentSignal.Timestamp {
		t.Error("Current timestamp mismatch")
	}
}

func TestImpedanceCalculationWithZeroCurrent(t *testing.T) {
	calculator := NewImpedanceCalculator()
	now := time.Now()

	voltageSignal := Signal{
		Timestamp:  now,
		Values:     []float64{1.0, 0.0, 1.0, 0.0},
		SampleRate: 4.0,
	}

	currentSignal := Signal{
		Timestamp:  now,
		Values:     []float64{0.0, 0.0, 0.0, 0.0},
		SampleRate: 4.0,
	}

	result, err := calculator.CalculateImpedance(voltageSignal, currentSignal)
	if err != nil {
		t.Fatalf("CalculateImpedance() failed: %v", err)
	}

	for i, imp := range result.Impedance {
		if cmplx.IsNaN(imp) || cmplx.IsInf(imp) {
			t.Errorf("Expected zero impedance for zero current at index %d, got %v", i, imp)
		}
	}
}

func TestImpedanceValidation(t *testing.T) {
	calculator := NewImpedanceCalculator()
	now := time.Now()

	tests := []struct {
		name          string
		voltageSignal Signal
		currentSignal Signal
		expectError   bool
	}{
		{
			name: "invalid voltage signal - NaN values",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, math.NaN(), 3.0, 4.0},
				SampleRate: 4.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2, 0.3, 0.4},
				SampleRate: 4.0,
			},
			expectError: true,
		},
		{
			name: "invalid current signal - zero timestamp",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0, 4.0},
				SampleRate: 4.0,
			},
			currentSignal: Signal{
				Timestamp:  time.Time{},
				Values:     []float64{0.1, 0.2, 0.3, 0.4},
				SampleRate: 4.0,
			},
			expectError: true,
		},
		{
			name: "valid signals",
			voltageSignal: Signal{
				Timestamp:  now,
				Values:     []float64{1.0, 2.0, 3.0, 4.0},
				SampleRate: 4.0,
			},
			currentSignal: Signal{
				Timestamp:  now,
				Values:     []float64{0.1, 0.2, 0.3, 0.4},
				SampleRate: 4.0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := calculator.ValidateSignals(tt.voltageSignal, tt.currentSignal)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateSignals() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestImpedanceCalculatorKnownValues(t *testing.T) {
	calculator := NewImpedanceCalculator()
	now := time.Now()

	voltageSignal := Signal{
		Timestamp:  now,
		Values:     []float64{2.0, 0.0, 0.0, 0.0},
		SampleRate: 4.0,
	}

	currentSignal := Signal{
		Timestamp:  now,
		Values:     []float64{1.0, 0.0, 0.0, 0.0},
		SampleRate: 4.0,
	}

	result, err := calculator.CalculateImpedance(voltageSignal, currentSignal)
	if err != nil {
		t.Fatalf("CalculateImpedance() failed: %v", err)
	}

	if len(result.Impedance) < 1 {
		t.Fatal("Expected at least one impedance value")
	}

	dcImpedance := result.Impedance[0]
	expectedDCImpedance := complex(2.0, 0.0) 

	tolerance := 1e-10
	if cmplx.Abs(dcImpedance-expectedDCImpedance) > tolerance {
		t.Errorf("DC impedance: expected %v, got %v", expectedDCImpedance, dcImpedance)
	}

	if len(result.Magnitude) < 1 {
		t.Fatal("Expected at least one magnitude value")
	}

	expectedMagnitude := 2.0
	if math.Abs(result.Magnitude[0]-expectedMagnitude) > tolerance {
		t.Errorf("DC magnitude: expected %f, got %f", expectedMagnitude, result.Magnitude[0])
	}

	if len(result.Phase) < 1 {
		t.Fatal("Expected at least one phase value")
	}

	expectedPhase := 0.0
	if math.Abs(result.Phase[0]-expectedPhase) > tolerance {
		t.Errorf("DC phase: expected %f, got %f", expectedPhase, result.Phase[0])
	}
}