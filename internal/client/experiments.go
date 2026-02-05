package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyExperimentID is returned when an experiment ID is empty
var ErrEmptyExperimentID = errors.New("experiment ID cannot be empty")

// experimentPath returns the API path for a specific experiment
func experimentPath(id string) string {
	return "/v1/experiment/" + url.PathEscape(id)
}

// Experiment represents a Braintrust experiment
type Experiment struct {
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Created     string                 `json:"created,omitempty"`
	DeletedAt   string                 `json:"deleted_at,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	OrgID       string                 `json:"org_id,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Public      bool                   `json:"public"`
}

// CreateExperimentRequest represents a request to create an experiment
type CreateExperimentRequest struct {
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Public      *bool                  `json:"public,omitempty"`
	ProjectID   string                 `json:"project_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// UpdateExperimentRequest represents a request to update an experiment
type UpdateExperimentRequest struct {
	Name        string                  `json:"name,omitempty"`
	Description string                  `json:"description,omitempty"`
	Public      *bool                   `json:"public,omitempty"`
	Metadata    *map[string]interface{} `json:"metadata"`
	Tags        []string                `json:"tags,omitempty"`
}

// ListExperimentsOptions represents options for listing experiments
type ListExperimentsOptions struct {
	ProjectID string
	Cursor    string
	Limit     int
}

// ListExperimentsResponse represents a list of experiments
type ListExperimentsResponse struct {
	Cursor      string       `json:"cursor,omitempty"`
	Experiments []Experiment `json:"objects"`
}

// CreateExperiment creates a new experiment
func (c *Client) CreateExperiment(ctx context.Context, req *CreateExperimentRequest) (*Experiment, error) {
	var experiment Experiment
	err := c.Do(ctx, "POST", "/v1/experiment", req, &experiment)
	if err != nil {
		return nil, err
	}
	return &experiment, nil
}

// GetExperiment retrieves an experiment by ID
func (c *Client) GetExperiment(ctx context.Context, id string) (*Experiment, error) {
	if id == "" {
		return nil, ErrEmptyExperimentID
	}
	var experiment Experiment
	err := c.Do(ctx, "GET", experimentPath(id), nil, &experiment)
	if err != nil {
		return nil, err
	}
	return &experiment, nil
}

// UpdateExperiment updates an existing experiment
func (c *Client) UpdateExperiment(ctx context.Context, id string, req *UpdateExperimentRequest) (*Experiment, error) {
	if id == "" {
		return nil, ErrEmptyExperimentID
	}
	var experiment Experiment
	err := c.Do(ctx, "PATCH", experimentPath(id), req, &experiment)
	if err != nil {
		return nil, err
	}
	return &experiment, nil
}

// DeleteExperiment deletes an experiment
func (c *Client) DeleteExperiment(ctx context.Context, id string) error {
	if id == "" {
		return ErrEmptyExperimentID
	}
	return c.Do(ctx, "DELETE", experimentPath(id), nil, nil)
}

// ListExperiments lists all experiments for a project
func (c *Client) ListExperiments(ctx context.Context, opts *ListExperimentsOptions) (*ListExperimentsResponse, error) {
	path := "/v1/experiment"

	// Build query parameters
	if opts != nil {
		params := url.Values{}
		if opts.ProjectID != "" {
			params.Set("project_id", opts.ProjectID)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Cursor != "" {
			params.Set("cursor", opts.Cursor)
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListExperimentsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
