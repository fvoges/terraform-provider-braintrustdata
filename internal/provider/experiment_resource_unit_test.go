package provider

import (
	"context"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestObjectToRepoInfoWithState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		input         types.Object
		name          string
		wantCommit    string
		wantState     repoInfoValueState
		wantNil       bool
		wantDirty     bool
		wantDirtyNil  bool
		wantHasErrors bool
	}{
		{
			name:      "null_object",
			input:     types.ObjectNull(experimentRepoInfoAttributeTypes),
			wantState: repoInfoValueStateNull,
			wantNil:   true,
		},
		{
			name:      "unknown_object",
			input:     types.ObjectUnknown(experimentRepoInfoAttributeTypes),
			wantState: repoInfoValueStateUnknown,
			wantNil:   true,
		},
		{
			name: "known_object",
			input: types.ObjectValueMust(
				experimentRepoInfoAttributeTypes,
				map[string]attr.Value{
					"commit":         types.StringValue("abc123"),
					"branch":         types.StringValue("main"),
					"tag":            types.StringNull(),
					"dirty":          types.BoolValue(true),
					"author_name":    types.StringValue("Jane"),
					"author_email":   types.StringValue("jane@example.com"),
					"commit_message": types.StringValue("message"),
					"commit_time":    types.StringValue("2026-02-18T12:00:00Z"),
					"git_diff":       types.StringNull(),
				},
			),
			wantState:    repoInfoValueStateKnown,
			wantCommit:   "abc123",
			wantDirty:    true,
			wantDirtyNil: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotState, diags := objectToRepoInfoWithState(ctx, tc.input)
			if diags.HasError() != tc.wantHasErrors {
				t.Fatalf("diagnostics mismatch: hasErrors=%v diags=%v", tc.wantHasErrors, diags)
			}
			if gotState != tc.wantState {
				t.Fatalf("state mismatch: got=%v want=%v", gotState, tc.wantState)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil repo info, got %#v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected repo info, got nil")
			}
			if got.Commit == nil || *got.Commit != tc.wantCommit {
				t.Fatalf("commit mismatch: got=%v want=%q", got.Commit, tc.wantCommit)
			}
			if tc.wantDirtyNil {
				if got.Dirty != nil {
					t.Fatalf("dirty mismatch: got=%v want=nil", *got.Dirty)
				}
			} else {
				if got.Dirty == nil || *got.Dirty != tc.wantDirty {
					t.Fatalf("dirty mismatch: got=%v want=%v", got.Dirty, tc.wantDirty)
				}
			}
		})
	}
}

func TestBuildExperimentUpdateRequestRepoInfoState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		repoInfo      types.Object
		name          string
		wantCommit    string
		wantNil       bool
		wantHasErrors bool
	}{
		{
			name:     "null_repo_info_is_omitted",
			repoInfo: types.ObjectNull(experimentRepoInfoAttributeTypes),
			wantNil:  true,
		},
		{
			name:     "unknown_repo_info_is_omitted",
			repoInfo: types.ObjectUnknown(experimentRepoInfoAttributeTypes),
			wantNil:  true,
		},
		{
			name: "known_repo_info_is_sent",
			repoInfo: types.ObjectValueMust(
				experimentRepoInfoAttributeTypes,
				map[string]attr.Value{
					"commit":         types.StringValue("def456"),
					"branch":         types.StringValue("feature/repo"),
					"tag":            types.StringNull(),
					"dirty":          types.BoolNull(),
					"author_name":    types.StringNull(),
					"author_email":   types.StringNull(),
					"commit_message": types.StringNull(),
					"commit_time":    types.StringNull(),
					"git_diff":       types.StringNull(),
				},
			),
			wantNil:    false,
			wantCommit: "def456",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, diags := buildExperimentUpdateRequest(ctx, ExperimentResourceModel{
				Name:        types.StringValue("experiment"),
				Description: types.StringValue("description"),
				Public:      types.BoolValue(true),
				Metadata:    types.MapNull(types.StringType),
				Tags:        types.SetNull(types.StringType),
				RepoInfo:    tc.repoInfo,
			})
			if diags.HasError() != tc.wantHasErrors {
				t.Fatalf("diagnostics mismatch: hasErrors=%v diags=%v", tc.wantHasErrors, diags)
			}

			if tc.wantNil {
				if req.RepoInfo != nil {
					t.Fatalf("expected nil repo_info, got %#v", req.RepoInfo)
				}
				return
			}

			if req.RepoInfo == nil {
				t.Fatal("expected repo_info to be sent")
			}
			if req.RepoInfo.Commit == nil || *req.RepoInfo.Commit != tc.wantCommit {
				t.Fatalf("repo_info.commit mismatch: got=%v want=%q", req.RepoInfo.Commit, tc.wantCommit)
			}
		})
	}
}

