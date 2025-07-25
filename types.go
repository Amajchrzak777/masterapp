package main

import (
	"math/cmplx"
	"time"
)

type Signal struct {
	Timestamp  time.Time `json:"timestamp"`
	Values     []float64 `json:"values"`
	SampleRate float64   `json:"sample_rate"`
}

type ComplexSignal struct {
	Timestamp   time.Time    `json:"timestamp"`
	Values      []complex128 `json:"values"`
	Frequencies []float64    `json:"frequencies"`
}

type ImpedanceData struct {
	Timestamp   time.Time    `json:"timestamp"`
	Impedance   []complex128 `json:"impedance"`
	Frequencies []float64    `json:"frequencies"`
	Magnitude   []float64    `json:"magnitude"`
	Phase       []float64    `json:"phase"`
}

type EISMeasurement struct {
	Voltage   ComplexSignal `json:"voltage"`
	Current   ComplexSignal `json:"current"`
	Impedance ImpedanceData `json:"impedance"`
}

func (z ImpedanceData) CalculateMagnitudePhase() ([]float64, []float64) {
	magnitude := make([]float64, len(z.Impedance))
	phase := make([]float64, len(z.Impedance))

	for i, imp := range z.Impedance {
		magnitude[i] = cmplx.Abs(imp)
		phase[i] = cmplx.Phase(imp)
	}

	return magnitude, phase
}
