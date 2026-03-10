package provider

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizedFunctionLookupInput(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		model            FunctionDataSourceModel
		wantID           string
		wantProjectID    string
		wantName         string
		wantSlug         string
		wantHasID        bool
		wantHasProjectID bool
		wantHasName      bool
		wantHasSlug      bool
	}{
		"trims all lookup inputs": {
			model: FunctionDataSourceModel{
				ID:        types.StringValue("  function-1  "),
				ProjectID: types.StringValue("\nproject-1\t"),
				Name:      types.StringValue("  tool-a "),
				Slug:      types.StringValue("  tool-a-slug "),
			},
			wantID:           "function-1",
			wantProjectID:    "project-1",
			wantName:         "tool-a",
			wantSlug:         "tool-a-slug",
			wantHasID:        true,
			wantHasProjectID: true,
			wantHasName:      true,
			wantHasSlug:      true,
		},
		"treats whitespace-only values as empty": {
			model: FunctionDataSourceModel{
				ID:        types.StringValue(" \n "),
				ProjectID: types.StringValue("  "),
				Name:      types.StringValue("\t"),
				Slug:      types.StringValue("\n"),
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			id, projectID, functionName, slug, hasID, hasProjectID, hasName, hasSlug := normalizedFunctionLookupInput(tc.model)
			if id != tc.wantID {
				t.Fatalf("id mismatch: got=%q want=%q", id, tc.wantID)
			}
			if projectID != tc.wantProjectID {
				t.Fatalf("project_id mismatch: got=%q want=%q", projectID, tc.wantProjectID)
			}
			if functionName != tc.wantName {
				t.Fatalf("name mismatch: got=%q want=%q", functionName, tc.wantName)
			}
			if slug != tc.wantSlug {
				t.Fatalf("slug mismatch: got=%q want=%q", slug, tc.wantSlug)
			}
			if hasID != tc.wantHasID {
				t.Fatalf("hasID mismatch: got=%t want=%t", hasID, tc.wantHasID)
			}
			if hasProjectID != tc.wantHasProjectID {
				t.Fatalf("hasProjectID mismatch: got=%t want=%t", hasProjectID, tc.wantHasProjectID)
			}
			if hasName != tc.wantHasName {
				t.Fatalf("hasName mismatch: got=%t want=%t", hasName, tc.wantHasName)
			}
			if hasSlug != tc.wantHasSlug {
				t.Fatalf("hasSlug mismatch: got=%t want=%t", hasSlug, tc.wantHasSlug)
			}
		})
	}
}

func TestValidateFunctionLookupInput(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		hasID       bool
		hasProject  bool
		hasName     bool
		hasSlug     bool
	}{
		"accepts id only": {
			hasID: true,
		},
		"accepts project_id and name": {
			hasProject: true,
			hasName:    true,
		},
		"accepts project_id and slug": {
			hasProject: true,
			hasSlug:    true,
		},
		"rejects missing lookup attributes": {
			wantErrLike: "Must specify either 'id' or one searchable pair",
		},
		"rejects id with project_id": {
			hasID:       true,
			hasProject:  true,
			wantErrLike: "Cannot combine 'id' with searchable attributes",
		},
		"rejects id with name": {
			hasID:       true,
			hasName:     true,
			wantErrLike: "Cannot combine 'id' with searchable attributes",
		},
		"rejects id with slug": {
			hasID:       true,
			hasSlug:     true,
			wantErrLike: "Cannot combine 'id' with searchable attributes",
		},
		"rejects name and slug together": {
			hasProject:  true,
			hasName:     true,
			hasSlug:     true,
			wantErrLike: "Cannot specify both 'name' and 'slug'",
		},
		"rejects name without project_id": {
			hasName:     true,
			wantErrLike: "'project_id' must be provided when using 'name' or 'slug'",
		},
		"rejects slug without project_id": {
			hasSlug:     true,
			wantErrLike: "'project_id' must be provided when using 'name' or 'slug'",
		},
		"rejects only project_id": {
			hasProject:  true,
			wantErrLike: "Must specify either 'id' or one searchable pair",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := validateFunctionLookupInput(tc.hasID, tc.hasProject, tc.hasName, tc.hasSlug)
			if tc.wantErrLike == "" {
				if diags.HasError() {
					t.Fatalf("unexpected diagnostics: %v", diags)
				}
				return
			}

			if !diags.HasError() {
				t.Fatalf("expected diagnostics with %q, got none", tc.wantErrLike)
			}
			if !strings.Contains(diags[0].Detail(), tc.wantErrLike) {
				t.Fatalf("expected detail containing %q, got %q", tc.wantErrLike, diags[0].Detail())
			}
		})
	}
}

