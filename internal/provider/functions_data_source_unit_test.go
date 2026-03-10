package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListFunctionsOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		assert      func(t *testing.T, opts *client.ListFunctionsOptions)
		wantErrLike string
		model       FunctionsDataSourceModel
	}{
		"builds options and allows zero limit": {
			model: FunctionsDataSourceModel{
				ProjectID:     types.StringValue(" project-1 "),
				Name:          types.StringValue(" tool-a "),
				Slug:          types.StringValue("\n tool-a "),
				StartingAfter: types.StringValue(" function-10 "),
				Limit:         types.Int64Value(0),
			},
			assert: func(t *testing.T, opts *client.ListFunctionsOptions) {
				t.Helper()
				if opts.ProjectID != "project-1" {
					t.Fatalf("project_id mismatch: got=%q", opts.ProjectID)
				}
				if opts.FunctionName != "tool-a" {
					t.Fatalf("function_name mismatch: got=%q", opts.FunctionName)
				}
				if opts.Slug != "tool-a" {
					t.Fatalf("slug mismatch: got=%q", opts.Slug)
				}
				if opts.StartingAfter != "function-10" {
					t.Fatalf("starting_after mismatch: got=%q", opts.StartingAfter)
				}
				if opts.Limit == nil {
					t.Fatalf("expected limit to be set")
				}
				if *opts.Limit != 0 {
					t.Fatalf("expected limit=0, got %d", *opts.Limit)
				}
			},
		},
		"trims ending_before": {
			model: FunctionsDataSourceModel{
				EndingBefore: types.StringValue("  function-5\n"),
			},
			assert: func(t *testing.T, opts *client.ListFunctionsOptions) {
				t.Helper()
				if opts.EndingBefore != "function-5" {
					t.Fatalf("ending_before mismatch: got=%q", opts.EndingBefore)
				}
				if opts.Limit != nil {
					t.Fatalf("expected limit to be nil when unset")
				}
			},
		},
		"ignores whitespace-only optional filters": {
			model: FunctionsDataSourceModel{
				ProjectID: types.StringValue(" \n "),
				Name:      types.StringValue("  "),
				Slug:      types.StringValue("\t"),
			},
			assert: func(t *testing.T, opts *client.ListFunctionsOptions) {
				t.Helper()
				if opts.ProjectID != "" {
					t.Fatalf("expected empty project_id, got %q", opts.ProjectID)
				}
				if opts.FunctionName != "" {
					t.Fatalf("expected empty function_name, got %q", opts.FunctionName)
				}
				if opts.Slug != "" {
					t.Fatalf("expected empty slug, got %q", opts.Slug)
				}
			},
		},
		"rejects conflicting pagination": {
			model: FunctionsDataSourceModel{
				StartingAfter: types.StringValue("function-1"),
				EndingBefore:  types.StringValue("function-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects negative limit": {
			model: FunctionsDataSourceModel{
				Limit: types.Int64Value(-1),
			},
			wantErrLike: "'limit' must be greater than or equal to 0",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListFunctionsOptions(tc.model)
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
			if tc.assert != nil {
				tc.assert(t, opts)
			}
		})
	}
}

func TestFunctionListItemFromFunction(t *testing.T) {
	t.Parallel()

	model, diags := functionListItemFromFunction(context.Background(), &client.Function{
		ID:             "function-1",
		Name:           "tool-a",
		FunctionType:   "tool",
		XactID:         "xact-1",
		FunctionData:   map[string]interface{}{"runtime": "python"},
		FunctionSchema: map[string]interface{}{"type": "object"},
		Origin:         map[string]interface{}{"source": "api"},
		PromptData:     map[string]interface{}{"prompt": "hello"},
		Metadata:       map[string]interface{}{"owner": "ml"},
		Tags:           []string{"prod"},
	})
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
	if model.FunctionData.IsNull() {
		t.Fatalf("expected function_data to be populated")
	}
}

func TestFunctionListItemFromFunction_JSONEncodeError(t *testing.T) {
	t.Parallel()

	model, diags := functionListItemFromFunction(context.Background(), &client.Function{
		ID: "function-2",
		FunctionData: map[string]interface{}{
			"unmarshallable": make(chan int),
		},
	})
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

func TestProviderDataSourcesIncludeFunctionPair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_function"]; !ok {
		t.Fatalf("expected braintrustdata_function to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_functions"]; !ok {
		t.Fatalf("expected braintrustdata_functions to be registered")
	}
}
