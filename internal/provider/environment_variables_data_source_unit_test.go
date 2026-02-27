package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListEnvironmentVariablesOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListEnvironmentVariablesOptions
		model       EnvironmentVariablesDataSourceModel
	}{
		"builds all supported api-native filters": {
			model: EnvironmentVariablesDataSourceModel{
				ObjectType:    types.StringValue("project"),
				ObjectID:      types.StringValue("project-1"),
				StartingAfter: types.StringValue("env-var-1"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListEnvironmentVariablesOptions{
				ObjectType:    "project",
				ObjectID:      "project-1",
				StartingAfter: "env-var-1",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: EnvironmentVariablesDataSourceModel{
				ObjectType:    types.StringValue("project"),
				ObjectID:      types.StringValue("project-1"),
				StartingAfter: types.StringValue("env-var-1"),
				EndingBefore:  types.StringValue("env-var-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: EnvironmentVariablesDataSourceModel{
				ObjectType: types.StringValue("project"),
				ObjectID:   types.StringValue("project-1"),
				Limit:      types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
		"rejects empty object_type": {
			model: EnvironmentVariablesDataSourceModel{
				ObjectType: types.StringValue(""),
				ObjectID:   types.StringValue("project-1"),
			},
			wantErrLike: "'object_type' must be provided and non-empty",
		},
		"rejects whitespace object_id": {
			model: EnvironmentVariablesDataSourceModel{
				ObjectType: types.StringValue("project"),
				ObjectID:   types.StringValue(" "),
			},
			wantErrLike: "'object_id' must be provided and non-empty",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListEnvironmentVariablesOptions(tc.model)
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
			if opts.ObjectType != tc.want.ObjectType ||
				opts.ObjectID != tc.want.ObjectID ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestEnvironmentVariablesDataSourceEnvironmentVariableFromEnvironmentVariable(t *testing.T) {
	t.Parallel()

	envVarModel := environmentVariablesDataSourceEnvironmentVariableFromEnvironmentVariable(&client.EnvironmentVariable{
		ID:          "env-var-1",
		Name:        "OPENAI_API_KEY",
		ObjectType:  "project",
		ObjectID:    "project-1",
		Description: "Used by evaluation prompts",
		Created:     "2024-01-15T10:30:00Z",
	})

	if envVarModel.ID.ValueString() != "env-var-1" {
		t.Fatalf("id mismatch: got=%q", envVarModel.ID.ValueString())
	}
	if envVarModel.Name.ValueString() != "OPENAI_API_KEY" {
		t.Fatalf("name mismatch: got=%q", envVarModel.Name.ValueString())
	}
	if envVarModel.ObjectType.ValueString() != "project" {
		t.Fatalf("object_type mismatch: got=%q", envVarModel.ObjectType.ValueString())
	}
	if envVarModel.ObjectID.ValueString() != "project-1" {
		t.Fatalf("object_id mismatch: got=%q", envVarModel.ObjectID.ValueString())
	}
	if envVarModel.Description.ValueString() != "Used by evaluation prompts" {
		t.Fatalf("description mismatch: got=%q", envVarModel.Description.ValueString())
	}
	if envVarModel.Created.ValueString() != "2024-01-15T10:30:00Z" {
		t.Fatalf("created mismatch: got=%q", envVarModel.Created.ValueString())
	}
}

func TestProviderDataSourcesIncludeEnvironmentVariablePair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_environment_variable"]; !ok {
		t.Fatalf("expected braintrustdata_environment_variable to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_environment_variables"]; !ok {
		t.Fatalf("expected braintrustdata_environment_variables to be registered")
	}
}
