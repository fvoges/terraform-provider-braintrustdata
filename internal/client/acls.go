package client

import (
	"context"
	"fmt"
)

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
	var acl ACL
	err := c.Do(ctx, "GET", fmt.Sprintf("/v1/acl/%s", id), nil, &acl)
	if err != nil {
		return nil, err
	}
	return &acl, nil
}

// DeleteACL deletes an ACL
func (c *Client) DeleteACL(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", fmt.Sprintf("/v1/acl/%s", id), nil, nil)
}

// ListACLs lists all ACLs for a given object
func (c *Client) ListACLs(ctx context.Context, opts *ListACLsOptions) (*ListACLsResponse, error) {
	path := "/v1/acl"

	// Build query parameters
	if opts != nil {
		separator := "?"

		// object_id and object_type are required
		if opts.ObjectID != "" {
			path += separator + "object_id=" + opts.ObjectID
			separator = "&"
		}

		if opts.ObjectType != "" {
			path += separator + "object_type=" + string(opts.ObjectType)
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
		}
	}

	var result ListACLsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
