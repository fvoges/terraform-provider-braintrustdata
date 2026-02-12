package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyDatasetID is returned when a dataset ID is empty
var ErrEmptyDatasetID = errors.New("dataset ID cannot be empty")

// datasetPath returns the API path for a specific dataset
func datasetPath(id string) string {
	return "/v1/dataset/" + url.PathEscape(id)
}

// Dataset represents a Braintrust dataset
type Dataset struct {
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

// CreateDatasetRequest represents a request to create a dataset
type CreateDatasetRequest struct {
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Public      *bool                  `json:"public,omitempty"`
	ProjectID   string                 `json:"project_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// UpdateDatasetRequest represents a request to update a dataset
type UpdateDatasetRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Public      *bool                  `json:"public,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// ListDatasetsOptions represents options for listing datasets
type ListDatasetsOptions struct {
	ProjectID string
	Cursor    string
	Limit     int
}

// ListDatasetsResponse represents a list of datasets
type ListDatasetsResponse struct {
	Cursor   string    `json:"cursor,omitempty"`
	Datasets []Dataset `json:"objects"`
}

// CreateDataset creates a new dataset
func (c *Client) CreateDataset(ctx context.Context, req *CreateDatasetRequest) (*Dataset, error) {
	var dataset Dataset
	err := c.Do(ctx, "POST", "/v1/dataset", req, &dataset)
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

// GetDataset retrieves a dataset by ID
func (c *Client) GetDataset(ctx context.Context, id string) (*Dataset, error) {
	if id == "" {
		return nil, ErrEmptyDatasetID
	}
	var dataset Dataset
	err := c.Do(ctx, "GET", datasetPath(id), nil, &dataset)
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

// UpdateDataset updates an existing dataset
func (c *Client) UpdateDataset(ctx context.Context, id string, req *UpdateDatasetRequest) (*Dataset, error) {
	if id == "" {
		return nil, ErrEmptyDatasetID
	}
	var dataset Dataset
	err := c.Do(ctx, "PATCH", datasetPath(id), req, &dataset)
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

// DeleteDataset deletes a dataset
func (c *Client) DeleteDataset(ctx context.Context, id string) error {
	if id == "" {
		return ErrEmptyDatasetID
	}
	return c.Do(ctx, "DELETE", datasetPath(id), nil, nil)
}

// ListDatasets lists all datasets for a project
func (c *Client) ListDatasets(ctx context.Context, opts *ListDatasetsOptions) (*ListDatasetsResponse, error) {
	path := "/v1/dataset"

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

	var result ListDatasetsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