func TestExperimentResourceSchema_RepoInfoUseStateForUnknown(t *testing.T) {
	t.Parallel()

	r := NewExperimentResource().(*ExperimentResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	repoInfoAttr, ok := schemaResp.Schema.Attributes["repo_info"]
	if !ok {
		t.Fatal("expected repo_info attribute in schema")
	}

	nestedAttr, ok := repoInfoAttr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected repo_info to be schema.SingleNestedAttribute, got %T", repoInfoAttr)
	}

	modifiers := nestedAttr.ObjectPlanModifiers()
	if len(modifiers) == 0 {
		t.Fatal("expected repo_info to have object plan modifiers")
	}

	stateValue := types.ObjectValueMust(
		experimentRepoInfoAttributeTypes,
		map[string]attr.Value{
			"commit":         types.StringValue("abc123"),
			"branch":         types.StringValue("main"),
			"tag":            types.StringNull(),
			"dirty":          types.BoolValue(false),
			"author_name":    types.StringNull(),
			"author_email":   types.StringNull(),
			"commit_message": types.StringNull(),
			"commit_time":    types.StringNull(),
			"git_diff":       types.StringNull(),
		},
	)

	modifierReq := planmodifier.ObjectRequest{
		State: tfsdk.State{
			Raw: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"repo_info": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"commit":         tftypes.String,
								"branch":         tftypes.String,
								"tag":            tftypes.String,
								"dirty":          tftypes.Bool,
								"author_name":    tftypes.String,
								"author_email":   tftypes.String,
								"commit_message": tftypes.String,
								"commit_time":    tftypes.String,
								"git_diff":       tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"repo_info": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"commit":         tftypes.String,
								"branch":         tftypes.String,
								"tag":            tftypes.String,
								"dirty":          tftypes.Bool,
								"author_name":    tftypes.String,
								"author_email":   tftypes.String,
								"commit_message": tftypes.String,
								"commit_time":    tftypes.String,
								"git_diff":       tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"commit":         tftypes.NewValue(tftypes.String, "abc123"),
							"branch":         tftypes.NewValue(tftypes.String, "main"),
							"tag":            tftypes.NewValue(tftypes.String, nil),
							"dirty":          tftypes.NewValue(tftypes.Bool, false),
							"author_name":    tftypes.NewValue(tftypes.String, nil),
							"author_email":   tftypes.NewValue(tftypes.String, nil),
							"commit_message": tftypes.NewValue(tftypes.String, nil),
							"commit_time":    tftypes.NewValue(tftypes.String, nil),
							"git_diff":       tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
		},
		PlanValue:   types.ObjectUnknown(experimentRepoInfoAttributeTypes),
		ConfigValue: types.ObjectNull(experimentRepoInfoAttributeTypes),
		StateValue:  stateValue,
	}
	modifierResp := planmodifier.ObjectResponse{
		PlanValue: modifierReq.PlanValue,
	}

	modifiers[0].PlanModifyObject(context.Background(), modifierReq, &modifierResp)
	if modifierResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", modifierResp.Diagnostics)
	}
	if !modifierResp.PlanValue.Equal(stateValue) {
		t.Fatalf("expected plan value to use prior state, got %s", modifierResp.PlanValue.String())
	}
}

