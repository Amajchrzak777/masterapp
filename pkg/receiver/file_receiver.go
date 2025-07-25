package receiver

import (
	"context"
	"log"
	"time"

	"github.com/adam/masterapp/pkg/config"
	"github.com/adam/masterapp/pkg/signal"
)

// FileReceiver implements data reception from CSV files
type FileReceiver struct {
	voltageChannel   chan signal.Signal
	currentChannel   chan signal.Signal
	voltageFile      string
	currentFile      string
	sampleRate       float64
	validator        signal.Validator
	loader           signal.DataLoader
	running          bool
	voltageSignals   []signal.Signal
	currentSignals   []signal.Signal
	currentIndex     int
}

// NewFileReceiver creates a new file-based data receiver
func NewFileReceiver(voltageFile, currentFile string, sampleRate float64) (DataReceiver, error) {
	loader := signal.NewDataLoader()
	validator := signal.NewValidator()

	// Pre-load all signals from files
	voltageSignals, currentSignals, err := loader.LoadVoltageAndCurrentFromCSV(voltageFile, currentFile, sampleRate)
	if err != nil {
		return nil, config.NewProcessingError("data loading", err)
	}

	log.Printf("Loaded %d signal pairs from files", len(voltageSignals))
	
	// Get data info for logging
	info, err := signal.GetDataInfo(voltageFile, currentFile)
	if err == nil {
		log.Printf("Data info: %+v", info)
	}

	return &FileReceiver{
		voltageChannel: make(chan signal.Signal, 10),
		currentChannel: make(chan signal.Signal, 10),
		voltageFile:    voltageFile,
		currentFile:    currentFile,
		sampleRate:     sampleRate,
		validator:      validator,
		loader:         loader,
		running:        false,
		voltageSignals: voltageSignals,
		currentSignals: currentSignals,
		currentIndex:   0,
	}, nil
}

// StartReceiving begins file-based data reception at 1-second intervals
func (fr *FileReceiver) StartReceiving(ctx context.Context) error {
	if len(fr.voltageSignals) == 0 {
		return config.NewValidationError("Data", "no signals loaded from files")
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fr.running = true
	log.Printf("Starting file-based data reception from %s and %s", fr.voltageFile, fr.currentFile)
	log.Printf("Will process %d signal pairs over %d seconds", len(fr.voltageSignals), len(fr.voltageSignals))

	for fr.running && fr.currentIndex < len(fr.voltageSignals) {
		select {
		case <-ctx.Done():
			fr.running = false
			return ctx.Err()
		case <-ticker.C:
			if fr.currentIndex >= len(fr.voltageSignals) {
				log.Println("All data processed, stopping receiver")
				fr.running = false
				return nil
			}

			voltageSignal := fr.voltageSignals[fr.currentIndex]
			currentSignal := fr.currentSignals[fr.currentIndex]

			// Validate signals before sending
			if err := fr.validator.ValidateSignal(voltageSignal); err != nil {
				log.Printf("Invalid voltage signal at index %d: %v", fr.currentIndex, err)
				fr.currentIndex++
				continue
			}

			if err := fr.validator.ValidateSignal(currentSignal); err != nil {
				log.Printf("Invalid current signal at index %d: %v", fr.currentIndex, err)
				fr.currentIndex++
				continue
			}

			// Send signals to channels
			select {
			case fr.voltageChannel <- voltageSignal:
			default:
				log.Println("Warning: Voltage channel buffer full, dropping sample")
			}

			select {
			case fr.currentChannel <- currentSignal:
			default:
				log.Println("Warning: Current channel buffer full, dropping sample")
			}

			log.Printf("Sent signal pair %d/%d (%.1f%% complete) - Time: %v", 
				fr.currentIndex+1, len(fr.voltageSignals), 
				float64(fr.currentIndex+1)/float64(len(fr.voltageSignals))*100,
				voltageSignal.Timestamp.Format("15:04:05"))

			fr.currentIndex++
		}
	}

	if fr.currentIndex >= len(fr.voltageSignals) {
		log.Println("✅ All file data has been processed successfully")
	}

	return nil
}

// GetVoltageChannel returns the channel for voltage signals
func (fr *FileReceiver) GetVoltageChannel() <-chan signal.Signal {
	return fr.voltageChannel
}

// GetCurrentChannel returns the channel for current signals
func (fr *FileReceiver) GetCurrentChannel() <-chan signal.Signal {
	return fr.currentChannel
}

// Stop gracefully stops the receiver and closes channels
func (fr *FileReceiver) Stop() error {
	fr.running = false
	close(fr.voltageChannel)
	close(fr.currentChannel)
	log.Printf("File receiver stopped after processing %d/%d signals", fr.currentIndex, len(fr.voltageSignals))
	return nil
}

// GetProgress returns the current progress of file processing
func (fr *FileReceiver) GetProgress() (current, total int, percentage float64) {
	total = len(fr.voltageSignals)
	current = fr.currentIndex
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}
	return current, total, percentage
}

// GetRemainingTime estimates remaining processing time
func (fr *FileReceiver) GetRemainingTime() time.Duration {
	remaining := len(fr.voltageSignals) - fr.currentIndex
	if remaining <= 0 {
		return 0
	}
	return time.Duration(remaining) * time.Second
}