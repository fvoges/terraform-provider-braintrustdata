package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyViewID is returned when a view ID is empty.
var ErrEmptyViewID = errors.New("view ID cannot be empty")

// ViewType represents the table type associated with a view.
type ViewType string

// View type constants.
const (
	ViewTypeProjects    ViewType = "projects"
	ViewTypeExperiments ViewType = "experiments"
	ViewTypeExperiment  ViewType = "experiment"
	ViewTypePlaygrounds ViewType = "playgrounds"
	ViewTypePlayground  ViewType = "playground"
	ViewTypeDatasets    ViewType = "datasets"
	ViewTypeDataset     ViewType = "dataset"
	ViewTypePrompts     ViewType = "prompts"
	ViewTypeTools       ViewType = "tools"
	ViewTypeScorers     ViewType = "scorers"
	ViewTypeLogs        ViewType = "logs"
)

// View represents a Braintrust view.
type View struct {
	Options    map[string]interface{} `json:"options,omitempty"`
	ViewData   map[string]interface{} `json:"view_data,omitempty"`
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	ObjectID   string                 `json:"object_id"`
	ObjectType ACLObjectType          `json:"object_type"`
	ViewType   ViewType               `json:"view_type"`
	Created    string                 `json:"created,omitempty"`
	DeletedAt  string                 `json:"deleted_at,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
}

// CreateViewRequest represents a request to create a view.
type CreateViewRequest struct {
	Options    map[string]interface{} `json:"options,omitempty"`
	ViewData   map[string]interface{} `json:"view_data,omitempty"`
	ObjectID   string                 `json:"object_id"`
	Name       string                 `json:"name,omitempty"`
	ObjectType ACLObjectType          `json:"object_type"`
	ViewType   ViewType               `json:"view_type"`
}

// UpdateViewRequest represents a request to update a view.
type UpdateViewRequest struct {
	Options    map[string]interface{} `json:"options,omitempty"`
	ViewData   map[string]interface{} `json:"view_data,omitempty"`
	ObjectID   string                 `json:"object_id"`
	Name       *string                `json:"name,omitempty"`
	ObjectType ACLObjectType          `json:"object_type"`
}

// DeleteViewRequest represents a request to delete a view.
type DeleteViewRequest struct {
	ObjectID   string        `json:"object_id"`
	ObjectType ACLObjectType `json:"object_type"`
}

// GetViewOptions represents query options for retrieving a view by ID.
type GetViewOptions struct {
	ObjectID   string
	ObjectType ACLObjectType
}

// ListViewsOptions represents options for listing views.
type ListViewsOptions struct {
	ObjectID      string
	ObjectType    ACLObjectType
	ViewName      string
	ViewType      ViewType
	StartingAfter string
	EndingBefore  string
	IDs           []string
	Limit         int
}

// ListViewsResponse represents a list of views.
type ListViewsResponse struct {
	Objects []View `json:"objects"`
}

func viewPath(id string) string {
	return "/v1/view/" + url.PathEscape(id)
}

// CreateView creates a new view.
func (c *Client) CreateView(ctx context.Context, req *CreateViewRequest) (*View, error) {
	var view View
	err := c.Do(ctx, "POST", "/v1/view", req, &view)
	if err != nil {
		return nil, err
	}

	return &view, nil
}

// GetView retrieves a view by ID.
func (c *Client) GetView(ctx context.Context, id string, opts *GetViewOptions) (*View, error) {
	if id == "" {
		return nil, ErrEmptyViewID
	}

	path := viewPath(id)
	if opts != nil {
		params := url.Values{}
		if opts.ObjectID != "" {
			params.Set("object_id", opts.ObjectID)
		}
		if opts.ObjectType != "" {
			params.Set("object_type", string(opts.ObjectType))
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var view View
	err := c.Do(ctx, "GET", path, nil, &view)
	if err != nil {
		return nil, err
	}

	return &view, nil
}

// UpdateView updates an existing view.
func (c *Client) UpdateView(ctx context.Context, id string, req *UpdateViewRequest) (*View, error) {
	if id == "" {
		return nil, ErrEmptyViewID
	}

	var view View
	err := c.Do(ctx, "PATCH", viewPath(id), req, &view)
	if err != nil {
		return nil, err
	}

	return &view, nil
}

// DeleteView deletes a view.
func (c *Client) DeleteView(ctx context.Context, id string, req *DeleteViewRequest) error {
	if id == "" {
		return ErrEmptyViewID
	}

	return c.Do(ctx, "DELETE", viewPath(id), req, nil)
}

// ListViews lists views using API-native filters.
func (c *Client) ListViews(ctx context.Context, opts *ListViewsOptions) (*ListViewsResponse, error) {
	path := "/v1/view"

	if opts != nil {
		params := url.Values{}
		if opts.ObjectID != "" {
			params.Set("object_id", opts.ObjectID)
		}
		if opts.ObjectType != "" {
			params.Set("object_type", string(opts.ObjectType))
		}
		if opts.ViewName != "" {
			params.Set("view_name", opts.ViewName)
		}
		if opts.ViewType != "" {
			params.Set("view_type", string(opts.ViewType))
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
		for _, id := range opts.IDs {
			if id != "" {
				params.Add("ids", id)
			}
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListViewsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
