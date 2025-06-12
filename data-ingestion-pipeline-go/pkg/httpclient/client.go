package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anurag/data-ingestion-pipeline-go/pkg/logger"
)

// Client interface defines HTTP client behavior
// Go Concept: Interface for testability and abstraction
type Client interface {
	Get(ctx context.Context, url string, result interface{}) error
	Post(ctx context.Context, url string, body interface{}, result interface{}) error
	Put(ctx context.Context, url string, body interface{}, result interface{}) error
	Delete(ctx context.Context, url string) error
}

// Config holds HTTP client configuration
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
	RateLimit  int // requests per second
}

// httpClient implements Client interface
// Go Concept: Struct implementing interface
type httpClient struct {
	client      *http.Client
	config      Config
	logger      logger.Logger
	rateLimiter chan struct{} // Channel for rate limiting
}

// New creates a new HTTP client
// Go Concept: Constructor function with dependency injection
func New(config Config, log logger.Logger) Client {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Create rate limiter channel
	// Go Concept: Using channel as semaphore for rate limiting
	rateLimiter := make(chan struct{}, config.RateLimit)

	// Fill rate limiter channel
	for i := 0; i < config.RateLimit; i++ {
		rateLimiter <- struct{}{}
	}

	// Start rate limiter refill goroutine
	// Go Concept: Goroutine for background processing
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(config.RateLimit))
		defer ticker.Stop()

		for range ticker.C {
			select {
			case rateLimiter <- struct{}{}:
				// Token added to bucket
			default:
				// Bucket is full, skip
			}
		}
	}()

	return &httpClient{
		client:      client,
		config:      config,
		logger:      log,
		rateLimiter: rateLimiter,
	}
}

// Get performs a GET request
// Go Concept: Method with context for cancellation
func (hc *httpClient) Get(ctx context.Context, url string, result interface{}) error {
	return hc.doRequestWithRetry(ctx, "GET", url, nil, result)
}

// Post performs a POST request
func (hc *httpClient) Post(ctx context.Context, url string, body interface{}, result interface{}) error {
	return hc.doRequestWithRetry(ctx, "POST", url, body, result)
}

// Put performs a PUT request
func (hc *httpClient) Put(ctx context.Context, url string, body interface{}, result interface{}) error {
	return hc.doRequestWithRetry(ctx, "PUT", url, body, result)
}

// Delete performs a DELETE request
func (hc *httpClient) Delete(ctx context.Context, url string) error {
	return hc.doRequestWithRetry(ctx, "DELETE", url, nil, nil)
}

// doRequestWithRetry performs HTTP request with retry logic
// Go Concept: Private method with complex retry logic
func (hc *httpClient) doRequestWithRetry(ctx context.Context, method, url string, body, result interface{}) error {
	var lastErr error

	// Build full URL
	fullURL := hc.buildURL(url)

	// Retry loop
	for attempt := 0; attempt <= hc.config.RetryCount; attempt++ {
		// Rate limiting - wait for token
		// Go Concept: Using select with context for cancellation
		select {
		case <-hc.rateLimiter:
			// Got rate limit token, proceed
		case <-ctx.Done():
			return ctx.Err() // Context cancelled
		}

		// Perform the request
		err := hc.doRequest(ctx, method, fullURL, body, result)
		if err == nil {
			// Success!
			hc.logger.Debug("HTTP request successful",
				logger.String("method", method),
				logger.String("url", fullURL),
				logger.Int("attempt", attempt+1),
			)
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !hc.isRetryableError(err) {
			hc.logger.Error("Non-retryable error, stopping retries",
				logger.String("method", method),
				logger.String("url", fullURL),
				logger.Error(err),
			)
			return err
		}

		// Don't sleep after last attempt
		if attempt < hc.config.RetryCount {
			hc.logger.Warn("Request failed, retrying",
				logger.String("method", method),
				logger.String("url", fullURL),
				logger.Int("attempt", attempt+1),
				logger.Duration("delay", hc.config.RetryDelay),
				logger.Error(err),
			)

			// Wait before retry with context cancellation support
			select {
			case <-time.After(hc.config.RetryDelay):
				// Continue to next retry
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// All retries exhausted
	hc.logger.Error("All retries exhausted",
		logger.String("method", method),
		logger.String("url", fullURL),
		logger.Int("attempts", hc.config.RetryCount+1),
		logger.Error(lastErr),
	)

	return fmt.Errorf("request failed after %d attempts: %w", hc.config.RetryCount+1, lastErr)
}

// doRequest performs a single HTTP request
// Go Concept: Method that handles the actual HTTP call
func (hc *httpClient) doRequest(ctx context.Context, method, url string, body, result interface{}) error {
	// Create request
	var bodyReader io.Reader
	if body != nil {
		// Marshal body to JSON
		// Go Concept: JSON marshaling for request body
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request with context
	// Go Concept: Using context with HTTP requests
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "data-ingestion-pipeline/1.0")

	// Perform request
	start := time.Now()
	resp, err := hc.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return &HTTPError{
			Method:     method,
			URL:        url,
			StatusCode: 0,
			Message:    err.Error(),
			Duration:   duration,
		}
	}
	defer resp.Body.Close() // Always close response body

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			Method:     method,
			URL:        url,
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
			Duration:   duration,
		}
	}

	// Unmarshal response if result is provided
	if result != nil && len(respBody) > 0 {
		// Go Concept: JSON unmarshaling for response
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// buildURL builds full URL from base URL and path
func (hc *httpClient) buildURL(path string) string {
	if hc.config.BaseURL == "" {
		return path
	}

	// Simple URL building - in production, use url.URL for proper handling
	if path[0] != '/' {
		path = "/" + path
	}

	return hc.config.BaseURL + path
}

// isRetryableError determines if an error should trigger a retry
func (hc *httpClient) isRetryableError(err error) bool {
	// Check if it's our custom HTTPError
	// Go Concept: Type assertion to check specific error types
	if httpErr, ok := err.(*HTTPError); ok {
		// Retry on 5xx errors, 429 (rate limit), and 408 (timeout)
		return httpErr.StatusCode >= 500 ||
			httpErr.StatusCode == http.StatusTooManyRequests ||
			httpErr.StatusCode == http.StatusRequestTimeout
	}

	// Retry on network errors
	return true
}

// HTTPError represents an HTTP error
// Go Concept: Custom error type with additional context
type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	Message    string
	Duration   time.Duration
}

// Error implements the error interface
func (he *HTTPError) Error() string {
	if he.StatusCode == 0 {
		return fmt.Sprintf("HTTP %s %s failed: %s (took %v)",
			he.Method, he.URL, he.Message, he.Duration)
	}
	return fmt.Sprintf("HTTP %s %s returned %d: %s (took %v)",
		he.Method, he.URL, he.StatusCode, he.Message, he.Duration)
}

// IsTemporary indicates if this error might be retryable
func (he *HTTPError) IsTemporary() bool {
	return he.StatusCode >= 500 ||
		he.StatusCode == http.StatusTooManyRequests ||
		he.StatusCode == http.StatusRequestTimeout
}
