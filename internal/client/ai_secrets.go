package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ErrEmptyAISecretID is returned when an AI secret ID is empty.
var ErrEmptyAISecretID = errors.New("ai secret ID cannot be empty")

// ErrNilCreateAISecretRequest is returned when the create request is nil.
var ErrNilCreateAISecretRequest = errors.New("create AI secret request cannot be nil")

// ErrNilUpdateAISecretRequest is returned when the update request is nil.
var ErrNilUpdateAISecretRequest = errors.New("update AI secret request cannot be nil")

// AISecret represents a Braintrust AI secret.
type AISecret struct {
	ID            string                 `json:"id"`
	Created       string                 `json:"created,omitempty"`
	UpdatedAt     string                 `json:"updated_at,omitempty"`
	OrgID         string                 `json:"org_id,omitempty"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	PreviewSecret string                 `json:"preview_secret,omitempty"`
}

// CreateAISecretRequest represents a request to create an AI secret.
type CreateAISecretRequest struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Secret   string                 `json:"secret"`
	OrgName  string                 `json:"org_name,omitempty"`
}

// UpdateAISecretRequest represents a request to update an AI secret.
type UpdateAISecretRequest struct {
	Name     *string                 `json:"name,omitempty"`
	Type     *string                 `json:"type,omitempty"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
	Secret   *string                 `json:"secret,omitempty"`
}

// ListAISecretsOptions represents options for listing AI secrets.
type ListAISecretsOptions struct {
	OrgName       string
	StartingAfter string
	EndingBefore  string
	AISecretName  string
	AISecretTypes []string
	IDs           []string
	Limit         int
}

// ListAISecretsResponse represents a list of AI secrets.
type ListAISecretsResponse struct {
	AISecrets []AISecret `json:"objects"`
}

func aiSecretPath(id string) string {
	return "/v1/ai_secret/" + url.PathEscape(id)
}

// CreateAISecret creates a new AI secret.
func (c *Client) CreateAISecret(ctx context.Context, req *CreateAISecretRequest) (*AISecret, error) {
	if req == nil {
		return nil, ErrNilCreateAISecretRequest
	}

	var aiSecret AISecret
	err := c.Do(ctx, http.MethodPost, "/v1/ai_secret", req, &aiSecret)
	if err != nil {
		return nil, err
	}

	return &aiSecret, nil
}

// UpdateAISecret updates an AI secret by ID.
func (c *Client) UpdateAISecret(ctx context.Context, id string, req *UpdateAISecretRequest) (*AISecret, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyAISecretID
	}
	if req == nil {
		return nil, ErrNilUpdateAISecretRequest
	}

	var aiSecret AISecret
	err := c.Do(ctx, http.MethodPatch, aiSecretPath(id), req, &aiSecret)
	if err != nil {
		return nil, err
	}

	return &aiSecret, nil
}

// DeleteAISecret deletes an AI secret by ID.
func (c *Client) DeleteAISecret(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrEmptyAISecretID
	}

	return c.Do(ctx, http.MethodDelete, aiSecretPath(id), nil, nil)
}

// GetAISecret retrieves an AI secret by ID.
func (c *Client) GetAISecret(ctx context.Context, id string) (*AISecret, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyAISecretID
	}

	var aiSecret AISecret
	err := c.Do(ctx, "GET", aiSecretPath(id), nil, &aiSecret)
	if err != nil {
		return nil, err
	}

	return &aiSecret, nil
}

// ListAISecrets lists AI secrets using API-native filters.
func (c *Client) ListAISecrets(ctx context.Context, opts *ListAISecretsOptions) (*ListAISecretsResponse, error) {
	path := "/v1/ai_secret"

	if opts != nil {
		params := url.Values{}

		if opts.OrgName != "" {
			params.Set("org_name", opts.OrgName)
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
		if opts.AISecretName != "" {
			params.Set("ai_secret_name", opts.AISecretName)
		}
		for _, aiSecretType := range opts.AISecretTypes {
			if aiSecretType != "" {
				params.Add("ai_secret_type", aiSecretType)
			}
		}
		for _, id := range opts.IDs {
			if id != "" {
				params.Add("ids", id)
			}
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListAISecretsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
