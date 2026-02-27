package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyTagID is returned when a tag ID is empty.
var ErrEmptyTagID = errors.New("tag ID cannot be empty")

// Tag represents a Braintrust project tag.
type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProjectID   string `json:"project_id"`
	UserID      string `json:"user_id,omitempty"`
	Color       string `json:"color,omitempty"`
	Created     string `json:"created,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListTagsOptions represents options for listing tags.
type ListTagsOptions struct {
	StartingAfter string
	EndingBefore  string
	OrgName       string
	ProjectID     string
	ProjectName   string
	TagName       string
	IDs           []string
	Limit         int
}

// ListTagsResponse represents a list of tags.
type ListTagsResponse struct {
	Tags []Tag `json:"objects"`
}

// GetTag retrieves a tag by ID.
func (c *Client) GetTag(ctx context.Context, id string) (*Tag, error) {
	if id == "" {
		return nil, ErrEmptyTagID
	}

	var tag Tag
	err := c.Do(ctx, "GET", "/v1/project_tag/"+url.PathEscape(id), nil, &tag)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// ListTags lists tags, optionally filtered by API-native query parameters.
func (c *Client) ListTags(ctx context.Context, opts *ListTagsOptions) (*ListTagsResponse, error) {
	path := "/v1/project_tag"

	if opts != nil {
		params := url.Values{}
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
		if opts.OrgName != "" {
			params.Set("org_name", opts.OrgName)
		}
		if opts.ProjectID != "" {
			params.Set("project_id", opts.ProjectID)
		}
		if opts.ProjectName != "" {
			params.Set("project_name", opts.ProjectName)
		}
		if opts.TagName != "" {
			params.Set("project_tag_name", opts.TagName)
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListTagsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
