package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrEmptyFunctionID is returned when a function ID is empty.
var ErrEmptyFunctionID = errors.New("function ID cannot be empty")

const functionNotFoundMessage = "Function does not exist or you do not have access"

// Function represents a Braintrust function.
type Function struct {
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	FunctionData   interface{}            `json:"function_data,omitempty"`
	FunctionSchema interface{}            `json:"function_schema,omitempty"`
	Origin         interface{}            `json:"origin,omitempty"`
	PromptData     interface{}            `json:"prompt_data,omitempty"`
	XactID         string                 `json:"_xact_id,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Description    string                 `json:"description,omitempty"`
	FunctionType   string                 `json:"function_type,omitempty"`
	ID             string                 `json:"id"`
	LogID          string                 `json:"log_id,omitempty"`
	Name           string                 `json:"name"`
	OrgID          string                 `json:"org_id,omitempty"`
	ProjectID      string                 `json:"project_id,omitempty"`
	Slug           string                 `json:"slug,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
}

// ListFunctionsOptions represents options for listing functions.
type ListFunctionsOptions struct {
	ProjectID     string
	FunctionName  string
	Slug          string
	Limit         *int
	StartingAfter string
	EndingBefore  string
}

// ListFunctionsResponse represents a list of functions.
type ListFunctionsResponse struct {
	Functions []Function `json:"objects"`
}

// CreateFunctionRequest represents a request to create a function.
type CreateFunctionRequest struct {
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	FunctionData   interface{}            `json:"function_data"`
	FunctionSchema interface{}            `json:"function_schema,omitempty"`
	Origin         interface{}            `json:"origin,omitempty"`
	PromptData     interface{}            `json:"prompt_data,omitempty"`
	ProjectID      string                 `json:"project_id"`
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Description    string                 `json:"description,omitempty"`
	FunctionType   string                 `json:"function_type,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
}

// UpdateFunctionRequest represents a request to update a function.
type UpdateFunctionRequest struct {
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
	FunctionData   *interface{}            `json:"function_data,omitempty"`
	FunctionSchema *interface{}            `json:"function_schema,omitempty"`
	Origin         *interface{}            `json:"origin,omitempty"`
	PromptData     *interface{}            `json:"prompt_data,omitempty"`
	Name           *string                 `json:"name,omitempty"`
	Slug           *string                 `json:"slug,omitempty"`
	Description    *string                 `json:"description,omitempty"`
	FunctionType   *string                 `json:"function_type,omitempty"`
	Tags           *[]string               `json:"tags,omitempty"`
}

func functionPath(id string) string {
	return "/v1/function/" + url.PathEscape(id)
}

// CreateFunction creates a new function.
func (c *Client) CreateFunction(ctx context.Context, req *CreateFunctionRequest) (*Function, error) {
	var function Function
	err := c.Do(ctx, "POST", "/v1/function", req, &function)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

// UpdateFunction updates an existing function.
func (c *Client) UpdateFunction(ctx context.Context, id string, req *UpdateFunctionRequest) (*Function, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyFunctionID
	}

	var function Function
	err := c.Do(ctx, "PATCH", functionPath(id), req, &function)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

// DeleteFunction deletes a function by ID.
func (c *Client) DeleteFunction(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrEmptyFunctionID
	}

	return c.Do(ctx, "DELETE", functionPath(id), nil, nil)
}

// GetFunction retrieves a function by ID.
func (c *Client) GetFunction(ctx context.Context, id string) (*Function, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyFunctionID
	}

	var function Function
	err := c.Do(ctx, "GET", functionPath(id), nil, &function)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

// ListFunctions lists functions using API-native filters.
func (c *Client) ListFunctions(ctx context.Context, opts *ListFunctionsOptions) (*ListFunctionsResponse, error) {
	path := "/v1/function"

	if opts != nil {
		params := url.Values{}
		if opts.ProjectID != "" {
			params.Set("project_id", opts.ProjectID)
		}
		if opts.FunctionName != "" {
			params.Set("function_name", opts.FunctionName)
		}
		if opts.Slug != "" {
			params.Set("slug", opts.Slug)
		}
		if opts.Limit != nil {
			params.Set("limit", fmt.Sprintf("%d", *opts.Limit))
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

	var result ListFunctionsResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// IsFunctionNotFound returns true when the API reports function access/not-found semantics.
func IsFunctionNotFound(err error) bool {
	apiErr := &APIError{}
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode != 400 {
			return false
		}
		return strings.Contains(strings.ToLower(apiErr.Message), strings.ToLower(functionNotFoundMessage))
	}

	return false
}
