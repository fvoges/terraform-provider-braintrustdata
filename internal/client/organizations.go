package client

import (
	"context"
	"errors"
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