func TestExperimentResourceSchema_RepoInfoNestedAttributesOptionalComputed(t *testing.T) {
	t.Parallel()

	r := NewExperimentResource().(*ExperimentResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	repoInfoAttr, ok := schemaResp.Schema.Attributes["repo_info"]
	if !ok {
		t.Fatal("expected repo_info attribute in schema")
	}

	nestedAttr, ok := repoInfoAttr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected repo_info to be schema.SingleNestedAttribute, got %T", repoInfoAttr)
	}

	testCases := []struct {
		name   string
		isBool bool
	}{
		{name: "commit"},
		{name: "branch"},
		{name: "tag"},
		{name: "dirty", isBool: true},
		{name: "author_name"},
		{name: "author_email"},
		{name: "commit_message"},
		{name: "commit_time"},
		{name: "git_diff"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			attrValue, ok := nestedAttr.Attributes[tc.name]
			if !ok {
				t.Fatalf("expected %q attribute in repo_info schema", tc.name)
			}

			if tc.isBool {
				boolAttr, ok := attrValue.(schema.BoolAttribute)
				if !ok {
					t.Fatalf("expected %q to be schema.BoolAttribute, got %T", tc.name, attrValue)
				}
				if !boolAttr.IsOptional() || !boolAttr.IsComputed() {
					t.Fatalf("expected %q to be optional+computed", tc.name)
				}

				return
			}

			stringAttr, ok := attrValue.(schema.StringAttribute)
			if !ok {
				t.Fatalf("expected %q to be schema.StringAttribute, got %T", tc.name, attrValue)
			}
			if !stringAttr.IsOptional() || !stringAttr.IsComputed() {
				t.Fatalf("expected %q to be optional+computed", tc.name)
			}
		})
	}
}

func TestExperimentResourceSchema_RepoInfoGitDiffSensitive(t *testing.T) {
	t.Parallel()

	r := NewExperimentResource().(*ExperimentResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	repoInfoAttr, ok := schemaResp.Schema.Attributes["repo_info"]
	if !ok {
		t.Fatal("expected repo_info attribute in schema")
	}

	nestedAttr, ok := repoInfoAttr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected repo_info to be schema.SingleNestedAttribute, got %T", repoInfoAttr)
	}

	gitDiffAttr, ok := nestedAttr.Attributes["git_diff"]
	if !ok {
		t.Fatal("expected git_diff attribute in repo_info schema")
	}

	gitDiffStringAttr, ok := gitDiffAttr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected git_diff to be schema.StringAttribute, got %T", gitDiffAttr)
	}

	if !gitDiffStringAttr.IsSensitive() {
		t.Fatal("expected repo_info.git_diff to be sensitive")
	}
}

