package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// ImpedancePoint matches the structure from mockinput
type ImpedancePoint struct {
	Frequency float64 `json:"frequency"`
	Real      float64 `json:"real"`
	Imag      float64 `json:"imag"`
}

// EISMeasurement is an array of impedance points
type EISMeasurement []ImpedancePoint

func main() {
	http.HandleFunc("/eis-data", handleEISData)
	http.HandleFunc("/", handleRoot)

	fmt.Println("Simple EIS data consumer server starting on :8080")
	fmt.Println("Endpoint: http://localhost:8080/eis-data")
	fmt.Println("Press Ctrl+C to stop")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "EIS Data Consumer Server\n")
	fmt.Fprintf(w, "POST to /eis-data to send impedance data\n")
}

func handleEISData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON data
	var measurement EISMeasurement
	if err := json.Unmarshal(body, &measurement); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Print received data
	fmt.Printf("\n=== EIS Measurement Received at %s ===\n", time.Now().Format("15:04:05"))
	fmt.Printf("Number of data points: %d\n", len(measurement))
	
	if len(measurement) > 0 {
		fmt.Printf("Frequency range: %.6g Hz to %.6g Hz\n", 
			measurement[0].Frequency, 
			measurement[len(measurement)-1].Frequency)
		
		// Print first few data points as example
		fmt.Println("Sample data points:")
		limit := 5
		if len(measurement) < limit {
			limit = len(measurement)
		}
		
		for i := 0; i < limit; i++ {
			point := measurement[i]
			fmt.Printf("  [%d] Freq: %.6g Hz, Real: %.6f, Imag: %.6f\n", 
				i, point.Frequency, point.Real, point.Imag)
		}
		
		if len(measurement) > limit {
			fmt.Printf("  ... and %d more data points\n", len(measurement)-limit)
		}
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "success", "received_points": %d}`, len(measurement))
}