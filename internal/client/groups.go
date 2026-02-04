package client

import (
	"context"
	"fmt"
)

// Group represents a Braintrust group
type Group struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	OrgID       string   `json:"org_id,omitempty"`
	Description string   `json:"description,omitempty"`
	Created     string   `json:"created,omitempty"`
	DeletedAt   string   `json:"deleted_at,omitempty"`
	MemberIDs   []string `json:"member_ids,omitempty"`
}

// CreateGroupRequest represents a request to create a group
// Note: member_ids are not supported during creation and must be added via update
type CreateGroupRequest struct {
	Name        string `json:"name"`
	OrgID       string `json:"org_id,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateGroupRequest represents a request to update a group
type UpdateGroupRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	MemberIDs   []string `json:"member_ids,omitempty"`
}

// ListGroupsOptions represents options for listing groups
type ListGroupsOptions struct {
	OrgID  string
	Cursor string
	Limit  int
}

// ListGroupsResponse represents a list of groups
type ListGroupsResponse struct {
	Cursor string  `json:"cursor,omitempty"`
	Groups []Group `json:"objects"`
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, req *CreateGroupRequest) (*Group, error) {
	var group Group
	err := c.Do(ctx, "POST", "/v1/group", req, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroup retrieves a group by ID
func (c *Client) GetGroup(ctx context.Context, id string) (*Group, error) {
	var group Group
	err := c.Do(ctx, "GET", fmt.Sprintf("/v1/group/%s", id), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup updates an existing group
func (c *Client) UpdateGroup(ctx context.Context, id string, req *UpdateGroupRequest) (*Group, error) {
	var group Group
	err := c.Do(ctx, "PATCH", fmt.Sprintf("/v1/group/%s", id), req, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// DeleteGroup deletes a group
func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", fmt.Sprintf("/v1/group/%s", id), nil, nil)
}

// ListGroups lists all groups for an organization
func (c *Client) ListGroups(ctx context.Context, opts *ListGroupsOptions) (*ListGroupsResponse, error) {
	path := "/v1/group"

	// Build query parameters
	if opts != nil && opts.OrgID != "" {
		path += "?org_id=" + opts.OrgID

		if opts.Limit > 0 {
			path += fmt.Sprintf("&limit=%d", opts.Limit)
		}

		if opts.Cursor != "" {
			path += "&cursor=" + opts.Cursor
		}
	}

	var result ListGroupsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
