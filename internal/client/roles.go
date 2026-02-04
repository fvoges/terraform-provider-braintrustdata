package client

import (
	"context"
	"fmt"
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
	Name                     string   `json:"name,omitempty"`
	Description              string   `json:"description,omitempty"`
	AddMemberPermissions     []string `json:"add_member_permissions,omitempty"`
	RemoveMemberPermissions  []string `json:"remove_member_permissions,omitempty"`
	AddMemberRoles           []string `json:"add_member_roles,omitempty"`
	RemoveMemberRoles        []string `json:"remove_member_roles,omitempty"`
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
	err := c.Do(ctx, "GET", fmt.Sprintf("/v1/role/%s", id), nil, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// UpdateRole updates an existing role
func (c *Client) UpdateRole(ctx context.Context, id string, req *UpdateRoleRequest) (*Role, error) {
	var role Role
	err := c.Do(ctx, "PATCH", fmt.Sprintf("/v1/role/%s", id), req, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// DeleteRole deletes a role (soft delete)
func (c *Client) DeleteRole(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", fmt.Sprintf("/v1/role/%s", id), nil, nil)
}

// ListRoles lists all roles
func (c *Client) ListRoles(ctx context.Context, opts *ListRolesOptions) (*ListRolesResponse, error) {
	path := "/v1/role"

	// Build query parameters
	if opts != nil {
		separator := "?"

		if opts.OrgName != "" {
			path += separator + "org_name=" + opts.OrgName
			separator = "&"
		}

		if opts.Limit > 0 {
			path += fmt.Sprintf("%slimit=%d", separator, opts.Limit)
			separator = "&"
		}

		if opts.StartingAfter != "" {
			path += separator + "starting_after=" + opts.StartingAfter
			separator = "&"
		}

		if opts.EndingBefore != "" {
			path += separator + "ending_before=" + opts.EndingBefore
			separator = "&"
		}

		if opts.RoleName != "" {
			path += separator + "role_name=" + opts.RoleName
		}
	}

	var result ListRolesResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
