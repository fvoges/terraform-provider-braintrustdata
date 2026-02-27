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

// GetView retrieves a view by ID.
func (c *Client) GetView(ctx context.Context, id string, opts *GetViewOptions) (*View, error) {
	if id == "" {
		return nil, ErrEmptyViewID
	}

	path := "/v1/view/" + url.PathEscape(id)
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
