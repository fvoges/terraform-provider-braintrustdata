package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrEmptyScoreID is returned when a score ID is empty.
var ErrEmptyScoreID = errors.New("score ID cannot be empty")

// ProjectScore represents a Braintrust project score.
type ProjectScore struct {
	Categories  interface{} `json:"categories,omitempty"`
	Config      interface{} `json:"config,omitempty"`
	Position    *string     `json:"position,omitempty"`
	ID          string      `json:"id"`
	ProjectID   string      `json:"project_id"`
	UserID      string      `json:"user_id,omitempty"`
	Created     string      `json:"created,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	ScoreType   string      `json:"score_type,omitempty"`
}

// CreateScoreRequest represents a request to create a score.
type CreateScoreRequest struct {
	Categories  interface{} `json:"categories,omitempty"`
	Config      interface{} `json:"config,omitempty"`
	ProjectID   string      `json:"project_id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	ScoreType   string      `json:"score_type"`
}

// UpdateScoreRequest represents a request to update a score.
type UpdateScoreRequest struct {
	Categories  *interface{} `json:"categories,omitempty"`
	Config      *interface{} `json:"config,omitempty"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	ScoreType   *string      `json:"score_type,omitempty"`
}

// ListScoresOptions represents options for listing scores.
type ListScoresOptions struct {
	StartingAfter string
	EndingBefore  string
	OrgName       string
	ProjectID     string
	ProjectName   string
	ScoreName     string
	ScoreType     string
	IDs           []string
	Limit         int
}

// ListScoresResponse represents a list of scores.
type ListScoresResponse struct {
	Objects []ProjectScore `json:"objects"`
}

func scorePath(id string) string {
	return "/v1/project_score/" + url.PathEscape(id)
}

// CreateScore creates a new score.
func (c *Client) CreateScore(ctx context.Context, req *CreateScoreRequest) (*ProjectScore, error) {
	var score ProjectScore
	err := c.Do(ctx, "POST", "/v1/project_score", req, &score)
	if err != nil {
		return nil, err
	}

	return &score, nil
}

// GetScore retrieves a score by ID.
func (c *Client) GetScore(ctx context.Context, id string) (*ProjectScore, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyScoreID
	}

	var score ProjectScore
	err := c.Do(ctx, "GET", scorePath(id), nil, &score)
	if err != nil {
		return nil, err
	}

	return &score, nil
}

// UpdateScore updates an existing score.
func (c *Client) UpdateScore(ctx context.Context, id string, req *UpdateScoreRequest) (*ProjectScore, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrEmptyScoreID
	}

	var score ProjectScore
	err := c.Do(ctx, "PATCH", scorePath(id), req, &score)
	if err != nil {
		return nil, err
	}

	return &score, nil
}

// DeleteScore deletes a score by ID.
func (c *Client) DeleteScore(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrEmptyScoreID
	}

	return c.Do(ctx, "DELETE", scorePath(id), nil, nil)
}

// ListScores lists scores using API-native filters.
func (c *Client) ListScores(ctx context.Context, opts *ListScoresOptions) (*ListScoresResponse, error) {
	path := "/v1/project_score"

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
		if opts.ScoreName != "" {
			params.Set("project_score_name", opts.ScoreName)
		}
		if opts.ProjectName != "" {
			params.Set("project_name", opts.ProjectName)
		}
		if opts.ProjectID != "" {
			params.Set("project_id", opts.ProjectID)
		}
		if opts.ScoreType != "" {
			params.Set("score_type", opts.ScoreType)
		}
		if opts.OrgName != "" {
			params.Set("org_name", opts.OrgName)
		}

		if encodedParams := params.Encode(); encodedParams != "" {
			path += "?" + encodedParams
		}
	}

	var result ListScoresResponse
	err := c.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
