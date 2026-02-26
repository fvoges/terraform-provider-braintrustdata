package provider

import (
	"errors"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSelectSingleEnvironmentVariableByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		envVarName  string
		wantID      string
		envVars     []client.EnvironmentVariable
	}{
		"finds exact environment variable": {
			envVars: []client.EnvironmentVariable{
				{ID: "env-var-a", Name: "OTHER_KEY"},
				{ID: "env-var-b", Name: "OPENAI_API_KEY"},
			},
			envVarName: "OPENAI_API_KEY",
			wantID:     "env-var-b",
		},
		"returns not found when no exact match": {
			envVars: []client.EnvironmentVariable{
				{ID: "env-var-a", Name: "OTHER_KEY"},
			},
			envVarName:  "OPENAI_API_KEY",
			wantErrType: errEnvironmentVariableNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			envVars: []client.EnvironmentVariable{
				{ID: "env-var-a", Name: "OPENAI_API_KEY"},
				{ID: "env-var-b", Name: "OPENAI_API_KEY"},
			},
			envVarName:  "OPENAI_API_KEY",
			wantErrType: errMultipleEnvironmentVariablesFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			envVar, err := selectSingleEnvironmentVariableByName(tc.envVars, tc.envVarName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if envVar == nil {
				t.Fatalf("expected environment variable, got nil")
			}
			if envVar.ID != tc.wantID {
				t.Fatalf("expected environment variable ID %q, got %q", tc.wantID, envVar.ID)
			}
		})
	}
}

func TestPopulateEnvironmentVariableDataSourceModel(t *testing.T) {
	t.Parallel()

	model := EnvironmentVariableDataSourceModel{}
	envVar := &client.EnvironmentVariable{
		ID:          "env-var-1",
		Name:        "OPENAI_API_KEY",
		ObjectType:  "project",
		ObjectID:    "project-1",
		Description: "Used by evaluation prompts",
		Created:     "2026-02-26T00:00:00Z",
	}

	populateEnvironmentVariableDataSourceModel(&model, envVar)

	if model.ID.ValueString() != "env-var-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "OPENAI_API_KEY" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.ObjectType.ValueString() != "project" {
		t.Fatalf("object_type mismatch: got=%q", model.ObjectType.ValueString())
	}
	if model.ObjectID.ValueString() != "project-1" {
		t.Fatalf("object_id mismatch: got=%q", model.ObjectID.ValueString())
	}
	if model.Description.ValueString() != "Used by evaluation prompts" {
		t.Fatalf("description mismatch: got=%q", model.Description.ValueString())
	}
	if model.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
}

func TestPopulateEnvironmentVariableDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := EnvironmentVariableDataSourceModel{}
	envVar := &client.EnvironmentVariable{
		ID:   "env-var-2",
		Name: "OPENAI_API_KEY",
	}

	populateEnvironmentVariableDataSourceModel(&model, envVar)

	if !model.ObjectType.IsNull() {
		t.Fatalf("expected object_type to be null")
	}
	if !model.ObjectID.IsNull() {
		t.Fatalf("expected object_id to be null")
	}
	if !model.Description.IsNull() {
		t.Fatalf("expected description to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
}

func TestTrimEnvironmentVariableLookupInputs(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		model          EnvironmentVariableDataSourceModel
		wantID         string
		wantName       string
		wantObjectType string
		wantObjectID   string
	}{
		"trims surrounding whitespace": {
			model: EnvironmentVariableDataSourceModel{
				ID:         types.StringValue("  env-var-123  "),
				Name:       types.StringValue("  OPENAI_API_KEY  "),
				ObjectType: types.StringValue("  project  "),
				ObjectID:   types.StringValue("  project-123  "),
			},
			wantID:         "env-var-123",
			wantName:       "OPENAI_API_KEY",
			wantObjectType: "project",
			wantObjectID:   "project-123",
		},
		"normalizes whitespace-only values to empty": {
			model: EnvironmentVariableDataSourceModel{
				ID:         types.StringValue(" \t "),
				Name:       types.StringValue("  "),
				ObjectType: types.StringValue("\n"),
				ObjectID:   types.StringValue("\t"),
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			id, lookupName, objectType, objectID := trimEnvironmentVariableLookupInputs(tc.model)
			if id != tc.wantID {
				t.Fatalf("expected trimmed id %q, got %q", tc.wantID, id)
			}
			if lookupName != tc.wantName {
				t.Fatalf("expected trimmed name %q, got %q", tc.wantName, lookupName)
			}
			if objectType != tc.wantObjectType {
				t.Fatalf("expected trimmed object_type %q, got %q", tc.wantObjectType, objectType)
			}
			if objectID != tc.wantObjectID {
				t.Fatalf("expected trimmed object_id %q, got %q", tc.wantObjectID, objectID)
			}
		})
	}
}

func TestValidateEnvironmentVariableLookupAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		name        string
		objectType  string
		objectID    string
		wantErrLike []string
	}{
		"accepts valid lookup values": {
			name:       "OPENAI_API_KEY",
			objectType: "project",
			objectID:   "project-123",
		},
		"rejects blank lookup values": {
			wantErrLike: []string{
				"'name' must be provided and non-empty",
				"'object_type' must be provided and non-empty",
				"'object_id' must be provided and non-empty",
			},
		},
		"rejects missing object_id only": {
			name:       "OPENAI_API_KEY",
			objectType: "project",
			wantErrLike: []string{
				"'object_id' must be provided and non-empty",
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := validateEnvironmentVariableLookupAttributes(tc.name, tc.objectType, tc.objectID)
			if len(tc.wantErrLike) == 0 {
				if diags.HasError() {
					t.Fatalf("unexpected diagnostics: %v", diags)
				}
				return
			}

			if !diags.HasError() {
				t.Fatalf("expected diagnostics, got none")
			}

			for _, expected := range tc.wantErrLike {
				found := false
				for _, diag := range diags {
					if strings.Contains(diag.Detail(), expected) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected diagnostic containing %q, got %v", expected, diags)
				}
			}
		})
	}
}
