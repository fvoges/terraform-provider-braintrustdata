package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateViewRequest(t *testing.T) {
	t.Parallel()

	req, diags := buildCreateViewRequest(ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		ViewType:   types.StringValue("experiments"),
		Name:       types.StringValue("default"),
		Options:    types.StringValue(`{"freezeColumns":false,"viewType":"table"}`),
		ViewData:   types.StringValue(`{"search":{"filter":[],"match":[],"sort":[],"tag":[]}}`),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.ObjectID != "project-123" {
		t.Fatalf("expected object_id project-123, got %q", req.ObjectID)
	}
	if req.ObjectType != client.ACLObjectTypeProject {
		t.Fatalf("expected object_type project, got %q", req.ObjectType)
	}
	if req.ViewType != client.ViewTypeExperiments {
		t.Fatalf("expected view_type experiments, got %q", req.ViewType)
	}
	if got := req.Options["viewType"]; got != "table" {
		t.Fatalf("expected options.viewType table, got %v", got)
	}
}

func TestBuildCreateViewRequest_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, diags := buildCreateViewRequest(ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		ViewType:   types.StringValue("experiments"),
		Name:       types.StringValue("default"),
		Options:    types.StringValue(`{"freezeColumns":`),
	})
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid options JSON")
	}
}

func TestBuildUpdateViewRequest(t *testing.T) {
	t.Parallel()

	req, diags := buildUpdateViewRequest(ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringValue("updated"),
		Options:    types.StringValue(`{"freezeColumns":true,"viewType":"cards"}`),
		ViewData:   types.StringValue(`{"search":{"filter":[],"match":[{"key":"name","operator":"contains","value":"demo"}],"sort":[],"tag":[]}}`),
	}, ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringValue("default"),
		Options:    types.StringValue(`{"freezeColumns":false,"viewType":"table"}`),
		ViewData:   types.StringValue(`{"search":{"filter":[],"match":[],"sort":[],"tag":[]}}`),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.ObjectID != "project-123" || req.ObjectType != client.ACLObjectTypeProject {
		t.Fatalf("expected scope to be preserved in update request, got %+v", req)
	}
	if req.Name == nil || *req.Name != "updated" {
		t.Fatalf("expected updated name pointer, got %#v", req.Name)
	}
	options := decodeViewJSONRawMessage(t, req.Options)
	if got := options["freezeColumns"]; got != true {
		t.Fatalf("expected options.freezeColumns true, got %v", got)
	}
	if !hasViewUpdateChanges(req) {
		t.Fatal("expected update request to contain mutable changes")
	}
}

func TestBuildUpdateViewRequest_ExplicitNullClears(t *testing.T) {
	t.Parallel()

	req, diags := buildUpdateViewRequest(ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringValue("default"),
		Options:    types.StringNull(),
		ViewData:   types.StringNull(),
	}, ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringValue("default"),
		Options:    types.StringValue(`{"freezeColumns":false}`),
		ViewData:   types.StringValue(`{"search":{"match":[]}}`),
	})
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

	if got := string(payload["options"]); got != "null" {
		t.Fatalf("expected options null clear, got %s", got)
	}
	if got := string(payload["view_data"]); got != "null" {
		t.Fatalf("expected view_data null clear, got %s", got)
	}
	if !hasViewUpdateChanges(req) {
		t.Fatal("expected explicit clears to count as update changes")
	}
}

func TestBuildUpdateViewRequest_OmitsUnknownOptionalFields(t *testing.T) {
	t.Parallel()

	req, diags := buildUpdateViewRequest(ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringUnknown(),
		Options:    types.StringUnknown(),
		ViewData:   types.StringUnknown(),
	}, ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringValue("default"),
		Options:    types.StringValue(`{"freezeColumns":false}`),
		ViewData:   types.StringValue(`{"search":{"match":[]}}`),
	})
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

	for _, key := range []string{"name", "options", "view_data"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected %q to be omitted when plan value is unknown", key)
		}
	}
}

func TestBuildUpdateViewRequest_NullToNullNoop(t *testing.T) {
	t.Parallel()

	req, diags := buildUpdateViewRequest(ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringValue("default"),
		Options:    types.StringNull(),
		ViewData:   types.StringNull(),
	}, ViewResourceModel{
		ObjectID:   types.StringValue("project-123"),
		ObjectType: types.StringValue("project"),
		Name:       types.StringValue("default"),
		Options:    types.StringNull(),
		ViewData:   types.StringNull(),
	})
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

	for _, key := range []string{"name", "options", "view_data"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected %q to be omitted when state and plan are both null", key)
		}
	}
	if hasViewUpdateChanges(req) {
		t.Fatal("expected null-to-null transition to be a no-op")
	}
}

func TestSetViewResourceModel(t *testing.T) {
	t.Parallel()

	model := ViewResourceModel{
		Options:  types.StringValue(`{"freezeColumns":true,"viewType":"cards"}`),
		ViewData: types.StringValue(`{"search":{"filter":[],"match":[{"key":"name","operator":"contains","value":"demo"}],"sort":[],"tag":[]}}`),
	}

	diags := setViewResourceModel(context.Background(), &model, &client.View{
		ID:         "view-123",
		Name:       "updated",
		ObjectID:   "project-123",
		ObjectType: client.ACLObjectTypeProject,
		ViewType:   client.ViewTypeExperiments,
		Created:    "2026-03-13T13:00:00Z",
		UserID:     "user-123",
		Options: map[string]interface{}{
			"freezeColumns": true,
			"viewType":      "cards",
		},
		ViewData: map[string]interface{}{
			"search": map[string]interface{}{
				"filter": []interface{}{},
				"match": []interface{}{
					map[string]interface{}{
						"key":      "name",
						"operator": "contains",
						"value":    "demo",
					},
				},
				"sort": []interface{}{},
				"tag":  []interface{}{},
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "view-123" {
		t.Fatalf("expected id view-123, got %q", model.ID.ValueString())
	}
	if model.Options.ValueString() != `{"freezeColumns":true,"viewType":"cards"}` {
		t.Fatalf("expected canonical options JSON, got %s", model.Options.ValueString())
	}
	if model.ViewData.ValueString() != `{"search":{"filter":[],"match":[{"key":"name","operator":"contains","value":"demo"}],"sort":[],"tag":[]}}` {
		t.Fatalf("expected canonical view_data JSON, got %s", model.ViewData.ValueString())
	}
}

func TestParseViewImportID(t *testing.T) {
	t.Parallel()

	viewID, objectID, objectType, err := parseViewImportID("view-123,project-123,project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if viewID != "view-123" || objectID != "project-123" || objectType != "project" {
		t.Fatalf("unexpected parsed values: %q %q %q", viewID, objectID, objectType)
	}
}

func TestParseViewImportID_Invalid(t *testing.T) {
	t.Parallel()

	if _, _, _, err := parseViewImportID("view-123"); err == nil {
		t.Fatal("expected invalid import id error")
	}
}

func TestProviderResourcesIncludeView(t *testing.T) {
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

	if _, ok := resourceNames["braintrustdata_view"]; !ok {
		t.Fatalf("expected braintrustdata_view to be registered")
	}
}

func decodeViewJSONRawMessage(t *testing.T, raw *json.RawMessage) map[string]interface{} {
	t.Helper()

	if raw == nil {
		t.Fatal("expected JSON raw message, got nil")
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(*raw, &decoded); err != nil {
		t.Fatalf("unmarshal raw message: %v", err)
	}

	return decoded
}
