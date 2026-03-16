package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateTagRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := TagResourceModel{
		ProjectID:   types.StringValue("project-123"),
		Name:        types.StringValue("priority"),
		Description: types.StringValue("Primary tag"),
		Color:       types.StringValue("#ff0000"),
	}

	req, diags := buildCreateTagRequest(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.ProjectID != "project-123" {
		t.Fatalf("expected project_id project-123, got %q", req.ProjectID)
	}
	if req.Name != "priority" {
		t.Fatalf("expected name priority, got %q", req.Name)
	}
	if req.Description != "Primary tag" {
		t.Fatalf("expected description Primary tag, got %q", req.Description)
	}
	if req.Color != "#ff0000" {
		t.Fatalf("expected color #ff0000, got %q", req.Color)
	}
}

func TestBuildUpdateTagRequest_OnlyChangedFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := TagResourceModel{
		Name:        types.StringValue("priority"),
		Description: types.StringValue("Primary tag"),
		Color:       types.StringValue("#ff0000"),
	}
	plan := TagResourceModel{
		Name:        types.StringValue("priority-updated"),
		Description: types.StringValue("Updated tag"),
		Color:       types.StringValue("#00ff00"),
	}

	req, diags := buildUpdateTagRequest(ctx, plan, state)
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

	for _, key := range []string{"name", "description", "color"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected %q in payload, got %v", key, payload)
		}
	}

	if len(payload) != 3 {
		t.Fatalf("expected exactly 3 updated fields, got %v", payload)
	}
}

func TestBuildUpdateTagRequest_DescriptionNullNoOp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := TagResourceModel{
		Name:        types.StringValue("priority"),
		Description: types.StringValue("Primary tag"),
	}
	plan := TagResourceModel{
		Name:        types.StringValue("priority"),
		Description: types.StringNull(),
	}

	req, diags := buildUpdateTagRequest(ctx, plan, state)
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

	if len(payload) != 0 {
		t.Fatalf("expected no-op update payload for null description, got %v", payload)
	}
}

func TestBuildUpdateTagRequest_ColorNullNoOp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := TagResourceModel{
		Name:  types.StringValue("priority"),
		Color: types.StringValue("#ff0000"),
	}
	plan := TagResourceModel{
		Name:  types.StringValue("priority"),
		Color: types.StringNull(),
	}

	req, diags := buildUpdateTagRequest(ctx, plan, state)
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

	if len(payload) != 0 {
		t.Fatalf("expected no-op update payload for null color, got %v", payload)
	}
}

func TestSetTagResourceModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	position := "0|hzzzz:"
	tag := &client.Tag{
		ID:          "tag-123",
		ProjectID:   "project-123",
		Name:        "priority",
		Description: "Primary tag",
		Color:       "#ff0000",
		UserID:      "user-1",
		Created:     "2026-03-12T10:00:00Z",
		Position:    &position,
	}

	model := TagResourceModel{}
	setTagResourceModel(ctx, &model, tag)

	if model.ID.ValueString() != "tag-123" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.ProjectID.ValueString() != "project-123" {
		t.Fatalf("project_id mismatch: got=%q", model.ProjectID.ValueString())
	}
	if model.Name.ValueString() != "priority" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Primary tag" {
		t.Fatalf("description mismatch: got=%q", model.Description.ValueString())
	}
	if model.Color.ValueString() != "#ff0000" {
		t.Fatalf("color mismatch: got=%q", model.Color.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}
	if model.Created.ValueString() != "2026-03-12T10:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.Position.ValueString() != "0|hzzzz:" {
		t.Fatalf("position mismatch: got=%q", model.Position.ValueString())
	}
}

func TestSetTagResourceModel_Nullables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tag := &client.Tag{
		ID:        "tag-2",
		ProjectID: "project-2",
		Name:      "staging",
	}

	model := TagResourceModel{}
	setTagResourceModel(ctx, &model, tag)

	if !model.Description.IsNull() {
		t.Fatal("expected description to be null")
	}
	if !model.Color.IsNull() {
		t.Fatal("expected color to be null")
	}
	if !model.UserID.IsNull() {
		t.Fatal("expected user_id to be null")
	}
	if !model.Created.IsNull() {
		t.Fatal("expected created to be null")
	}
	if !model.Position.IsNull() {
		t.Fatal("expected position to be null")
	}
}

func TestTagResourceSchema_PositionComputedOnly(t *testing.T) {
	t.Parallel()

	r := NewTagResource().(*TagResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	attrValue, ok := schemaResp.Schema.Attributes["position"]
	if !ok {
		t.Fatal("expected position attribute in schema")
	}

	positionAttr, ok := attrValue.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected position to be schema.StringAttribute, got %T", attrValue)
	}

	if !positionAttr.IsComputed() {
		t.Fatal("expected position to be computed")
	}
	if positionAttr.IsOptional() {
		t.Fatal("expected position to not be optional")
	}
	if positionAttr.IsRequired() {
		t.Fatal("expected position to not be required")
	}
}

func TestProviderResourcesIncludeTag(t *testing.T) {
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

	if _, ok := resourceNames["braintrustdata_tag"]; !ok {
		t.Fatalf("expected braintrustdata_tag to be registered")
	}
}
