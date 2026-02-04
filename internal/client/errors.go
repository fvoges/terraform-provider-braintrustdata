package client

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// APIError represents an error response from the Braintrust API
type APIError struct {
	StatusCode int
	Message    string
	Details    map[string]interface{}
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
		Error   string                 `json:"error"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details"`
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
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests
func IsRateLimited(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 429
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 Unauthorized
func IsUnauthorized(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 401
	}
	return false
}
