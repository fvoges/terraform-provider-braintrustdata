package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListPromptsOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListPromptsOptions
		model       PromptsDataSourceModel
	}{
		"builds all supported API-native filters": {
			model: PromptsDataSourceModel{
				ProjectID:     types.StringValue("project-1"),
				Name:          types.StringValue("support-agent"),
				Slug:          types.StringValue("support-agent"),
				Version:       types.StringValue("v1"),
				StartingAfter: types.StringValue("prompt-1"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListPromptsOptions{
				ProjectID:     "project-1",
				PromptName:    "support-agent",
				Slug:          "support-agent",
				Version:       "v1",
				StartingAfter: "prompt-1",
				Limit:         10,
			},
		},
		"trims optional filters before list options": {
			model: PromptsDataSourceModel{
				ProjectID: types.StringValue("  project-1 "),
				Name:      types.StringValue("  support-agent\t"),
				Slug:      types.StringValue("\n support-agent "),
				Version:   types.StringValue(" v1 "),
			},
			want: client.ListPromptsOptions{
				ProjectID:  "project-1",
				PromptName: "support-agent",
				Slug:       "support-agent",
				Version:    "v1",
			},
		},
		"ignores whitespace-only optional filters": {
			model: PromptsDataSourceModel{
				ProjectID: types.StringValue("project-1"),
				Name:      types.StringValue("  "),
				Slug:      types.StringValue("\n"),
				Version:   types.StringValue("\t"),
			},
			want: client.ListPromptsOptions{
				ProjectID: "project-1",
			},
		},
		"rejects empty project_id": {
			model: PromptsDataSourceModel{
				ProjectID: types.StringValue(""),
			},
			wantErrLike: "'project_id' must be provided and non-empty",
		},
		"rejects conflicting pagination": {
			model: PromptsDataSourceModel{
				ProjectID:     types.StringValue("project-1"),
				StartingAfter: types.StringValue("prompt-1"),
				EndingBefore:  types.StringValue("prompt-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: PromptsDataSourceModel{
				ProjectID: types.StringValue("project-1"),
				Limit:     types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListPromptsOptions(tc.model)
			if tc.wantErrLike != "" {
				if !diags.HasError() {
					t.Fatalf("expected diagnostic containing %q, got none", tc.wantErrLike)
				}

				found := false
				for _, diag := range diags {
					if strings.Contains(diag.Detail(), tc.wantErrLike) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected diagnostic containing %q, got %v", tc.wantErrLike, diags)
				}
				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if opts == nil {
				t.Fatalf("expected options, got nil")
			}
			if opts.ProjectID != tc.want.ProjectID ||
				opts.PromptName != tc.want.PromptName ||
				opts.Slug != tc.want.Slug ||
				opts.Version != tc.want.Version ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestPromptsDataSourcePromptFromPrompt(t *testing.T) {
	t.Parallel()

	model, diags := promptsDataSourcePromptFromPrompt(context.Background(), &client.Prompt{
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
		},
		Tags: []string{"support"},
	})
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
}

func TestProviderDataSourcesIncludePromptPair(t *testing.T) {
	t.Parallel()

	p, ok := New("test")().(*BraintrustProvider)
	if !ok {
		t.Fatalf("expected *BraintrustProvider")
	}

	dataSourceFactories := p.DataSources(context.Background())
	dataSourceNames := make(map[string]struct{}, len(dataSourceFactories))

	for _, factory := range dataSourceFactories {
		ds := factory()
		resp := &datasource.MetadataResponse{}
		ds.Metadata(context.Background(), datasource.MetadataRequest{
			ProviderTypeName: "braintrustdata",
		}, resp)

		dataSourceNames[resp.TypeName] = struct{}{}
	}

	if _, ok := dataSourceNames["braintrustdata_prompt"]; !ok {
		t.Fatalf("expected braintrustdata_prompt to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_prompts"]; !ok {
		t.Fatalf("expected braintrustdata_prompts to be registered")
	}
}
