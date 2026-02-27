package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrEmptyAISecretID is returned when an AI secret ID is empty.
var ErrEmptyAISecretID = errors.New("ai secret ID cannot be empty")

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
