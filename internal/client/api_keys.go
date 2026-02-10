package client

import (
	"context"
	"fmt"
	"net/url"
)

// APIKey represents a Braintrust API key
type APIKey struct {
	ID          string `json:"id"`
	OrgID       string `json:"org_id,omitempty"`
	Name        string `json:"name"`
	PreviewName string `json:"preview_name,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	UserEmail   string `json:"user_email,omitempty"`
	Created     string `json:"created,omitempty"`
	Key         string `json:"key,omitempty"` // Only returned on creation
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name    string `json:"name"`
	OrgName string `json:"org_name,omitempty"`
}

// UpdateAPIKeyRequest represents a request to update an API key
type UpdateAPIKeyRequest struct {
	Name string `json:"name,omitempty"`
}

// ListAPIKeysOptions represents options for listing API keys
type ListAPIKeysOptions struct {
	OrgName       string
	StartingAfter string
	EndingBefore  string
	APIKeyName    string
	Limit         int
}

// ListAPIKeysResponse represents a list of API keys
type ListAPIKeysResponse struct {
	APIKeys []APIKey `json:"objects"`
}

// CreateAPIKey creates a new API key
func (c *Client) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKey, error) {
	var apiKey APIKey
	err := c.Do(ctx, "POST", "/v1/api_key", req, &apiKey)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// GetAPIKey retrieves an API key by ID
func (c *Client) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	var apiKey APIKey
	err := c.Do(ctx, "GET", "/v1/api_key/"+url.PathEscape(id), nil, &apiKey)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// UpdateAPIKey updates an existing API key
func (c *Client) UpdateAPIKey(ctx context.Context, id string, req *UpdateAPIKeyRequest) (*APIKey, error) {
	var apiKey APIKey
	err := c.Do(ctx, "PATCH", "/v1/api_key/"+url.PathEscape(id), req, &apiKey)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// DeleteAPIKey deletes an API key
func (c *Client) DeleteAPIKey(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/v1/api_key/"+url.PathEscape(id), nil, nil)
}

// ListAPIKeys lists all API keys
func (c *Client) ListAPIKeys(ctx context.Context, opts *ListAPIKeysOptions) (*ListAPIKeysResponse, error) {
	path := "/v1/api_key"

	// Build query parameters
	if opts != nil {
		separator := "?"

		if opts.OrgName != "" {
			path += separator + "org_name=" + opts.OrgName
			separator = "&"
		}

		if opts.Limit > 0 {
			path += fmt.Sprintf("%slimit=%d", separator, opts.Limit)
			separator = "&"
		}

		if opts.StartingAfter != "" {
			path += separator + "starting_after=" + opts.StartingAfter
			separator = "&"
		}

		if opts.EndingBefore != "" {
			path += separator + "ending_before=" + opts.EndingBefore
			separator = "&"
		}

		if opts.APIKeyName != "" {
			path += separator + "api_key_name=" + opts.APIKeyName
		}
	}

	var result ListAPIKeysResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
