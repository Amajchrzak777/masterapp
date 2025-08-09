package impedance

import (
	"math"
	"math/cmplx"
	"time"

	"github.com/adam/masterapp/pkg/signal"
)

// EISGenerator generates EIS data directly from circuit models (like the Python code)
type EISGenerator struct {
	spectrumCounter int
}

// NewEISGenerator creates a new EIS data generator
func NewEISGenerator() *EISGenerator {
	return &EISGenerator{
		spectrumCounter: 0,
	}
}

// CircuitParameters defines time-varying parameters for R_s + (R_ct || CPE) model
type CircuitParameters struct {
	Rs         float64 // Solution resistance (constant)
	RctInitial float64 // Initial charge transfer resistance  
	RctGrowth  float64 // Growth rate per spectrum
	Q          float64 // CPE coefficient
	N          float64 // CPE exponent
}

// GenerateLogFrequencies creates logarithmically spaced frequencies like the Python code
func (g *EISGenerator) GenerateLogFrequencies(numPoints int) []float64 {
	// Python: frequencies = np.logspace(5, -2, 50)  # 100kHz to 0.01Hz
	frequencies := make([]float64, numPoints)
	
	logStart := 5.0  // log10(100000) = 5 (100 kHz)
	logEnd := -2.0   // log10(0.01) = -2 (0.01 Hz)
	
	for i := 0; i < numPoints; i++ {
		logFreq := logStart + float64(i)*(logEnd-logStart)/float64(numPoints-1)
		frequencies[i] = math.Pow(10, logFreq)
	}
	
	return frequencies
}

// GenerateEISSpectrum generates one EIS spectrum for current time point
// This replicates the Python circuit calculation exactly
func (g *EISGenerator) GenerateEISSpectrum(params CircuitParameters) signal.ImpedanceData {
	frequencies := g.GenerateLogFrequencies(50) // 50 points like Python code
	
	// Calculate time-varying R_ct: R_ct = R_ct_initial + i * 8
	Rct := params.RctInitial + float64(g.spectrumCounter)*params.RctGrowth
	
	impedance := make([]complex128, len(frequencies))
	
	for i, freq := range frequencies {
		// Python code equivalent:
		// w = 2 * np.pi * frequencies
		// Z_cpe = 1 / (Q * (1j * w)**n)
		// Z_parallel = (R_ct * Z_cpe) / (R_ct + Z_cpe)
		// Z_total = R_s + Z_parallel
		
		w := 2 * math.Pi * freq
		
		// CPE impedance: Z_cpe = 1 / (Q * (jω)^n)
		jwPowN := cmplx.Pow(complex(0, w), complex(params.N, 0))
		ZCpe := complex(1, 0) / (complex(params.Q, 0) * jwPowN)
		
		// Parallel combination: Z_parallel = (R_ct * Z_cpe) / (R_ct + Z_cpe)  
		RctComplex := complex(Rct, 0)
		ZParallel := (RctComplex * ZCpe) / (RctComplex + ZCpe)
		
		// Total impedance: Z_total = R_s + Z_parallel
		ZTotal := complex(params.Rs, 0) + ZParallel
		
		impedance[i] = ZTotal
	}
	
	// Create ImpedanceData structure
	data := signal.ImpedanceData{
		Timestamp:   time.Now(),
		Impedance:   impedance,
		Frequencies: frequencies,
	}
	
	// Calculate magnitude and phase
	magnitude, phase := data.CalculateMagnitudePhase()
	data.Magnitude = magnitude
	data.Phase = phase
	
	// Increment spectrum counter for next call (simulates time evolution)
	g.spectrumCounter++
	
	return data
}

// GetDefaultParameters returns the same parameters as Python code
func (g *EISGenerator) GetDefaultParameters() CircuitParameters {
	return CircuitParameters{
		Rs:         10.0,   // R_s = 10
		RctInitial: 20.0,   // R_ct_initial = 20  
		RctGrowth:  8.0,    // Growth: i * 8
		Q:          1e-5,   // Q = 1e-5
		N:          0.85,   // n = 0.85
	}
}

// ResetCounter resets the spectrum counter (useful for testing)
func (g *EISGenerator) ResetCounter() {
	g.spectrumCounter = 0
}

// GetCurrentSpectrum returns current spectrum number
func (g *EISGenerator) GetCurrentSpectrum() int {
	return g.spectrumCounter
}