package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyGroupID is returned when a group ID is empty.
var ErrEmptyGroupID = errors.New("group ID cannot be empty")

// Group represents a Braintrust group
type Group struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	OrgID        string   `json:"org_id,omitempty"`
	Description  string   `json:"description,omitempty"`
	Created      string   `json:"created,omitempty"`
	DeletedAt    string   `json:"deleted_at,omitempty"`
	MemberUsers  []string `json:"member_users,omitempty"`
	MemberGroups []string `json:"member_groups,omitempty"`
}

// CreateGroupRequest represents a request to create a group
type CreateGroupRequest struct {
	Name         string   `json:"name"`
	OrgID        string   `json:"org_id,omitempty"`
	Description  string   `json:"description,omitempty"`
	MemberUsers  []string `json:"member_users,omitempty"`
	MemberGroups []string `json:"member_groups,omitempty"`
}

// UpdateGroupRequest represents a request to update a group
type UpdateGroupRequest struct {
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	MemberUsers  []string `json:"member_users,omitempty"`
	MemberGroups []string `json:"member_groups,omitempty"`
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
	if id == "" {
		return nil, ErrEmptyGroupID
	}
	var group Group
	err := c.Do(ctx, "GET", "/v1/group/"+url.PathEscape(id), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup updates an existing group
func (c *Client) UpdateGroup(ctx context.Context, id string, req *UpdateGroupRequest) (*Group, error) {
	if id == "" {
		return nil, ErrEmptyGroupID
	}
	var group Group
	err := c.Do(ctx, "PATCH", "/v1/group/"+url.PathEscape(id), req, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// DeleteGroup deletes a group
func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	if id == "" {
		return ErrEmptyGroupID
	}
	return c.Do(ctx, "DELETE", "/v1/group/"+url.PathEscape(id), nil, nil)
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
