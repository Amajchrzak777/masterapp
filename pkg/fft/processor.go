package fft

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/adam/masterapp/pkg/config"
	"github.com/adam/masterapp/pkg/signal"
)

// DefaultProcessor implements FFT processing with validation
type DefaultProcessor struct {
	validator signal.Validator
}

// NewProcessor creates a new FFT processor
func NewProcessor() Processor {
	return &DefaultProcessor{
		validator: signal.NewValidator(),
	}
}

// ValidateSignal validates the input signal for FFT processing
func (fft *DefaultProcessor) ValidateSignal(sig signal.Signal) error {
	return fft.validator.ValidateSignal(sig)
}

// ProcessSignal performs FFT on the input signal and returns frequency domain representation
func (fft *DefaultProcessor) ProcessSignal(sig signal.Signal) (signal.ComplexSignal, error) {
	if err := fft.ValidateSignal(sig); err != nil {
		return signal.ComplexSignal{}, config.NewProcessingError("signal validation", err)
	}

	n := len(sig.Values)
	
	if n == 0 {
		return signal.ComplexSignal{}, config.NewProcessingError("FFT processing", config.ErrInvalidSignalLength)
	}
	
	complexValues := make([]complex128, n)
	for i, val := range sig.Values {
		complexValues[i] = complex(val, 0)
	}

	fftResult, err := fft.computeFFT(complexValues)
	if err != nil {
		return signal.ComplexSignal{}, config.NewProcessingError("FFT computation", err)
	}
	
	frequencies, err := fft.generateFrequencies(n, sig.SampleRate)
	if err != nil {
		return signal.ComplexSignal{}, config.NewProcessingError("frequency generation", err)
	}

	result := signal.ComplexSignal{
		Timestamp:   sig.Timestamp,
		Values:      fftResult,
		Frequencies: frequencies,
	}

	if err := fft.validator.ValidateComplexSignal(result); err != nil {
		return signal.ComplexSignal{}, config.NewProcessingError("result validation", err)
	}

	return result, nil
}

// GetPositiveFrequencies extracts only the positive frequency components
func (fft *DefaultProcessor) GetPositiveFrequencies(complexSignal signal.ComplexSignal) (signal.ComplexSignal, error) {
	if err := fft.validator.ValidateComplexSignal(complexSignal); err != nil {
		return signal.ComplexSignal{}, config.NewProcessingError("input validation", err)
	}
	
	n := len(complexSignal.Values)
	if n == 0 {
		return signal.ComplexSignal{}, config.ErrInvalidSignalLength
	}
	
	halfN := n / 2
	if halfN == 0 {
		halfN = 1
	}

	result := signal.ComplexSignal{
		Timestamp:   complexSignal.Timestamp,
		Values:      complexSignal.Values[:halfN],
		Frequencies: complexSignal.Frequencies[:halfN],
	}

	if err := fft.validator.ValidatePositiveFrequencySignal(result); err != nil {
		return signal.ComplexSignal{}, config.NewProcessingError("result validation", err)
	}

	return result, nil
}

// computeFFT performs the actual FFT computation using radix-2 algorithm
func (fft *DefaultProcessor) computeFFT(x []complex128) ([]complex128, error) {
	n := len(x)
	if n <= 0 {
		return nil, config.ErrInvalidSignalLength
	}
	
	if n <= 1 {
		return x, nil
	}

	if n%2 != 0 {
		return fft.dft(x)
	}

	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)

	for i := 0; i < n/2; i++ {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}

	evenFFT, err := fft.computeFFT(even)
	if err != nil {
		return nil, err
	}
	
	oddFFT, err := fft.computeFFT(odd)
	if err != nil {
		return nil, err
	}

	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		angle := -2 * math.Pi * float64(k) / float64(n)
		if math.IsNaN(angle) || math.IsInf(angle, 0) {
			return nil, config.NewProcessingError("FFT computation", fmt.Errorf("invalid angle at k=%d", k))
		}
		
		t := cmplx.Exp(complex(0, angle)) * oddFFT[k]
		result[k] = evenFFT[k] + t
		result[k+n/2] = evenFFT[k] - t
	}

	return result, nil
}

// dft performs discrete Fourier transform for non-power-of-2 lengths
func (fft *DefaultProcessor) dft(x []complex128) ([]complex128, error) {
	n := len(x)
	if n <= 0 {
		return nil, config.ErrInvalidSignalLength
	}
	
	result := make([]complex128, n)

	for k := 0; k < n; k++ {
		sum := complex(0, 0)
		for j := 0; j < n; j++ {
			angle := -2 * math.Pi * float64(k) * float64(j) / float64(n)
			if math.IsNaN(angle) || math.IsInf(angle, 0) {
				return nil, config.NewProcessingError("DFT computation", fmt.Errorf("invalid angle at k=%d, j=%d", k, j))
			}
			sum += x[j] * cmplx.Exp(complex(0, angle))
		}
		result[k] = sum
	}

	return result, nil
}

// generateFrequencies creates the frequency array for FFT results
func (fft *DefaultProcessor) generateFrequencies(n int, sampleRate float64) ([]float64, error) {
	if n <= 0 {
		return nil, config.ErrInvalidSignalLength
	}
	
	if sampleRate <= 0 {
		return nil, config.ErrInvalidSampleRate
	}
	
	frequencies := make([]float64, n)
	
	for i := 0; i < n; i++ {
		if i < n/2 {
			frequencies[i] = float64(i) * sampleRate / float64(n)
		} else {
			frequencies[i] = float64(i-n) * sampleRate / float64(n)
		}
	}

	return frequencies, nil
}