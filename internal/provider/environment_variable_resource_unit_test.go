package provider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateEnvironmentVariableRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		want        *client.CreateEnvironmentVariableRequest
		wantErrLike string
		model       EnvironmentVariableResourceModel
	}{
		"accepts valid request": {
			model: EnvironmentVariableResourceModel{
				ObjectType:     types.StringValue("project"),
				ObjectID:       types.StringValue("project-123"),
				Name:           types.StringValue("OPENAI_API_KEY"),
				Value:          types.StringValue("sk-secret"),
				Description:    types.StringValue("Used by prompts"),
				Metadata:       types.MapValueMust(types.StringType, map[string]attr.Value{"owner": types.StringValue("ml-platform")}),
				SecretType:     types.StringValue("text"),
				SecretCategory: types.StringValue("api"),
			},
			want: &client.CreateEnvironmentVariableRequest{
				ObjectType: "project",
				ObjectID:   "project-123",
				Name:       "OPENAI_API_KEY",
				Value:      "sk-secret",
				Metadata: map[string]interface{}{
					"owner": "ml-platform",
				},
				SecretType:     "text",
				SecretCategory: "api",
			},
		},
		"preserves surrounding whitespace in value": {
			model: EnvironmentVariableResourceModel{
				ObjectType: types.StringValue("project"),
				ObjectID:   types.StringValue("project-123"),
				Name:       types.StringValue("OPENAI_API_KEY"),
				Value:      types.StringValue("  secret  "),
			},
			want: &client.CreateEnvironmentVariableRequest{
				ObjectType: "project",
				ObjectID:   "project-123",
				Name:       "OPENAI_API_KEY",
				Value:      "  secret  ",
			},
		},
		"rejects null value": {
			model: EnvironmentVariableResourceModel{
				ObjectType: types.StringValue("project"),
				ObjectID:   types.StringValue("project-123"),
				Name:       types.StringValue("OPENAI_API_KEY"),
				Value:      types.StringNull(),
			},
			wantErrLike: "'value' must be provided and non-empty",
		},
		"rejects whitespace value": {
			model: EnvironmentVariableResourceModel{
				ObjectType: types.StringValue("project"),
				ObjectID:   types.StringValue("project-123"),
				Name:       types.StringValue("OPENAI_API_KEY"),
				Value:      types.StringValue("   "),
			},
			wantErrLike: "'value' must be provided and non-empty",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := buildCreateEnvironmentVariableRequest(ctx, tc.model)
			if tc.wantErrLike != "" {
				if !diags.HasError() {
					t.Fatalf("expected diagnostics containing %q, got none", tc.wantErrLike)
				}

				found := false
				for _, diag := range diags {
					if strings.Contains(diag.Detail(), tc.wantErrLike) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected diagnostics containing %q, got %v", tc.wantErrLike, diags)
				}

				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got == nil {
				t.Fatalf("expected create request, got nil")
			}
			if got.ObjectType != tc.want.ObjectType ||
				got.ObjectID != tc.want.ObjectID ||
				got.Name != tc.want.Name ||
				got.Value != tc.want.Value ||
				got.SecretType != tc.want.SecretType ||
				got.SecretCategory != tc.want.SecretCategory {
				t.Fatalf("request mismatch: got=%+v want=%+v", *got, *tc.want)
			}
			if tc.want.Metadata == nil {
				if got.Metadata != nil {
					t.Fatalf("expected metadata to be nil, got %v", got.Metadata)
				}
				return
			}
			if got.Metadata["owner"] != "ml-platform" {
				t.Fatalf("metadata.owner mismatch: got=%v", got.Metadata["owner"])
			}
		})
	}
}

func TestBuildUpdateEnvironmentVariableRequest_OnlyChangedFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := EnvironmentVariableResourceModel{
		Name:           types.StringValue("OPENAI_API_KEY"),
		Value:          types.StringValue("old-value"),
		Description:    types.StringValue("old description"),
		Metadata:       types.MapValueMust(types.StringType, map[string]attr.Value{"owner": types.StringValue("ml-platform")}),
		SecretType:     types.StringValue("text"),
		SecretCategory: types.StringValue("api"),
	}
	plan := EnvironmentVariableResourceModel{
		Name:           types.StringValue("OPENAI_API_KEY"),
		Value:          types.StringValue("new-value"),
		Description:    types.StringNull(),
		Metadata:       types.MapNull(types.StringType),
		SecretType:     types.StringValue("text"),
		SecretCategory: types.StringUnknown(),
	}

	req, diags := buildUpdateEnvironmentVariableRequest(ctx, plan, state)
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

	for _, key := range []string{"value", "metadata"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected %q in payload, got %v", key, payload)
		}
	}

	for _, key := range []string{"name", "description", "secret_type", "secret_category"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected %q to be omitted from payload, got %v", key, payload)
		}
	}
}

func TestEnvironmentVariableResourceSchema_DescriptionComputedOnly(t *testing.T) {
	t.Parallel()

	r := NewEnvironmentVariableResource().(*EnvironmentVariableResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	attrValue, ok := schemaResp.Schema.Attributes["description"]
	if !ok {
		t.Fatal("expected description attribute in schema")
	}

	descriptionAttr, ok := attrValue.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected description to be schema.StringAttribute, got %T", attrValue)
	}

	if !descriptionAttr.IsComputed() {
		t.Fatal("expected description to be computed")
	}
	if descriptionAttr.IsOptional() {
		t.Fatal("expected description to not be optional")
	}
	if descriptionAttr.IsRequired() {
		t.Fatal("expected description to not be required")
	}
}

func TestSetEnvironmentVariableResourceModel_PreservesValueWhenAPIOmitsIt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := EnvironmentVariableResourceModel{
		Value: types.StringValue("prior-secret"),
	}
	envVar := &client.EnvironmentVariable{
		ID:             "env-var-1",
		ObjectType:     "project",
		ObjectID:       "project-1",
		Name:           "OPENAI_API_KEY",
		Description:    "Used by prompts",
		SecretType:     "text",
		SecretCategory: "api",
		Created:        "2024-01-15T10:30:00Z",
		Used:           true,
		// API omits Value on read.
	}

	diags := setEnvironmentVariableResourceModel(ctx, &model, envVar)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.Value.ValueString(); got != "prior-secret" {
		t.Fatalf("expected value to be preserved from prior state, got %q", got)
	}
	if got := model.Name.ValueString(); got != "OPENAI_API_KEY" {
		t.Fatalf("name mismatch: got=%q", got)
	}
	if got := model.ObjectType.ValueString(); got != "project" {
		t.Fatalf("object_type mismatch: got=%q", got)
	}
}

func TestProviderResourcesIncludeEnvironmentVariable(t *testing.T) {
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

	if _, ok := resourceNames["braintrustdata_environment_variable"]; !ok {
		t.Fatalf("expected braintrustdata_environment_variable to be registered")
	}
}