func TestSelectSingleFunctionByField(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		fieldName   string
		fieldValue  string
		wantID      string
		functions   []client.Function
	}{
		"finds exact name match": {
			fieldName:  "name",
			fieldValue: "tool-a",
			functions: []client.Function{
				{ID: "function-1", Name: "tool-b"},
				{ID: "function-2", Name: "tool-a"},
			},
			wantID: "function-2",
		},
		"finds exact slug match": {
			fieldName:  "slug",
			fieldValue: "tool-a",
			functions: []client.Function{
				{ID: "function-1", Slug: "tool-b"},
				{ID: "function-2", Slug: "tool-a"},
			},
			wantID: "function-2",
		},
		"returns not found": {
			fieldName:   "name",
			fieldValue:  "tool-z",
			functions:   []client.Function{{ID: "function-1", Name: "tool-a"}},
			wantErrType: errFunctionNotFoundByField,
		},
		"returns multiple": {
			fieldName:  "name",
			fieldValue: "tool-a",
			functions: []client.Function{
				{ID: "function-1", Name: "tool-a"},
				{ID: "function-2", Name: "tool-a"},
			},
			wantErrType: errMultipleFunctionsFoundByField,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := selectSingleFunctionByField(tc.functions, tc.fieldName, tc.fieldValue)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected function, got nil")
			}
			if got.ID != tc.wantID {
				t.Fatalf("expected function ID %q, got %q", tc.wantID, got.ID)
			}
		})
	}
}

func TestPopulateFunctionDataSourceModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := FunctionDataSourceModel{}
	fn := &client.Function{
		XactID:         "xact-1",
		Created:        "2026-03-10T00:00:00Z",
		Description:    "Tool function",
		FunctionData:   map[string]interface{}{"runtime": "python", "version": 3},
		FunctionSchema: map[string]interface{}{"type": "object"},
		FunctionType:   "tool",
		ID:             "function-1",
		LogID:          "log-1",
		Metadata:       map[string]interface{}{"owner": "ml", "tier": 1},
		Name:           "tool-a",
		OrgID:          "org-1",
		Origin:         map[string]interface{}{"source": "api"},
		ProjectID:      "project-1",
		PromptData:     map[string]interface{}{"prompt": "hello"},
		Slug:           "tool-a",
		Tags:           []string{"prod", "tool"},
	}

	diags := populateFunctionDataSourceModel(ctx, &model, fn)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "function-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "tool-a" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.FunctionType.ValueString() != "tool" {
		t.Fatalf("function_type mismatch: got=%q", model.FunctionType.ValueString())
	}
	if model.XactID.ValueString() != "xact-1" {
		t.Fatalf("xact_id mismatch: got=%q", model.XactID.ValueString())
	}

	assertJSONFieldContainsKey(t, model.FunctionData.ValueString(), "runtime")
	assertJSONFieldContainsKey(t, model.FunctionSchema.ValueString(), "type")
	assertJSONFieldContainsKey(t, model.Origin.ValueString(), "source")
	assertJSONFieldContainsKey(t, model.PromptData.ValueString(), "prompt")

	var metadata map[string]string
	diags = model.Metadata.ElementsAs(ctx, &metadata, false)
	if diags.HasError() {
		t.Fatalf("unexpected metadata diagnostics: %v", diags)
	}
	if metadata["owner"] != "ml" || metadata["tier"] != "1" {
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

func TestPopulateFunctionDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := FunctionDataSourceModel{}
	fn := &client.Function{
		ID: "function-2",
	}

	diags := populateFunctionDataSourceModel(ctx, &model, fn)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.Description.IsNull() {
		t.Fatalf("expected description to be null")
	}
	if !model.FunctionData.IsNull() {
		t.Fatalf("expected function_data to be null")
	}
	if !model.FunctionSchema.IsNull() {
		t.Fatalf("expected function_schema to be null")
	}
	if !model.Origin.IsNull() {
		t.Fatalf("expected origin to be null")
	}
	if !model.PromptData.IsNull() {
		t.Fatalf("expected prompt_data to be null")
	}
	if !model.Metadata.IsNull() {
		t.Fatalf("expected metadata to be null")
	}
	if !model.Tags.IsNull() {
		t.Fatalf("expected tags to be null")
	}
}

func TestPopulateFunctionDataSourceModel_JSONEncodeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := FunctionDataSourceModel{}
	fn := &client.Function{
		ID: "function-3",
		FunctionData: map[string]interface{}{
			"unmarshallable": func() {},
		},
	}

	diags := populateFunctionDataSourceModel(ctx, &model, fn)
	if !diags.HasError() {
		t.Fatal("expected diagnostics from marshal failure, got none")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Error Encoding function_data" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error summary %q, got %v", "Error Encoding function_data", diags)
	}

	if !model.FunctionData.IsNull() {
		t.Fatalf("expected function_data to be null when encoding fails")
	}
}

func assertJSONFieldContainsKey(t *testing.T, raw, key string) {
	t.Helper()

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if _, ok := decoded[key]; !ok {
		t.Fatalf("expected key %q in JSON payload: %s", key, raw)
	}
}
