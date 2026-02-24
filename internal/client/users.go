package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrEmptyUserID is returned when a user ID is empty.
var ErrEmptyUserID = errors.New("user ID cannot be empty")

// userPath returns the API path for a specific user.
func userPath(id string) string {
	return "/v1/user/" + url.PathEscape(id)
}

// User represents a Braintrust user.
type User struct {
	ID         string `json:"id"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
	Email      string `json:"email,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Created    string `json:"created,omitempty"`
}

// ListUsersOptions represents options for listing users.
type ListUsersOptions struct {
	StartingAfter string
	EndingBefore  string
	OrgName       string
	IDs           []string
	GivenNames    []string
	FamilyNames   []string
	Emails        []string
	Limit         int
}

// ListUsersResponse represents a list of users.
type ListUsersResponse struct {
	Users []User `json:"objects"`
}

// GetUser retrieves a user by ID.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, ErrEmptyUserID
	}

	var user User
	err := c.Do(ctx, "GET", userPath(id), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers lists users, optionally filtered by API-native query parameters.
func (c *Client) ListUsers(ctx context.Context, opts *ListUsersOptions) (*ListUsersResponse, error) {
	path := "/v1/user"

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
		for _, givenName := range opts.GivenNames {
			if givenName != "" {
				params.Add("given_name", givenName)
			}
		}
		for _, familyName := range opts.FamilyNames {
			if familyName != "" {
				params.Add("family_name", familyName)
			}
		}
		for _, email := range opts.Emails {
			if email != "" {
				params.Add("email", email)
			}
		}
		if opts.OrgName != "" {
			params.Set("org_name", opts.OrgName)
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListUsersResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
