package main

import (
	"fmt"
	"math"
	"math/cmplx"
)

type DefaultFFTProcessor struct {
	validator SignalValidator
}

func NewFFTProcessor() FFTProcessor {
	return &DefaultFFTProcessor{
		validator: NewSignalValidator(),
	}
}

func (fft *DefaultFFTProcessor) ValidateSignal(signal Signal) error {
	return fft.validator.ValidateSignal(signal)
}

func (fft *DefaultFFTProcessor) ProcessSignal(signal Signal) (ComplexSignal, error) {
	if err := fft.ValidateSignal(signal); err != nil {
		return ComplexSignal{}, NewProcessingError("signal validation", err)
	}
	n := len(signal.Values)
	
	if n == 0 {
		return ComplexSignal{}, NewProcessingError("FFT processing", ErrInvalidSignalLength)
	}
	
	complexValues := make([]complex128, n)
	for i, val := range signal.Values {
		complexValues[i] = complex(val, 0)
	}

	fftResult, err := fft.computeFFT(complexValues)
	if err != nil {
		return ComplexSignal{}, NewProcessingError("FFT computation", err)
	}
	
	frequencies, err := fft.generateFrequencies(n, signal.SampleRate)
	if err != nil {
		return ComplexSignal{}, NewProcessingError("frequency generation", err)
	}

	result := ComplexSignal{
		Timestamp:   signal.Timestamp,
		Values:      fftResult,
		Frequencies: frequencies,
	}

	if err := fft.validator.ValidateComplexSignal(result); err != nil {
		return ComplexSignal{}, NewProcessingError("result validation", err)
	}

	return result, nil
}

func (fft *DefaultFFTProcessor) computeFFT(x []complex128) ([]complex128, error) {
	n := len(x)
	if n <= 0 {
		return nil, ErrInvalidSignalLength
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
			return nil, NewProcessingError("FFT computation", fmt.Errorf("invalid angle at k=%d", k))
		}
		
		t := cmplx.Exp(complex(0, angle)) * oddFFT[k]
		result[k] = evenFFT[k] + t
		result[k+n/2] = evenFFT[k] - t
	}

	return result, nil
}

func (fft *DefaultFFTProcessor) dft(x []complex128) ([]complex128, error) {
	n := len(x)
	if n <= 0 {
		return nil, ErrInvalidSignalLength
	}
	
	result := make([]complex128, n)

	for k := 0; k < n; k++ {
		sum := complex(0, 0)
		for j := 0; j < n; j++ {
			angle := -2 * math.Pi * float64(k) * float64(j) / float64(n)
			if math.IsNaN(angle) || math.IsInf(angle, 0) {
				return nil, NewProcessingError("DFT computation", fmt.Errorf("invalid angle at k=%d, j=%d", k, j))
			}
			sum += x[j] * cmplx.Exp(complex(0, angle))
		}
		result[k] = sum
	}

	return result, nil
}

func (fft *DefaultFFTProcessor) generateFrequencies(n int, sampleRate float64) ([]float64, error) {
	if n <= 0 {
		return nil, ErrInvalidSignalLength
	}
	
	if sampleRate <= 0 {
		return nil, ErrInvalidSampleRate
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

func (fft *DefaultFFTProcessor) GetPositiveFrequencies(complexSignal ComplexSignal) (ComplexSignal, error) {
	if err := fft.validator.ValidateComplexSignal(complexSignal); err != nil {
		return ComplexSignal{}, NewProcessingError("input validation", err)
	}
	
	n := len(complexSignal.Values)
	if n == 0 {
		return ComplexSignal{}, ErrInvalidSignalLength
	}
	
	halfN := n / 2
	if halfN == 0 {
		halfN = 1
	}

	result := ComplexSignal{
		Timestamp:   complexSignal.Timestamp,
		Values:      complexSignal.Values[:halfN],
		Frequencies: complexSignal.Frequencies[:halfN],
	}

	if err := fft.validator.ValidatePositiveFrequencySignal(result); err != nil {
		return ComplexSignal{}, NewProcessingError("result validation", err)
	}

	return result, nil
}