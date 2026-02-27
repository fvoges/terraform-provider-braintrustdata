package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrEmptyPromptID is returned when a prompt ID is empty.
var ErrEmptyPromptID = errors.New("prompt ID cannot be empty")

// Prompt represents a Braintrust prompt.
type Prompt struct {
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	PromptData   interface{}            `json:"prompt_data,omitempty"`
	FunctionType string                 `json:"function_type,omitempty"`
	ID           string                 `json:"id"`
	ProjectID    string                 `json:"project_id"`
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Version      string                 `json:"version,omitempty"`
	Created      string                 `json:"created,omitempty"`
	DeletedAt    string                 `json:"deleted_at,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	OrgID        string                 `json:"org_id,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
}

// ListPromptsOptions represents options for listing prompts.
type ListPromptsOptions struct {
	ProjectID     string
	PromptName    string
	Slug          string
	Version       string
	StartingAfter string
	EndingBefore  string
	Limit         int
}

// ListPromptsResponse represents a list of prompts.
type ListPromptsResponse struct {
	Prompts []Prompt `json:"objects"`
}

func promptPath(id string) string {
	return "/v1/prompt/" + url.PathEscape(id)
}

// GetPrompt retrieves a prompt by ID.
func (c *Client) GetPrompt(ctx context.Context, id string) (*Prompt, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyPromptID
	}

	var prompt Prompt
	err := c.Do(ctx, "GET", promptPath(id), nil, &prompt)
	if err != nil {
		return nil, err
	}

	return &prompt, nil
}

// ListPrompts lists prompts using API-native filters.
func (c *Client) ListPrompts(ctx context.Context, opts *ListPromptsOptions) (*ListPromptsResponse, error) {
	path := "/v1/prompt"

	if opts != nil {
		params := url.Values{}
		if opts.ProjectID != "" {
			params.Set("project_id", opts.ProjectID)
		}
		if opts.PromptName != "" {
			params.Set("prompt_name", opts.PromptName)
		}
		if opts.Slug != "" {
			params.Set("slug", opts.Slug)
		}
		if opts.Version != "" {
			params.Set("version", opts.Version)
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

	var result ListPromptsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