func TestIsRepoInfoConfigured(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		repoInfo types.Object
		name     string
		want     bool
	}{
		{
			name:     "null_is_not_configured",
			repoInfo: types.ObjectNull(experimentRepoInfoAttributeTypes),
			want:     false,
		},
		{
			name:     "unknown_is_not_configured",
			repoInfo: types.ObjectUnknown(experimentRepoInfoAttributeTypes),
			want:     false,
		},
		{
			name: "known_is_configured",
			repoInfo: types.ObjectValueMust(
				experimentRepoInfoAttributeTypes,
				map[string]attr.Value{
					"commit":         types.StringValue("abc123"),
					"branch":         types.StringNull(),
					"tag":            types.StringNull(),
					"dirty":          types.BoolNull(),
					"author_name":    types.StringNull(),
					"author_email":   types.StringNull(),
					"commit_message": types.StringNull(),
					"commit_time":    types.StringNull(),
					"git_diff":       types.StringNull(),
				},
			),
			want: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isRepoInfoConfigured(tc.repoInfo)
			if got != tc.want {
				t.Fatalf("configured mismatch: got=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestShouldPreserveRepoInfoState(t *testing.T) {
	t.Parallel()

	repoInfoFromAPI := &client.RepoInfo{}

	testCases := []struct {
		apiRepoInfo        *client.RepoInfo
		name               string
		repoInfoState      repoInfoValueState
		repoInfoConfigured bool
		want               bool
	}{
		{
			name:               "omitted_in_config_preserves_state_even_if_api_returns_repo_info",
			repoInfoConfigured: false,
			repoInfoState:      repoInfoValueStateKnown,
			apiRepoInfo:        &client.RepoInfo{},
			want:               true,
		},
		{
			name:               "configured_and_unknown_payload_with_nil_api_repo_info_preserves_state",
			repoInfoConfigured: true,
			repoInfoState:      repoInfoValueStateUnknown,
			apiRepoInfo:        nil,
			want:               true,
		},
		{
			name:               "configured_and_null_payload_with_nil_api_repo_info_preserves_state",
			repoInfoConfigured: true,
			repoInfoState:      repoInfoValueStateNull,
			apiRepoInfo:        nil,
			want:               true,
		},
		{
			name:               "configured_and_unknown_payload_with_api_repo_info_uses_api_value",
			repoInfoConfigured: true,
			repoInfoState:      repoInfoValueStateUnknown,
			apiRepoInfo:        repoInfoFromAPI,
			want:               false,
		},
		{
			name:               "configured_and_known_payload_uses_api_value",
			repoInfoConfigured: true,
			repoInfoState:      repoInfoValueStateKnown,
			apiRepoInfo:        repoInfoFromAPI,
			want:               false,
		},
		{
			name:               "configured_and_known_payload_with_nil_api_repo_info_uses_api_value",
			repoInfoConfigured: true,
			repoInfoState:      repoInfoValueStateKnown,
			apiRepoInfo:        nil,
			want:               false,
		},
		{
			name:               "configured_and_null_payload_with_api_repo_info_uses_api_value",
			repoInfoConfigured: true,
			repoInfoState:      repoInfoValueStateNull,
			apiRepoInfo:        repoInfoFromAPI,
			want:               false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := shouldPreserveRepoInfoState(tc.repoInfoConfigured, tc.repoInfoState, tc.apiRepoInfo)
			if got != tc.want {
				t.Fatalf("preserve decision mismatch: got=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestApplyRepoInfoConfigToUpdateRequest(t *testing.T) {
	t.Parallel()

	repoCommit := "abc123"
	testCases := []struct {
		name               string
		repoInfoConfigured bool
		expectNilRepoInfo  bool
	}{
		{
			name:               "omitted_in_config_removes_repo_info_from_update_request",
			repoInfoConfigured: false,
			expectNilRepoInfo:  true,
		},
		{
			name:               "configured_repo_info_is_left_unchanged_on_update_request",
			repoInfoConfigured: true,
			expectNilRepoInfo:  false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			originalRepoInfo := &client.RepoInfo{Commit: &repoCommit}
			updateReq := &client.UpdateExperimentRequest{
				RepoInfo: originalRepoInfo,
			}

			applyRepoInfoConfigToUpdateRequest(updateReq, tc.repoInfoConfigured)

			if tc.expectNilRepoInfo {
				if updateReq.RepoInfo != nil {
					t.Fatalf("expected repo_info to be omitted from update request, got %#v", updateReq.RepoInfo)
				}
				return
			}

			if updateReq.RepoInfo == nil {
				t.Fatalf("expected repo_info to be preserved on update request, got nil")
			}
			if updateReq.RepoInfo != originalRepoInfo {
				t.Fatalf("expected repo_info pointer to be unchanged; got %p want %p", updateReq.RepoInfo, originalRepoInfo)
			}
			if updateReq.RepoInfo.Commit == nil || *updateReq.RepoInfo.Commit != repoCommit {
				t.Fatalf("expected repo_info.commit to be %q, got %#v", repoCommit, updateReq.RepoInfo.Commit)
			}
		})
	}
}
