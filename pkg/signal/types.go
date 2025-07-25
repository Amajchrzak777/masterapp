package signal

import (
	"encoding/json"
	"math/cmplx"
	"time"
)

// Signal represents a time-domain signal with associated metadata
type Signal struct {
	Timestamp  time.Time `json:"timestamp"`
	Values     []float64 `json:"values"`
	SampleRate float64   `json:"sample_rate"`
}

// ComplexSignal represents a frequency-domain signal after FFT processing
type ComplexSignal struct {
	Timestamp   time.Time    `json:"timestamp"`
	Values      []complex128 `json:"-"`
	Frequencies []float64    `json:"frequencies"`
}

// MarshalJSON custom JSON marshaling for ComplexSignal
func (cs ComplexSignal) MarshalJSON() ([]byte, error) {
	type Alias ComplexSignal
	complexValues := make([]map[string]float64, len(cs.Values))
	for i, v := range cs.Values {
		complexValues[i] = map[string]float64{
			"real": real(v),
			"imag": imag(v),
		}
	}
	return json.Marshal(&struct {
		Values []map[string]float64 `json:"values"`
		*Alias
	}{
		Values: complexValues,
		Alias:  (*Alias)(&cs),
	})
}

// ImpedanceData represents calculated impedance with magnitude and phase
type ImpedanceData struct {
	Timestamp   time.Time    `json:"timestamp"`
	Impedance   []complex128 `json:"-"`
	Frequencies []float64    `json:"frequencies"`
	Magnitude   []float64    `json:"magnitude"`
	Phase       []float64    `json:"phase"`
}

// MarshalJSON custom JSON marshaling for ImpedanceData
func (id ImpedanceData) MarshalJSON() ([]byte, error) {
	type Alias ImpedanceData
	complexImpedance := make([]map[string]float64, len(id.Impedance))
	for i, v := range id.Impedance {
		complexImpedance[i] = map[string]float64{
			"real": real(v),
			"imag": imag(v),
		}
	}
	return json.Marshal(&struct {
		Impedance []map[string]float64 `json:"impedance"`
		*Alias
	}{
		Impedance: complexImpedance,
		Alias:     (*Alias)(&id),
	})
}

// EISMeasurement represents a complete electrochemical impedance spectroscopy measurement
type EISMeasurement struct {
	Voltage   ComplexSignal `json:"voltage"`
	Current   ComplexSignal `json:"current"`
	Impedance ImpedanceData `json:"impedance"`
}

// CalculateMagnitudePhase calculates the magnitude and phase from complex impedance values
func (z *ImpedanceData) CalculateMagnitudePhase() ([]float64, []float64) {
	magnitude := make([]float64, len(z.Impedance))
	phase := make([]float64, len(z.Impedance))
	
	for i, imp := range z.Impedance {
		magnitude[i] = cmplx.Abs(imp)
		phase[i] = cmplx.Phase(imp)
	}
	
	return magnitude, phase
}

// IsEmpty returns true if the signal contains no data
func (s *Signal) IsEmpty() bool {
	return len(s.Values) == 0
}

// Length returns the number of samples in the signal
func (s *Signal) Length() int {
	return len(s.Values)
}

// Duration returns the duration of the signal in seconds
func (s *Signal) Duration() float64 {
	if s.SampleRate <= 0 || len(s.Values) == 0 {
		return 0
	}
	return float64(len(s.Values)) / s.SampleRate
}

// IsEmpty returns true if the complex signal contains no data
func (cs *ComplexSignal) IsEmpty() bool {
	return len(cs.Values) == 0
}

// Length returns the number of samples in the complex signal
func (cs *ComplexSignal) Length() int {
	return len(cs.Values)
}

// IsEmpty returns true if the impedance data contains no data
func (z *ImpedanceData) IsEmpty() bool {
	return len(z.Impedance) == 0
}

// Length returns the number of impedance measurements
func (z *ImpedanceData) Length() int {
	return len(z.Impedance)
}