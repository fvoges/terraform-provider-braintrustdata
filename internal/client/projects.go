package client

import (
	"context"
	"fmt"
	"net/url"
)

// Project represents a Braintrust project
type Project struct {
	ID          string `json:"id"`
	OrgID       string `json:"org_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Created     string `json:"created,omitempty"`
	DeletedAt   string `json:"deleted_at,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Settings    string `json:"settings,omitempty"` // JSON string for now, will expand if needed
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	OrgName     string `json:"org_name,omitempty"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Settings    string `json:"settings,omitempty"`
}

// ListProjectsOptions represents options for listing projects
type ListProjectsOptions struct {
	OrgName       string
	StartingAfter string
	EndingBefore  string
	ProjectName   string
	Limit         int
}

// ListProjectsResponse represents a list of projects
type ListProjectsResponse struct {
	Projects []Project `json:"objects"`
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, req *CreateProjectRequest) (*Project, error) {
	var project Project
	err := c.Do(ctx, "POST", "/v1/project", req, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProject retrieves a project by ID
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	var project Project
	err := c.Do(ctx, "GET", "/v1/project/"+url.PathEscape(id), nil, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, id string, req *UpdateProjectRequest) (*Project, error) {
	var project Project
	err := c.Do(ctx, "PATCH", "/v1/project/"+url.PathEscape(id), req, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject deletes a project (soft delete)
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/v1/project/"+url.PathEscape(id), nil, nil)
}

// ListProjects lists all projects
func (c *Client) ListProjects(ctx context.Context, opts *ListProjectsOptions) (*ListProjectsResponse, error) {
	path := "/v1/project"

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

		if opts.ProjectName != "" {
			path += separator + "project_name=" + opts.ProjectName
		}
	}

	var result ListProjectsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
