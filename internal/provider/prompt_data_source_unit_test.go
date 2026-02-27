package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSelectSinglePromptByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		promptName  string
		wantID      string
		prompts     []client.Prompt
	}{
		"finds exact active prompt": {
			prompts: []client.Prompt{
				{ID: "prompt-a", Name: "other"},
				{ID: "prompt-b", Name: "target"},
			},
			promptName: "target",
			wantID:     "prompt-b",
		},
		"ignores deleted prompts": {
			prompts: []client.Prompt{
				{ID: "prompt-a", Name: "target", DeletedAt: "2026-02-01T00:00:00Z"},
				{ID: "prompt-b", Name: "target"},
			},
			promptName: "target",
			wantID:     "prompt-b",
		},
		"returns not found when no exact match": {
			prompts: []client.Prompt{
				{ID: "prompt-a", Name: "other"},
			},
			promptName:  "target",
			wantErrType: errPromptNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			prompts: []client.Prompt{
				{ID: "prompt-a", Name: "target"},
				{ID: "prompt-b", Name: "target"},
			},
			promptName:  "target",
			wantErrType: errMultiplePromptsFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prompt, err := selectSinglePromptByName(tc.prompts, tc.promptName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if prompt == nil {
				t.Fatalf("expected prompt, got nil")
			}
			if prompt.ID != tc.wantID {
				t.Fatalf("expected prompt ID %q, got %q", tc.wantID, prompt.ID)
			}
		})
	}
}

func TestPopulatePromptDataSourceModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := PromptDataSourceModel{}
	prompt := &client.Prompt{
		ID:           "prompt-1",
		Name:         "support-agent",
		ProjectID:    "project-1",
		Slug:         "support-agent",
		Description:  "Support assistant",
		FunctionType: "chat",
		Created:      "2026-02-27T00:00:00Z",
		UserID:       "user-1",
		OrgID:        "org-1",
		Metadata: map[string]interface{}{
			"owner": "ml-team",
			"tier":  1,
		},
		Tags: []string{"production", "support"},
	}

	diags := populatePromptDataSourceModel(ctx, &model, prompt)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "prompt-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "support-agent" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.ProjectID.ValueString() != "project-1" {
		t.Fatalf("project_id mismatch: got=%q", model.ProjectID.ValueString())
	}
	if model.Slug.ValueString() != "support-agent" {
		t.Fatalf("slug mismatch: got=%q", model.Slug.ValueString())
	}
	if model.FunctionType.ValueString() != "chat" {
		t.Fatalf("function_type mismatch: got=%q", model.FunctionType.ValueString())
	}
	if model.Created.ValueString() != "2026-02-27T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}
	if model.OrgID.ValueString() != "org-1" {
		t.Fatalf("org_id mismatch: got=%q", model.OrgID.ValueString())
	}

	var metadata map[string]string
	diags = model.Metadata.ElementsAs(ctx, &metadata, false)
	if diags.HasError() {
		t.Fatalf("unexpected metadata diagnostics: %v", diags)
	}
	if metadata["owner"] != "ml-team" || metadata["tier"] != "1" {
		t.Fatalf("metadata mismatch: got=%v", metadata)
	}

	var tags []string
	diags = model.Tags.ElementsAs(ctx, &tags, false)
	if diags.HasError() {
		t.Fatalf("unexpected tags diagnostics: %v", diags)
	}
	if len(tags) != 2 {
		t.Fatalf("tags mismatch: got=%v", tags)
	}
}

func TestPopulatePromptDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := PromptDataSourceModel{}
	prompt := &client.Prompt{
		ID:        "prompt-2",
		Name:      "fallback-agent",
		ProjectID: "project-2",
	}

	diags := populatePromptDataSourceModel(ctx, &model, prompt)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.Slug.IsNull() {
		t.Fatalf("expected slug to be null")
	}
	if !model.Description.IsNull() {
		t.Fatalf("expected description to be null")
	}
	if !model.FunctionType.IsNull() {
		t.Fatalf("expected function_type to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
	if !model.UserID.IsNull() {
		t.Fatalf("expected user_id to be null")
	}
	if !model.OrgID.IsNull() {
		t.Fatalf("expected org_id to be null")
	}
	if !model.Metadata.IsNull() {
		t.Fatalf("expected metadata to be null")
	}
	if !model.Tags.IsNull() {
		t.Fatalf("expected tags to be null")
	}
}

func TestNormalizedPromptLookupInput(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		model            PromptDataSourceModel
		wantID           string
		wantName         string
		wantProjectID    string
		wantHasID        bool
		wantHasName      bool
		wantHasProjectID bool
	}{
		"trims all lookup inputs": {
			model: PromptDataSourceModel{
				ID:        types.StringValue("  prompt-1  "),
				Name:      types.StringValue("  support-agent\t"),
				ProjectID: types.StringValue("\nproject-1  "),
			},
			wantID:           "prompt-1",
			wantName:         "support-agent",
			wantProjectID:    "project-1",
			wantHasID:        true,
			wantHasName:      true,
			wantHasProjectID: true,
		},
		"treats whitespace-only values as empty": {
			model: PromptDataSourceModel{
				ID:        types.StringValue(" \t "),
				Name:      types.StringValue("\n"),
				ProjectID: types.StringValue("  "),
			},
			wantHasID:        false,
			wantHasName:      false,
			wantHasProjectID: false,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			id, promptName, projectID, hasID, hasName, hasProjectID := normalizedPromptLookupInput(tc.model)
			if id != tc.wantID {
				t.Fatalf("id mismatch: got=%q want=%q", id, tc.wantID)
			}
			if promptName != tc.wantName {
				t.Fatalf("name mismatch: got=%q want=%q", promptName, tc.wantName)
			}
			if projectID != tc.wantProjectID {
				t.Fatalf("project_id mismatch: got=%q want=%q", projectID, tc.wantProjectID)
			}
			if hasID != tc.wantHasID {
				t.Fatalf("hasID mismatch: got=%t want=%t", hasID, tc.wantHasID)
			}
			if hasName != tc.wantHasName {
				t.Fatalf("hasName mismatch: got=%t want=%t", hasName, tc.wantHasName)
			}
			if hasProjectID != tc.wantHasProjectID {
				t.Fatalf("hasProjectID mismatch: got=%t want=%t", hasProjectID, tc.wantHasProjectID)
			}
		})
	}
}
