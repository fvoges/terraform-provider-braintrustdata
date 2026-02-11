package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyACLID is returned when an ACL ID is empty.
var ErrEmptyACLID = errors.New("ACL ID cannot be empty")

// ACLObjectType represents the type of object an ACL applies to
type ACLObjectType string

// ACL object type constants
const (
	ACLObjectTypeOrganization  ACLObjectType = "organization"
	ACLObjectTypeProject       ACLObjectType = "project"
	ACLObjectTypeExperiment    ACLObjectType = "experiment"
	ACLObjectTypeDataset       ACLObjectType = "dataset"
	ACLObjectTypePrompt        ACLObjectType = "prompt"
	ACLObjectTypePromptSession ACLObjectType = "prompt_session"
	ACLObjectTypeGroup         ACLObjectType = "group"
	ACLObjectTypeRole          ACLObjectType = "role"
	ACLObjectTypeOrgMember     ACLObjectType = "org_member"
	ACLObjectTypeProjectLog    ACLObjectType = "project_log"
	ACLObjectTypeOrgProject    ACLObjectType = "org_project"
)

// Permission represents the permission level granted by an ACL
type Permission string

// Permission level constants
const (
	PermissionCreate     Permission = "create"
	PermissionRead       Permission = "read"
	PermissionUpdate     Permission = "update"
	PermissionDelete     Permission = "delete"
	PermissionCreateACLs Permission = "create_acls"
	PermissionReadACLs   Permission = "read_acls"
	PermissionUpdateACLs Permission = "update_acls"
	PermissionDeleteACLs Permission = "delete_acls"
)

// ACL represents a Braintrust access control list entry
type ACL struct {
	ID                 string        `json:"id"`
	ObjectOrgID        string        `json:"_object_org_id,omitempty"`
	ObjectID           string        `json:"object_id"`
	ObjectType         ACLObjectType `json:"object_type"`
	Created            string        `json:"created,omitempty"`
	GroupID            string        `json:"group_id,omitempty"`
	Permission         Permission    `json:"permission,omitempty"`
	RestrictObjectType ACLObjectType `json:"restrict_object_type,omitempty"`
	RoleID             string        `json:"role_id,omitempty"`
	UserID             string        `json:"user_id,omitempty"`
}

// CreateACLRequest represents a request to create an ACL
type CreateACLRequest struct {
	ObjectID           string        `json:"object_id"`
	ObjectType         ACLObjectType `json:"object_type"`
	GroupID            string        `json:"group_id,omitempty"`
	UserID             string        `json:"user_id,omitempty"`
	RoleID             string        `json:"role_id,omitempty"`
	Permission         Permission    `json:"permission,omitempty"`
	RestrictObjectType ACLObjectType `json:"restrict_object_type,omitempty"`
}

// ListACLsOptions represents options for listing ACLs
type ListACLsOptions struct {
	ObjectID      string
	ObjectType    ACLObjectType
	StartingAfter string
	EndingBefore  string
	Limit         int
}

// ListACLsResponse represents a list of ACLs
type ListACLsResponse struct {
	Objects []ACL `json:"objects"`
}

// CreateACL creates a new ACL
func (c *Client) CreateACL(ctx context.Context, req *CreateACLRequest) (*ACL, error) {
	var acl ACL
	err := c.Do(ctx, "POST", "/v1/acl", req, &acl)
	if err != nil {
		return nil, err
	}
	return &acl, nil
}

// GetACL retrieves an ACL by ID
func (c *Client) GetACL(ctx context.Context, id string) (*ACL, error) {
	if id == "" {
		return nil, ErrEmptyACLID
	}
	var acl ACL
	err := c.Do(ctx, "GET", "/v1/acl/"+url.PathEscape(id), nil, &acl)
	if err != nil {
		return nil, err
	}
	return &acl, nil
}

// DeleteACL deletes an ACL
func (c *Client) DeleteACL(ctx context.Context, id string) error {
	if id == "" {
		return ErrEmptyACLID
	}
	return c.Do(ctx, "DELETE", "/v1/acl/"+url.PathEscape(id), nil, nil)
}

// ListACLs lists all ACLs for a given object
func (c *Client) ListACLs(ctx context.Context, opts *ListACLsOptions) (*ListACLsResponse, error) {
	path := "/v1/acl"

	// Build query parameters
	if opts != nil {
		params := url.Values{}

		// object_id and object_type are required
		if opts.ObjectID != "" {
			params.Set("object_id", opts.ObjectID)
		}

		if opts.ObjectType != "" {
			params.Set("object_type", string(opts.ObjectType))
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

	var result ListACLsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
