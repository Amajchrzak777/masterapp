package signal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/adam/masterapp/pkg/config"
)


// CSVDataLoader implements loading signals from CSV files
type CSVDataLoader struct {
	validator Validator
}

// NewDataLoader creates a new CSV data loader
func NewDataLoader() DataLoader {
	return &CSVDataLoader{
		validator: NewValidator(),
	}
}

// LoadSignalFromCSV loads signal data from a CSV file
// Expected CSV format: timestamp,time_offset,value
func (loader *CSVDataLoader) LoadSignalFromCSV(filename string, sampleRate float64) ([]Signal, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, config.NewProcessingError("file opening", fmt.Errorf("failed to open %s: %w", filename, err))
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, config.NewProcessingError("CSV reading", fmt.Errorf("failed to read CSV: %w", err))
	}

	if len(records) < 2 {
		return nil, config.NewValidationError("Data", "CSV file must have at least header and one data row")
	}

	// Skip header row
	records = records[1:]

	// Group data into 1-second chunks (assuming 1000 samples per second)
	samplesPerSecond := int(sampleRate)
	totalSignals := (len(records) + samplesPerSecond - 1) / samplesPerSecond
	signals := make([]Signal, 0, totalSignals)

	for i := 0; i < len(records); i += samplesPerSecond {
		end := i + samplesPerSecond
		if end > len(records) {
			end = len(records)
		}

		chunk := records[i:end]
		signal, err := loader.parseSignalChunk(chunk, sampleRate)
		if err != nil {
			return nil, config.NewProcessingError("signal parsing", err)
		}

		if err := loader.validator.ValidateSignal(signal); err != nil {
			return nil, config.NewProcessingError("signal validation", err)
		}

		signals = append(signals, signal)
	}

	return signals, nil
}

// LoadVoltageAndCurrentFromCSV loads both voltage and current signals from separate CSV files
func (loader *CSVDataLoader) LoadVoltageAndCurrentFromCSV(voltageFile, currentFile string, sampleRate float64) ([]Signal, []Signal, error) {
	voltageSignals, err := loader.LoadSignalFromCSV(voltageFile, sampleRate)
	if err != nil {
		return nil, nil, config.NewProcessingError("voltage loading", err)
	}

	currentSignals, err := loader.LoadSignalFromCSV(currentFile, sampleRate)
	if err != nil {
		return nil, nil, config.NewProcessingError("current loading", err)
	}

	if len(voltageSignals) != len(currentSignals) {
		return nil, nil, config.NewValidationError("DataLength", 
			fmt.Sprintf("voltage and current must have same number of signals: got %d voltage, %d current", 
				len(voltageSignals), len(currentSignals)))
	}

	// Validate that corresponding signals are compatible
	for i, voltageSignal := range voltageSignals {
		if err := ValidateSignalsMatch(voltageSignal, currentSignals[i]); err != nil {
			return nil, nil, config.NewProcessingError(fmt.Sprintf("signal pair %d validation", i), err)
		}
	}

	return voltageSignals, currentSignals, nil
}

// parseSignalChunk converts a chunk of CSV records into a Signal
func (loader *CSVDataLoader) parseSignalChunk(records [][]string, sampleRate float64) (Signal, error) {
	if len(records) == 0 {
		return Signal{}, config.NewValidationError("Records", "empty record chunk")
	}

	values := make([]float64, len(records))
	var timestamp time.Time

	for i, record := range records {
		if len(record) < 3 {
			return Signal{}, config.NewValidationError("Record", fmt.Sprintf("record %d must have at least 3 columns", i))
		}

		// Parse timestamp (first record sets the timestamp for the whole signal)
		if i == 0 {
			parsedTime, err := time.Parse(time.RFC3339Nano, record[0])
			if err != nil {
				return Signal{}, config.NewProcessingError("timestamp parsing", 
					fmt.Errorf("invalid timestamp format in record %d: %w", i, err))
			}
			timestamp = parsedTime
		}

		// Parse value (third column)
		value, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return Signal{}, config.NewProcessingError("value parsing", 
				fmt.Errorf("invalid value in record %d: %w", i, err))
		}

		values[i] = value
	}

	return Signal{
		Timestamp:  timestamp,
		Values:     values,
		SampleRate: sampleRate,
	}, nil
}

// GetDataInfo returns information about the loaded data files
func GetDataInfo(voltageFile, currentFile string) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Check voltage file
	vFile, err := os.Open(voltageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open voltage file: %w", err)
	}
	defer vFile.Close()

	vReader := csv.NewReader(vFile)
	vRecords, err := vReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read voltage CSV: %w", err)
	}

	// Check current file
	cFile, err := os.Open(currentFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open current file: %w", err)
	}
	defer cFile.Close()

	cReader := csv.NewReader(cFile)
	cRecords, err := cReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read current CSV: %w", err)
	}

	info["voltage_samples"] = len(vRecords) - 1 // Exclude header
	info["current_samples"] = len(cRecords) - 1 // Exclude header
	info["voltage_file"] = voltageFile
	info["current_file"] = currentFile

	if len(vRecords) > 1 && len(cRecords) > 1 {
		// Parse first and last timestamps to get duration
		firstTime, _ := time.Parse(time.RFC3339Nano, vRecords[1][0])
		lastTime, _ := time.Parse(time.RFC3339Nano, vRecords[len(vRecords)-1][0])
		duration := lastTime.Sub(firstTime).Seconds()
		
		info["duration_seconds"] = duration
		info["estimated_sample_rate"] = float64(len(vRecords)-1) / duration
	}

	return info, nil
}