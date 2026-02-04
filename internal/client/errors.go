package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// APIError represents an error response from the Braintrust API
type APIError struct {
	Details    map[string]interface{}
	Message    string
	StatusCode int
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status %d, message: %s", e.StatusCode, e.Message)
}

// parseAPIError attempts to parse an error response from the API
func parseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}

	// Try to parse as JSON
	var errResp struct {
		Details map[string]interface{} `json:"details"`
		Error   string                 `json:"error"`
		Message string                 `json:"message"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil {
		// Successfully parsed JSON
		if errResp.Error != "" {
			apiErr.Message = errResp.Error
		} else if errResp.Message != "" {
			apiErr.Message = errResp.Message
		}
		if errResp.Details != nil {
			apiErr.Details = errResp.Details
		}
	} else {
		// Not JSON, use raw body as message
		apiErr.Message = string(body)
	}

	// Sanitize sensitive data from error message
	apiErr.Message = sanitizeSensitiveData(apiErr.Message)

	return apiErr
}

// sanitizeSensitiveData removes sensitive information from error messages
func sanitizeSensitiveData(msg string) string {
	// Redact API keys (sk-*)
	apiKeyRegex := regexp.MustCompile(`sk-[a-zA-Z0-9-]+`)
	msg = apiKeyRegex.ReplaceAllString(msg, "[REDACTED]")

	// Redact Bearer tokens
	bearerRegex := regexp.MustCompile(`Bearer\s+sk-[a-zA-Z0-9-]+`)
	msg = bearerRegex.ReplaceAllString(msg, "Bearer [REDACTED]")

	// Redact full Authorization headers
	authHeaderRegex := regexp.MustCompile(`Authorization:\s*Bearer\s+[a-zA-Z0-9-]+`)
	msg = authHeaderRegex.ReplaceAllString(msg, "Authorization: Bearer [REDACTED]")

	return msg
}

// IsNotFound returns true if the error is a 404 Not Found
func IsNotFound(err error) bool {
	apiErr := &APIError{}
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests
func IsRateLimited(err error) bool {
	apiErr := &APIError{}
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 Unauthorized
func IsUnauthorized(err error) bool {
	apiErr := &APIError{}
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}
