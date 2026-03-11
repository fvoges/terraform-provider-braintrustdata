package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ErrEmptyEnvironmentVariableID is returned when an environment variable ID is empty.
var ErrEmptyEnvironmentVariableID = errors.New("environment variable ID cannot be empty")

// EnvironmentVariable represents a Braintrust environment variable.
type EnvironmentVariable struct {
	ID             string                  `json:"id"`
	ObjectType     string                  `json:"object_type,omitempty"`
	ObjectID       string                  `json:"object_id,omitempty"`
	Name           string                  `json:"name,omitempty"`
	Value          string                  `json:"value,omitempty"`
	Description    string                  `json:"description,omitempty"`
	Metadata       map[string]interface{}  `json:"metadata,omitempty"`
	Created        string                  `json:"created,omitempty"`
	SecretType     string                  `json:"secret_type,omitempty"`
	SecretCategory string                  `json:"secret_category,omitempty"`
	Used           EnvironmentVariableUsed `json:"used,omitempty"`
}

// EnvironmentVariableUsed decodes Braintrust's inconsistent `used` response field.
// The API may return a boolean, a timestamp string, or null. Non-empty strings
// indicate the variable has been used.
type EnvironmentVariableUsed bool

// UnmarshalJSON accepts `used` as bool, string, or null.
func (u *EnvironmentVariableUsed) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	switch trimmed {
	case "", "null":
		*u = false
		return nil
	case "true", "false":
		parsed, err := strconv.ParseBool(trimmed)
		if err != nil {
			return err
		}
		*u = EnvironmentVariableUsed(parsed)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		stringValue = strings.TrimSpace(stringValue)
		if stringValue == "" {
			*u = false
			return nil
		}

		if parsed, err := strconv.ParseBool(stringValue); err == nil {
			*u = EnvironmentVariableUsed(parsed)
			return nil
		}

		*u = true
		return nil
	}

	return fmt.Errorf("unsupported used value: %s", trimmed)
}

// CreateEnvironmentVariableRequest represents a request to create an environment variable.
type CreateEnvironmentVariableRequest struct {
	ObjectType     string                 `json:"object_type"`
	ObjectID       string                 `json:"object_id"`
	Name           string                 `json:"name"`
	Value          string                 `json:"value"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	SecretType     string                 `json:"secret_type,omitempty"`
	SecretCategory string                 `json:"secret_category,omitempty"`
}

// UpdateEnvironmentVariableRequest represents a request to update an environment variable.
type UpdateEnvironmentVariableRequest struct {
	Name           *string                 `json:"name,omitempty"`
	Value          *string                 `json:"value,omitempty"`
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
	SecretType     *string                 `json:"secret_type,omitempty"`
	SecretCategory *string                 `json:"secret_category,omitempty"`
}

// ListEnvironmentVariablesOptions represents options for listing environment variables.
type ListEnvironmentVariablesOptions struct {
	ObjectType    string
	ObjectID      string
	StartingAfter string
	EndingBefore  string
	Limit         int
}

// ListEnvironmentVariablesResponse represents a list of environment variables.
type ListEnvironmentVariablesResponse struct {
	EnvironmentVariables []EnvironmentVariable `json:"objects"`
}

func environmentVariablePath(id string) string {
	return "/v1/env_var/" + url.PathEscape(id)
}

// CreateEnvironmentVariable creates a new environment variable.
func (c *Client) CreateEnvironmentVariable(ctx context.Context, req *CreateEnvironmentVariableRequest) (*EnvironmentVariable, error) {
	var environmentVariable EnvironmentVariable
	if err := c.Do(ctx, "POST", "/v1/env_var", req, &environmentVariable); err != nil {
		return nil, err
	}

	return &environmentVariable, nil
}

// UpdateEnvironmentVariable updates an existing environment variable.
func (c *Client) UpdateEnvironmentVariable(ctx context.Context, id string, req *UpdateEnvironmentVariableRequest) (*EnvironmentVariable, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyEnvironmentVariableID
	}

	var environmentVariable EnvironmentVariable
	if err := c.Do(ctx, "PATCH", environmentVariablePath(id), req, &environmentVariable); err != nil {
		return nil, err
	}

	return &environmentVariable, nil
}

// DeleteEnvironmentVariable deletes an environment variable by ID.
func (c *Client) DeleteEnvironmentVariable(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrEmptyEnvironmentVariableID
	}

	return c.Do(ctx, "DELETE", environmentVariablePath(id), nil, nil)
}

// GetEnvironmentVariable retrieves an environment variable by ID.
func (c *Client) GetEnvironmentVariable(ctx context.Context, id string) (*EnvironmentVariable, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyEnvironmentVariableID
	}

	var environmentVariable EnvironmentVariable
	if err := c.Do(ctx, "GET", environmentVariablePath(id), nil, &environmentVariable); err != nil {
		return nil, err
	}

	return &environmentVariable, nil
}

// ListEnvironmentVariables lists environment variables using API-native filters.
func (c *Client) ListEnvironmentVariables(ctx context.Context, opts *ListEnvironmentVariablesOptions) (*ListEnvironmentVariablesResponse, error) {
	path := "/v1/env_var"

	if opts != nil {
		params := url.Values{}

		if opts.ObjectType != "" {
			params.Set("object_type", opts.ObjectType)
		}
		if opts.ObjectID != "" {
			params.Set("object_id", opts.ObjectID)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.StartingAfter != "" {
			params.Set("starting_after", opts.StartingAfter)
		}
		if opts.EndingBefore != "" {
			params.Set("ending_before", opts.EndingBefore)
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListEnvironmentVariablesResponse
	if err := c.Do(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
