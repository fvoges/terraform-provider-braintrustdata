package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrEmptyEnvironmentVariableID is returned when an environment variable ID is empty.
var ErrEmptyEnvironmentVariableID = errors.New("environment variable ID cannot be empty")

// EnvironmentVariable represents a Braintrust environment variable.
type EnvironmentVariable struct {
	ID             string `json:"id"`
	ObjectType     string `json:"object_type,omitempty"`
	ObjectID       string `json:"object_id,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Created        string `json:"created,omitempty"`
	SecretType     string `json:"secret_type,omitempty"`
	SecretCategory string `json:"secret_category,omitempty"`
	Used           bool   `json:"used,omitempty"`
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
