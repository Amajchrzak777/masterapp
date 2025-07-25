package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type DefaultDataSender struct {
	targetURL string
	client    *http.Client
	healthy   bool
}

func NewDataSender(targetURL string) DataSender {
	// Validate URL
	if _, err := url.Parse(targetURL); err != nil {
		log.Printf("Warning: Invalid target URL %s: %v", targetURL, err)
	}

	return &DefaultDataSender{
		targetURL: targetURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		healthy: true,
	}
}

func (ds *DefaultDataSender) SendEISMeasurement(measurement EISMeasurement) error {
	if ds.targetURL == "" {
		return NewNetworkError(ds.targetURL, 0, ErrInvalidURL)
	}

	jsonData, err := json.Marshal(measurement)
	if err != nil {
		ds.healthy = false
		return NewProcessingError("JSON marshaling", ErrJSONMarshalFailed)
	}

	req, err := http.NewRequest("POST", ds.targetURL, bytes.NewBuffer(jsonData))
	if err != nil {
		ds.healthy = false
		return NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Data-Type", "EIS-Measurement")

	resp, err := ds.client.Do(req)
	if err != nil {
		ds.healthy = false
		return NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to send request: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		ds.healthy = false
		return NewNetworkError(ds.targetURL, resp.StatusCode, ErrInvalidHTTPResponse)
	}

	ds.healthy = true
	log.Printf("Successfully sent EIS measurement data at %v", measurement.Impedance.Timestamp.Format("15:04:05"))
	return nil
}

func (ds *DefaultDataSender) SendImpedanceData(impedanceData ImpedanceData) error {
	if ds.targetURL == "" {
		return NewNetworkError(ds.targetURL, 0, ErrInvalidURL)
	}

	jsonData, err := json.Marshal(impedanceData)
	if err != nil {
		ds.healthy = false
		return NewProcessingError("JSON marshaling", ErrJSONMarshalFailed)
	}

	req, err := http.NewRequest("POST", ds.targetURL, bytes.NewBuffer(jsonData))
	if err != nil {
		ds.healthy = false
		return NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Data-Type", "Impedance-Data")

	resp, err := ds.client.Do(req)
	if err != nil {
		ds.healthy = false
		return NewNetworkError(ds.targetURL, 0, fmt.Errorf("failed to send request: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		ds.healthy = false
		return NewNetworkError(ds.targetURL, resp.StatusCode, ErrInvalidHTTPResponse)
	}

	ds.healthy = true
	log.Printf("Successfully sent impedance data at %v", impedanceData.Timestamp.Format("15:04:05"))
	return nil
}

func (ds *DefaultDataSender) FormatAsJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", NewProcessingError("JSON formatting", ErrJSONMarshalFailed)
	}
	return string(jsonData), nil
}

func (ds *DefaultDataSender) IsHealthy() bool {
	return ds.healthy
}