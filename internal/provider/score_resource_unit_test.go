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

func TestBuildCreateScoreRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := ScoreResourceModel{
		ProjectID:   types.StringValue("project-123"),
		Name:        types.StringValue("quality"),
		ScoreType:   types.StringValue("categorical"),
		Description: types.StringValue("Quality score"),
		Categories:  types.StringValue(`["good","bad"]`),
		Config:      types.StringValue(`{"max":5}`),
	}

	req, diags := buildCreateScoreRequest(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.ProjectID != "project-123" {
		t.Fatalf("expected project_id project-123, got %q", req.ProjectID)
	}
	if req.Name != "quality" {
		t.Fatalf("expected name quality, got %q", req.Name)
	}
	if req.ScoreType != "categorical" {
		t.Fatalf("expected score_type categorical, got %q", req.ScoreType)
	}

	categories, ok := req.Categories.([]interface{})
	if !ok {
		t.Fatalf("expected categories slice, got %T", req.Categories)
	}
	if len(categories) != 2 || categories[0] != "good" || categories[1] != "bad" {
		t.Fatalf("unexpected categories: %v", categories)
	}

	config, ok := req.Config.(map[string]interface{})
	if !ok {
		t.Fatalf("expected config object, got %T", req.Config)
	}
	if config["max"] != float64(5) {
		t.Fatalf("unexpected config: %v", config)
	}
}

func TestBuildCreateScoreRequest_InvalidCategories(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := ScoreResourceModel{
		ProjectID:  types.StringValue("project-123"),
		Name:       types.StringValue("quality"),
		ScoreType:  types.StringValue("categorical"),
		Categories: types.StringValue(`["good"`),
	}

	_, diags := buildCreateScoreRequest(ctx, model)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid categories JSON")
	}
}

func TestBuildUpdateScoreRequest_OnlyChangedFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := ScoreResourceModel{
		Name:        types.StringValue("quality"),
		ScoreType:   types.StringValue("categorical"),
		Description: types.StringValue("Initial quality score"),
		Categories:  types.StringValue(`["good","bad"]`),
		Config:      types.StringValue(`{"max":5}`),
	}
	plan := ScoreResourceModel{
		Name:        types.StringValue("quality-v2"),
		ScoreType:   types.StringUnknown(),
		Description: types.StringValue("Updated quality score"),
		Categories:  types.StringValue(`["great","bad"]`),
		Config:      types.StringValue(`{"max":10}`),
	}

	req, diags := buildUpdateScoreRequest(ctx, plan, state)
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

	for _, key := range []string{"name", "description", "categories", "config"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected %q in payload, got %v", key, payload)
		}
	}
	if _, ok := payload["score_type"]; ok {
		t.Fatalf("expected score_type to be omitted, got %v", payload)
	}
}

func TestBuildUpdateScoreRequest_ClearsJSONFieldsWithExplicitNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := ScoreResourceModel{
		Categories: types.StringValue(`["good","bad"]`),
		Config:     types.StringValue(`{"max":5}`),
	}
	plan := ScoreResourceModel{
		Categories: types.StringNull(),
		Config:     types.StringNull(),
	}

	req, diags := buildUpdateScoreRequest(ctx, plan, state)
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

	for _, key := range []string{"categories", "config"} {
		raw, ok := payload[key]
		if !ok {
			t.Fatalf("expected %q in payload, got %v", key, payload)
		}
		if string(raw) != "null" {
			t.Fatalf("expected %q to be null, got %s", key, raw)
		}
	}
}

func TestBuildUpdateScoreRequest_SemanticallyEquivalentJSONNoOp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := ScoreResourceModel{
		Categories: types.StringValue(`["good","bad"]`),
		Config:     types.StringValue(`{"max":5,"nested":{"a":1,"b":2}}`),
	}
	plan := ScoreResourceModel{
		Categories: types.StringValue("[\n  \"good\",\n  \"bad\"\n]"),
		Config:     types.StringValue("{\n  \"nested\": {\"b\": 2, \"a\": 1},\n  \"max\": 5\n}"),
	}

	req, diags := buildUpdateScoreRequest(ctx, plan, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.Categories != nil {
		t.Fatalf("expected categories to be omitted for semantic no-op, got %v", *req.Categories)
	}
	if req.Config != nil {
		t.Fatalf("expected config to be omitted for semantic no-op, got %v", *req.Config)
	}
	if hasScoreUpdateChanges(req) {
		t.Fatalf("expected no update changes, got %+v", req)
	}
}

func TestSetScoreResourceModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	position := "0|hzzzz:"
	model := ScoreResourceModel{}
	score := &client.ProjectScore{
		ID:          "score-123",
		ProjectID:   "project-123",
		Name:        "quality",
		ScoreType:   "categorical",
		Description: "Quality score",
		Categories:  []string{"good", "bad"},
		Config:      map[string]interface{}{"max": float64(5)},
		Position:    &position,
		UserID:      "user-1",
		Created:     "2026-03-12T10:00:00Z",
	}

	diags := setScoreResourceModel(ctx, &model, score)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "score-123" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.ProjectID.ValueString() != "project-123" {
		t.Fatalf("project_id mismatch: got=%q", model.ProjectID.ValueString())
	}
	if model.Name.ValueString() != "quality" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.ScoreType.ValueString() != "categorical" {
		t.Fatalf("score_type mismatch: got=%q", model.ScoreType.ValueString())
	}
	if model.Position.ValueString() != "0|hzzzz:" {
		t.Fatalf("position mismatch: got=%q", model.Position.ValueString())
	}
	if model.Categories.ValueString() != `["good","bad"]` {
		t.Fatalf("categories mismatch: got=%q", model.Categories.ValueString())
	}
	if model.Config.ValueString() != `{"max":5}` {
		t.Fatalf("config mismatch: got=%q", model.Config.ValueString())
	}
}

func TestScoreResourceSchema_PositionComputedOnly(t *testing.T) {
	t.Parallel()

	r := NewScoreResource().(*ScoreResource)
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

func TestProviderResourcesIncludeScore(t *testing.T) {
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

	if _, ok := resourceNames["braintrustdata_score"]; !ok {
		t.Fatalf("expected braintrustdata_score to be registered")
	}
}

func TestSetScoreResourceModel_Nullables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := ScoreResourceModel{
		Categories: types.StringNull(),
		Config:     types.StringNull(),
	}
	score := &client.ProjectScore{
		ID:        "score-2",
		ProjectID: "project-2",
		Name:      "latency",
	}

	diags := setScoreResourceModel(ctx, &model, score)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.Description.IsNull() {
		t.Fatalf("expected description to be null")
	}
	if !model.Categories.IsNull() {
		t.Fatalf("expected categories to be null")
	}
	if !model.Config.IsNull() {
		t.Fatalf("expected config to be null")
	}
	if !model.Position.IsNull() {
		t.Fatalf("expected position to be null")
	}
}

func TestBuildCreateScoreRequest_PreservesNullOptionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := ScoreResourceModel{
		ProjectID:   types.StringValue("project-123"),
		Name:        types.StringValue("quality"),
		ScoreType:   types.StringValue("categorical"),
		Description: types.StringNull(),
		Categories:  types.StringNull(),
		Config:      types.StringNull(),
	}

	req, diags := buildCreateScoreRequest(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Description != "" {
		t.Fatalf("expected empty description, got %q", req.Description)
	}
	if req.Categories != nil {
		t.Fatalf("expected nil categories, got %v", req.Categories)
	}
	if req.Config != nil {
		t.Fatalf("expected nil config, got %v", req.Config)
	}
}

func TestBuildUpdateScoreRequest_InvalidConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := ScoreResourceModel{
		Config: types.StringValue(`{"max":5}`),
	}
	plan := ScoreResourceModel{
		Config: types.StringValue(`{"max":`),
	}

	_, diags := buildUpdateScoreRequest(ctx, plan, state)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid config JSON")
	}
}

func TestSetScoreResourceModel_PreservesExplicitEmptyJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := ScoreResourceModel{
		Categories: types.StringValue(`[]`),
		Config:     types.StringValue(`{}`),
	}
	score := &client.ProjectScore{
		ID:         "score-3",
		ProjectID:  "project-3",
		Name:       "quality",
		Categories: []string{},
		Config:     map[string]interface{}{},
	}

	diags := setScoreResourceModel(ctx, &model, score)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.Categories.ValueString() != `[]` {
		t.Fatalf("expected empty categories array, got %q", model.Categories.ValueString())
	}
	if model.Config.ValueString() != `{}` {
		t.Fatalf("expected empty config object, got %q", model.Config.ValueString())
	}
}
