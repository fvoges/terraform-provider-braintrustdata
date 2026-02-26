package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyOrganizationID is returned when an organization ID is empty.
var ErrEmptyOrganizationID = errors.New("organization ID cannot be empty")

// Organization represents a Braintrust organization.
type Organization struct {
	APIURL             *string `json:"api_url,omitempty"`
	IsUniversalAPI     *bool   `json:"is_universal_api,omitempty"`
	IsDataplanePrivate *bool   `json:"is_dataplane_private,omitempty"`
	ProxyURL           *string `json:"proxy_url,omitempty"`
	RealtimeURL        *string `json:"realtime_url,omitempty"`
	Created            *string `json:"created,omitempty"`
	ImageRenderingMode *string `json:"image_rendering_mode,omitempty"`
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
}

// PatchOrganizationRequest represents fields to update for an organization.
type PatchOrganizationRequest struct {
	Name               *string `json:"name,omitempty"`
	APIURL             *string `json:"api_url,omitempty"`
	IsUniversalAPI     *bool   `json:"is_universal_api,omitempty"`
	IsDataplanePrivate *bool   `json:"is_dataplane_private,omitempty"`
	ProxyURL           *string `json:"proxy_url,omitempty"`
	RealtimeURL        *string `json:"realtime_url,omitempty"`
	ImageRenderingMode *string `json:"image_rendering_mode,omitempty"`
}

// ListOrganizationsOptions represents options for listing organizations.
type ListOrganizationsOptions struct {
	OrgName       string
	StartingAfter string
	EndingBefore  string
	Limit         int
}

// ListOrganizationsResponse represents a list of organizations.
type ListOrganizationsResponse struct {
	Organizations []Organization `json:"objects"`
}

// GetOrganization retrieves an organization by ID.
func (c *Client) GetOrganization(ctx context.Context, id string) (*Organization, error) {
	if id == "" {
		return nil, ErrEmptyOrganizationID
	}

	var org Organization
	if err := c.Do(ctx, "GET", "/v1/organization/"+url.PathEscape(id), nil, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// UpdateOrganization partially updates an organization by ID.
func (c *Client) UpdateOrganization(ctx context.Context, id string, req *PatchOrganizationRequest) (*Organization, error) {
	if id == "" {
		return nil, ErrEmptyOrganizationID
	}

	var org Organization
	if err := c.Do(ctx, "PATCH", "/v1/organization/"+url.PathEscape(id), req, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// ListOrganizations lists organizations using API-native filters.
func (c *Client) ListOrganizations(ctx context.Context, opts *ListOrganizationsOptions) (*ListOrganizationsResponse, error) {
	path := "/v1/organization"

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

		if encoded := params.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var result ListOrganizationsResponse
	if err := c.Do(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
