package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/adam/masterapp/pkg/config"
	"github.com/adam/masterapp/pkg/signal"
)

// DefaultSender implements HTTP-based data transmission
type DefaultSender struct {
	targetURL string
	client    *http.Client
	healthy   bool
}

// NewSender creates a new network data sender
func NewSender(targetURL string) Sender {
	// Validate URL
	if _, err := url.Parse(targetURL); err != nil {
		log.Printf("Warning: Invalid target URL %s: %v", targetURL, err)
	}

	return &DefaultSender{
		targetURL: targetURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		healthy: true,
	}
}

// SendEISMeasurement sends a complete EIS measurement to the target server
func (ds *DefaultSender) SendEISMeasurement(measurement signal.EISMeasurement) error {
	if ds.targetURL == "" {
		return config.NewNetworkError(ds.targetURL, 0, config.ErrInvalidURL)
	}

	jsonData, err := json.Marshal(measurement)
	if err != nil {
		ds.healthy = false
		return config.NewProcessingError("JSON marshaling", config.ErrJSONMarshalFailed)
	}

	req, err := http.NewRequest("POST", ds.targetURL, bytes.NewBuffer(jsonData))
	if err != nil {
		ds.healthy = false
		return config.NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Data-Type", "EIS-Measurement")

	resp, err := ds.client.Do(req)
	if err != nil {
		ds.healthy = false
		return config.NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to send request: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		ds.healthy = false
		return config.NewNetworkError(ds.targetURL, resp.StatusCode, config.ErrInvalidHTTPResponse)
	}

	ds.healthy = true
	log.Printf("Successfully sent EIS measurement data")
	return nil
}

// SendBatchImpedanceData sends a batch of impedance data to the target server
func (ds *DefaultSender) SendBatchImpedanceData(batch []signal.ImpedanceDataWithIteration) error {
	if ds.targetURL == "" {
		return config.NewNetworkError(ds.targetURL, 0, config.ErrInvalidURL)
	}

	// Create batch with unique ID
	batchData := signal.ImpedanceBatch{
		BatchID:   fmt.Sprintf("batch_%d_%d", time.Now().Unix(), len(batch)),
		Timestamp: time.Now(),
		Spectra:   batch,
	}

	jsonData, err := json.Marshal(batchData)
	if err != nil {
		ds.healthy = false
		return config.NewProcessingError("JSON marshaling", config.ErrJSONMarshalFailed)
	}

	// Use batch endpoint
	batchURL := ds.targetURL + "/batch"
	req, err := http.NewRequest("POST", batchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		ds.healthy = false
		return config.NewNetworkError(batchURL, 0, fmt.Errorf("failed to create batch request: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Data-Type", "Impedance-Batch")

	resp, err := ds.client.Do(req)
	if err != nil {
		ds.healthy = false
		return config.NewNetworkError(batchURL, 0, fmt.Errorf("failed to send batch request: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		ds.healthy = false
		return config.NewNetworkError(batchURL, resp.StatusCode, config.ErrInvalidHTTPResponse)
	}

	ds.healthy = true
	log.Printf("Successfully sent batch of %d spectra", len(batch))
	return nil
}

// SendImpedanceData sends impedance data to the target server
func (ds *DefaultSender) SendImpedanceData(impedanceData signal.ImpedanceData) error {
	if ds.targetURL == "" {
		return config.NewNetworkError(ds.targetURL, 0, config.ErrInvalidURL)
	}

	jsonData, err := json.Marshal(impedanceData)
	if err != nil {
		ds.healthy = false
		return config.NewProcessingError("JSON marshaling", config.ErrJSONMarshalFailed)
	}

	req, err := http.NewRequest("POST", ds.targetURL, bytes.NewBuffer(jsonData))
	if err != nil {
		ds.healthy = false
		return config.NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Data-Type", "Impedance-Data")

	resp, err := ds.client.Do(req)
	if err != nil {
		ds.healthy = false
		return config.NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to send request: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		ds.healthy = false
		return config.NewNetworkError(ds.targetURL, resp.StatusCode, config.ErrInvalidHTTPResponse)
	}

	ds.healthy = true
	log.Printf("Successfully sent impedance data at %v", impedanceData.Timestamp.Format("15:04:05"))
	return nil
}

// FormatAsJSON formats data as pretty-printed JSON
func (ds *DefaultSender) FormatAsJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", config.NewProcessingError("JSON formatting", config.ErrJSONMarshalFailed)
	}
	return string(jsonData), nil
}

// IsHealthy returns the current health status of the sender
func (ds *DefaultSender) IsHealthy() bool {
	return ds.healthy
}