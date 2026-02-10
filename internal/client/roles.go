package client

import (
	"context"
	"fmt"
	"net/url"
)

// Role represents a Braintrust role
type Role struct {
	ID                string   `json:"id"`
	OrgID             string   `json:"org_id,omitempty"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	Created           string   `json:"created,omitempty"`
	DeletedAt         string   `json:"deleted_at,omitempty"`
	UserID            string   `json:"user_id,omitempty"`
	MemberPermissions []string `json:"member_permissions,omitempty"` // Simplified for now
	MemberRoles       []string `json:"member_roles,omitempty"`
}

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	OrgName           string   `json:"org_name,omitempty"`
	MemberPermissions []string `json:"member_permissions,omitempty"`
	MemberRoles       []string `json:"member_roles,omitempty"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name                    string   `json:"name,omitempty"`
	Description             string   `json:"description,omitempty"`
	AddMemberPermissions    []string `json:"add_member_permissions,omitempty"`
	RemoveMemberPermissions []string `json:"remove_member_permissions,omitempty"`
	AddMemberRoles          []string `json:"add_member_roles,omitempty"`
	RemoveMemberRoles       []string `json:"remove_member_roles,omitempty"`
}

// ListRolesOptions represents options for listing roles
type ListRolesOptions struct {
	OrgName       string
	StartingAfter string
	EndingBefore  string
	RoleName      string
	Limit         int
}

// ListRolesResponse represents a list of roles
type ListRolesResponse struct {
	Roles []Role `json:"objects"`
}

// CreateRole creates a new role
func (c *Client) CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error) {
	var role Role
	err := c.Do(ctx, "POST", "/v1/role", req, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRole retrieves a role by ID
func (c *Client) GetRole(ctx context.Context, id string) (*Role, error) {
	var role Role
	err := c.Do(ctx, "GET", "/v1/role/"+url.PathEscape(id), nil, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// UpdateRole updates an existing role
func (c *Client) UpdateRole(ctx context.Context, id string, req *UpdateRoleRequest) (*Role, error) {
	var role Role
	err := c.Do(ctx, "PATCH", "/v1/role/"+url.PathEscape(id), req, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// DeleteRole deletes a role (soft delete)
func (c *Client) DeleteRole(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/v1/role/"+url.PathEscape(id), nil, nil)
}

// ListRoles lists all roles
func (c *Client) ListRoles(ctx context.Context, opts *ListRolesOptions) (*ListRolesResponse, error) {
	path := "/v1/role"

	// Build query parameters
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

		if opts.RoleName != "" {
			params.Set("role_name", opts.RoleName)
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListRolesResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
