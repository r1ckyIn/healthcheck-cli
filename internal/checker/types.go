// Data type definitions
// Defines core data structures for health check
package checker

import (
	"time"
)

// Version is the application version, set by cmd package at init
var Version = "dev"

// Endpoint represents an endpoint to check
type Endpoint struct {
	Name            string            // Endpoint name for display
	URL             string            // URL to check
	Timeout         time.Duration     // Request timeout
	Retries         int               // Retry count on failure
	ExpectedStatus  int               // Expected HTTP status code
	FollowRedirects bool              // Whether to follow redirects
	Insecure        bool              // Whether to skip SSL verification
	Headers         map[string]string // Custom request headers
}

// Result represents health check result
type Result struct {
	Name       string        // Endpoint name
	URL        string        // Checked URL
	Healthy    bool          // Whether healthy
	StatusCode *int          // HTTP status code (nil if connection failed)
	Latency    time.Duration // Response latency
	Error      error         // Error message
}

// Summary represents batch check summary
type Summary struct {
	Total     int           // Total endpoints
	Healthy   int           // Healthy count
	Unhealthy int           // Unhealthy count
	Duration  time.Duration // Total duration
}

// BatchResult represents complete batch check result
type BatchResult struct {
	Timestamp time.Time // Check start time
	Summary   Summary   // Summary info
	Results   []Result  // Detailed results
}

// DefaultEndpoint creates an endpoint with default config
func DefaultEndpoint(url string) Endpoint {
	return Endpoint{
		Name:            url,
		URL:             url,
		Timeout:         5 * time.Second,
		Retries:         0,
		ExpectedStatus:  200,
		FollowRedirects: true,
		Insecure:        false,
		Headers:         make(map[string]string),
	}
}
