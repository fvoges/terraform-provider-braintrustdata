package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateFunctionRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := FunctionResourceModel{
		ProjectID:      types.StringValue("project-123"),
		Name:           types.StringValue("support-tool"),
		Slug:           types.StringValue("support-tool"),
		Description:    types.StringValue("Support workflow tool"),
		FunctionType:   types.StringValue("tool"),
		FunctionData:   types.StringValue(`{"runtime":"node","entrypoint":"index.ts"}`),
		FunctionSchema: types.StringValue(`{"type":"object"}`),
		PromptData:     types.StringValue(`{"prompt":{"type":"chat"}}`),
		Metadata: types.MapValueMust(types.StringType, map[string]attr.Value{
			"owner": types.StringValue("ml"),
		}),
		Tags: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("prod"),
			types.StringValue("support"),
		}),
	}

	req, diags := buildCreateFunctionRequest(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.ProjectID != "project-123" {
		t.Fatalf("expected project_id project-123, got %q", req.ProjectID)
	}
	if req.FunctionType != "tool" {
		t.Fatalf("expected function_type tool, got %q", req.FunctionType)
	}
	if got := req.Metadata["owner"]; got != "ml" {
		t.Fatalf("expected metadata owner=ml, got %v", got)
	}
	if len(req.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(req.Tags))
	}

	functionData, ok := req.FunctionData.(map[string]interface{})
	if !ok {
		t.Fatalf("expected function_data object, got %T", req.FunctionData)
	}
	if functionData["runtime"] != "node" {
		t.Fatalf("expected function_data.runtime=node, got %v", functionData["runtime"])
	}
}

func TestBuildCreateFunctionRequest_InvalidFunctionData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := FunctionResourceModel{
		ProjectID:    types.StringValue("project-123"),
		Name:         types.StringValue("support-tool"),
		FunctionData: types.StringValue(`{"runtime":`),
	}

	_, diags := buildCreateFunctionRequest(ctx, model)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid function_data JSON")
	}
}

func TestBuildUpdateFunctionRequest_OnlyChangedFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := FunctionResourceModel{
		Name:           types.StringValue("support-tool"),
		Slug:           types.StringValue("support-tool"),
		Description:    types.StringValue("Support workflow tool"),
		FunctionType:   types.StringValue("tool"),
		FunctionData:   types.StringValue(`{"runtime":"node"}`),
		FunctionSchema: types.StringValue(`{"type":"object"}`),
		PromptData:     types.StringValue(`{"prompt":{"type":"chat"}}`),
		Metadata: types.MapValueMust(types.StringType, map[string]attr.Value{
			"owner": types.StringValue("ml"),
		}),
		Tags: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("prod"),
		}),
	}
	plan := FunctionResourceModel{
		Name:           types.StringValue("support-tool-v2"),
		Slug:           types.StringNull(),
		Description:    types.StringNull(),
		FunctionType:   types.StringValue("tool"),
		FunctionData:   types.StringValue(`{"runtime":"python"}`),
		FunctionSchema: types.StringUnknown(),
		PromptData:     types.StringUnknown(),
		Metadata: types.MapValueMust(types.StringType, map[string]attr.Value{
			"owner": types.StringValue("platform"),
		}),
		Tags: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("prod"),
			types.StringValue("v2"),
		}),
	}

	req, diags := buildUpdateFunctionRequest(ctx, plan, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	for _, key := range []string{"name", "function_data", "metadata", "tags"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected %q in payload, got %v", key, payload)
		}
	}

	for _, key := range []string{"slug", "description", "function_schema", "prompt_data", "function_type"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected %q to be omitted, got %v", key, payload)
		}
	}
}

func TestSetFunctionResourceModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := FunctionResourceModel{}
	function := &client.Function{
		ID:             "function-123",
		ProjectID:      "project-123",
		Name:           "support-tool",
		Slug:           "support-tool",
		Description:    "Support workflow tool",
		FunctionType:   "tool",
		XactID:         "xact-1",
		LogID:          "log-1",
		Created:        "2026-03-12T10:00:00Z",
		OrgID:          "org-123",
		FunctionData:   map[string]interface{}{"runtime": "node"},
		FunctionSchema: map[string]interface{}{"type": "object"},
		PromptData:     map[string]interface{}{"prompt": map[string]interface{}{"type": "chat"}},
		Origin:         map[string]interface{}{"source": "api"},
		Metadata: map[string]interface{}{
			"owner": "ml",
			"tier":  "prod",
		},
		Tags: []string{"prod", "support"},
	}

	diags := setFunctionResourceModel(ctx, &model, function)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "function-123" {
		t.Fatalf("expected id function-123, got %q", model.ID.ValueString())
	}
	if model.ProjectID.ValueString() != "project-123" {
		t.Fatalf("expected project_id project-123, got %q", model.ProjectID.ValueString())
	}
	if model.FunctionType.ValueString() != "tool" {
		t.Fatalf("expected function_type tool, got %q", model.FunctionType.ValueString())
	}
	if model.FunctionData.IsNull() {
		t.Fatal("expected function_data to be set")
	}
	if model.Origin.IsNull() {
		t.Fatal("expected origin to be set")
	}
	if model.Tags.IsNull() {
		t.Fatal("expected tags to be set")
	}
}

func TestFunctionResourceSchema_ProjectIDRequiresReplace(t *testing.T) {
	t.Parallel()

	r := NewFunctionResource().(*FunctionResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	attrValue, ok := schemaResp.Schema.Attributes["project_id"]
	if !ok {
		t.Fatal("expected project_id attribute in schema")
	}

	projectIDAttr, ok := attrValue.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected project_id to be schema.StringAttribute, got %T", attrValue)
	}

	if !projectIDAttr.IsRequired() {
		t.Fatal("expected project_id to be required")
	}
	if len(projectIDAttr.PlanModifiers) == 0 {
		t.Fatal("expected project_id to have plan modifiers")
	}
}

func TestProviderResourcesIncludeFunction(t *testing.T) {
	t.Parallel()

	p, ok := New("test")().(*BraintrustProvider)
	if !ok {
		t.Fatalf("expected *BraintrustProvider")
	}

	resourceFactories := p.Resources(context.Background())
	resourceNames := make(map[string]struct{}, len(resourceFactories))

	for _, factory := range resourceFactories {
		r := factory()
		resp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), resource.MetadataRequest{
			ProviderTypeName: "braintrustdata",
		}, resp)
		resourceNames[resp.TypeName] = struct{}{}
	}

	if _, ok := resourceNames["braintrustdata_function"]; !ok {
		t.Fatalf("expected braintrustdata_function to be registered")
	}
}
