package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

func main() {
	// EIS measurement parameters
	const (
		duration      = 10.0   // 10 seconds of data
		sampleRate    = 1000.0 // 1000 Hz sampling
		samplesPerSec = 1000   // 1000 samples per second interval
		totalSamples  = int(duration * sampleRate)

		// Electrochemical parameters
		baseVoltage    = 1.0   // 1V base voltage
		baseResistance = 100.0 // 100Ω base resistance
		capacitance    = 1e-6  // 1μF capacitance for RC circuit

		// Noise parameters
		voltageNoise = 0.02 // 2% voltage noise
		currentNoise = 0.01 // 1% current noise
	)

	fmt.Printf("Generating EIS data:\n")
	fmt.Printf("- Duration: %.1f seconds\n", duration)
	fmt.Printf("- Sample rate: %.0f Hz\n", sampleRate)
	fmt.Printf("- Total samples: %d\n", totalSamples)
	fmt.Printf("- Samples per measurement: %d\n", samplesPerSec)

	// Create voltage data
	voltageFile, err := os.Create("examples/data/voltage_10s.csv")
	if err != nil {
		panic(err)
	}
	defer voltageFile.Close()

	voltageWriter := csv.NewWriter(voltageFile)
	defer voltageWriter.Flush()

	// Write CSV header
	voltageWriter.Write([]string{"timestamp", "time_offset", "voltage"})

	// Create current data
	currentFile, err := os.Create("examples/data/current_10s.csv")
	if err != nil {
		panic(err)
	}
	defer currentFile.Close()

	currentWriter := csv.NewWriter(currentFile)
	defer currentWriter.Flush()

	currentWriter.Write([]string{"timestamp", "time_offset", "current"})

	// Generate data
	startTime := time.Now()
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < totalSamples; i++ {
		t := float64(i) / sampleRate
		timestamp := startTime.Add(time.Duration(t * float64(time.Second)))

		// Generate realistic EIS voltage signal with multiple frequency components
		voltage := generateVoltageSignal(t, baseVoltage, voltageNoise)
		current := generateCurrentSignal(t, voltage, baseResistance, capacitance, currentNoise)

		// Write voltage data
		voltageWriter.Write([]string{
			timestamp.Format(time.RFC3339Nano),
			fmt.Sprintf("%.6f", t),
			fmt.Sprintf("%.6f", voltage),
		})

		// Write current data
		currentWriter.Write([]string{
			timestamp.Format(time.RFC3339Nano),
			fmt.Sprintf("%.6f", t),
			fmt.Sprintf("%.6f", current),
		})

		if i%10000 == 0 {
			fmt.Printf("Generated %d/%d samples (%.1f%%)\n", i, totalSamples, float64(i)/float64(totalSamples)*100)
		}
	}

	fmt.Printf("✅ Data generation complete!\n")
	fmt.Printf("📁 Files created:\n")
	fmt.Printf("   - examples/data/voltage_10s.csv (%d samples)\n", totalSamples)
	fmt.Printf("   - examples/data/current_10s.csv (%d samples)\n", totalSamples)
}

func generateVoltageSignal(t, baseVoltage, noiseLevel float64) float64 {
	// Multi-frequency voltage signal typical for EIS
	signal := baseVoltage

	// Add fundamental frequency (10 Hz)
	signal += 0.3 * math.Sin(2*math.Pi*10*t)

	// Add harmonics for realistic EIS
	signal += 0.1 * math.Sin(2*math.Pi*25*t+math.Pi/4)
	signal += 0.05 * math.Sin(2*math.Pi*50*t+math.Pi/3)

	// Add some low-frequency drift
	signal += 0.02 * math.Sin(2*math.Pi*0.5*t)

	// Add noise
	signal += noiseLevel * baseVoltage * (rand.Float64() - 0.5)

	return signal
}

func generateCurrentSignal(t, voltage, resistance, capacitance, noiseLevel float64) float64 {
	// Calculate current based on RC circuit impedance
	// Z = R + 1/(jωC) for RC circuit

	// For simplicity, use multiple frequency components with phase shifts
	baseFreq := 10.0 // Hz
	omega := 2 * math.Pi * baseFreq

	// RC circuit response at different frequencies
	current := voltage / resistance // DC component

	// AC components with phase shift due to capacitance
	reactance := 1.0 / (omega * capacitance)
	impedanceMag := math.Sqrt(resistance*resistance + reactance*reactance)
	phaseShift := math.Atan2(-reactance, resistance)

	current += 0.3 * voltage * math.Sin(2*math.Pi*10*t+phaseShift) / impedanceMag
	current += 0.1 * voltage * math.Sin(2*math.Pi*25*t+phaseShift*1.2) / (impedanceMag * 0.8)
	current += 0.05 * voltage * math.Sin(2*math.Pi*50*t+phaseShift*1.5) / (impedanceMag * 0.6)

	// Add noise
	current += noiseLevel * current * (rand.Float64() - 0.5)

	return current
}
