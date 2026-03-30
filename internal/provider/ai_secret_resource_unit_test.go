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

func TestBuildCreateAISecretRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		want        *client.CreateAISecretRequest
		wantErrLike string
		model       AISecretResourceModel
	}{
		"accepts valid request": {
			model: AISecretResourceModel{
				Name:     types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
				Type:     types.StringValue("openai"),
				Secret:   types.StringValue("sk-secret"),
				OrgName:  types.StringValue("test-org"),
				Metadata: types.MapValueMust(types.StringType, map[string]attr.Value{"provider": types.StringValue("openai")}),
			},
			want: &client.CreateAISecretRequest{
				Name:    "PROVIDER_OPENAI_CREDENTIAL",
				Type:    "openai",
				Secret:  "sk-secret",
				OrgName: "test-org",
				Metadata: map[string]interface{}{
					"provider": "openai",
				},
			},
		},
		"preserves surrounding whitespace in secret": {
			model: AISecretResourceModel{
				Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
				Secret: types.StringValue("  secret  "),
			},
			want: &client.CreateAISecretRequest{
				Name:   "PROVIDER_OPENAI_CREDENTIAL",
				Secret: "  secret  ",
			},
		},
		"rejects null secret": {
			model: AISecretResourceModel{
				Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
				Secret: types.StringNull(),
			},
			wantErrLike: "'secret' must be provided and non-empty when creating an AI secret.",
		},
		"rejects whitespace secret": {
			model: AISecretResourceModel{
				Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
				Secret: types.StringValue("   "),
			},
			wantErrLike: "'secret' must be provided and non-empty when creating an AI secret.",
		},
		"rejects whitespace name": {
			model: AISecretResourceModel{
				Name:   types.StringValue("   "),
				Secret: types.StringValue("sk-secret"),
			},
			wantErrLike: "'name' must be provided and non-empty when creating an AI secret.",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := buildCreateAISecretRequest(ctx, tc.model)
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
			if got.Name != tc.want.Name ||
				got.Type != tc.want.Type ||
				got.Secret != tc.want.Secret ||
				got.OrgName != tc.want.OrgName {
				t.Fatalf("request mismatch: got=%+v want=%+v", *got, *tc.want)
			}
			if tc.want.Metadata == nil {
				if got.Metadata != nil {
					t.Fatalf("expected metadata to be nil, got %v", got.Metadata)
				}
				return
			}
			if got.Metadata["provider"] != "openai" {
				t.Fatalf("metadata.provider mismatch: got=%v", got.Metadata["provider"])
			}
		})
	}
}

func TestBuildUpdateAISecretRequest_OnlyChangedFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := AISecretResourceModel{
		Name:     types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:     types.StringValue("openai"),
		Secret:   types.StringValue("old-secret"),
		Metadata: types.MapValueMust(types.StringType, map[string]attr.Value{"provider": types.StringValue("openai")}),
	}
	plan := AISecretResourceModel{
		Name:     types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:     types.StringValue("anthropic"),
		Secret:   types.StringValue("new-secret"),
		Metadata: types.MapNull(types.StringType),
	}

	req, diags := buildUpdateAISecretRequest(ctx, plan, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	//nolint:gosec // Test validates request serialization shape; no real secrets are used.
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	for _, key := range []string{"type", "secret", "metadata"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected %q in payload, got %v", key, payload)
		}
	}

	for _, key := range []string{"name", "org_name"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected %q to be omitted from payload, got %v", key, payload)
		}
	}
}

func TestBuildUpdateAISecretRequest_NormalizesNameAndType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := AISecretResourceModel{
		Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:   types.StringValue("openai"),
		Secret: types.StringValue("prior-secret"),
	}
	plan := AISecretResourceModel{
		Name:   types.StringValue("  PROVIDER_OPENAI_CREDENTIAL  "),
		Type:   types.StringValue("  anthropic  "),
		Secret: types.StringValue("prior-secret"),
	}

	req, diags := buildUpdateAISecretRequest(ctx, plan, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Name == nil || *req.Name != "PROVIDER_OPENAI_CREDENTIAL" {
		t.Fatalf("expected normalized name, got %#v", req.Name)
	}
	if req.Type == nil || *req.Type != "anthropic" {
		t.Fatalf("expected normalized type, got %#v", req.Type)
	}
}

func TestBuildUpdateAISecretRequest_OmitsSecretWhenNotSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := AISecretResourceModel{
		Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:   types.StringValue("openai"),
		Secret: types.StringValue("prior-secret"),
	}
	plan := AISecretResourceModel{
		Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:   types.StringValue("anthropic"),
		Secret: types.StringNull(),
	}

	req, diags := buildUpdateAISecretRequest(ctx, plan, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Secret != nil {
		t.Fatalf("expected secret to be omitted when not explicitly set, got %v", *req.Secret)
	}
	if req.Type == nil || *req.Type != "anthropic" {
		t.Fatalf("expected type update to be sent, got %#v", req.Type)
	}
}

func TestBuildUpdateAISecretRequest_RejectsWhitespaceSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := AISecretResourceModel{
		Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:   types.StringValue("openai"),
		Secret: types.StringValue("prior-secret"),
	}
	plan := AISecretResourceModel{
		Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:   types.StringValue("anthropic"),
		Secret: types.StringValue("   "),
	}

	_, diags := buildUpdateAISecretRequest(ctx, plan, state)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for whitespace-only secret, got none")
	}

	found := false
	for _, diag := range diags {
		if strings.Contains(diag.Detail(), "'secret' must be provided and non-empty when updating an AI secret.") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected whitespace secret validation error, got %v", diags)
	}
}

func TestBuildUpdateAISecretRequest_RejectsWhitespaceName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := AISecretResourceModel{
		Name:   types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
		Type:   types.StringValue("openai"),
		Secret: types.StringValue("prior-secret"),
	}
	plan := AISecretResourceModel{
		Name:   types.StringValue("   "),
		Type:   types.StringValue("anthropic"),
		Secret: types.StringValue("prior-secret"),
	}

	_, diags := buildUpdateAISecretRequest(ctx, plan, state)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for whitespace-only name, got none")
	}

	found := false
	for _, diag := range diags {
		if strings.Contains(diag.Detail(), "'name' must be provided and non-empty when updating an AI secret.") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected whitespace name validation error, got %v", diags)
	}
}

func TestSetAISecretResourceModel_PreservesSecretWhenAPIOmitsIt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := AISecretResourceModel{
		Secret: types.StringValue("prior-secret"),
	}
	aiSecret := &client.AISecret{
		ID:            "ai-secret-1",
		Name:          "PROVIDER_OPENAI_CREDENTIAL",
		Type:          "openai",
		OrgID:         "org-123",
		PreviewSecret: "********",
		Created:       "2026-03-16T18:00:00Z",
		UpdatedAt:     "2026-03-16T18:05:00Z",
		// API omits secret on read.
	}

	diags := setAISecretResourceModel(ctx, &model, aiSecret)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.Secret.ValueString(); got != "prior-secret" {
		t.Fatalf("expected secret to be preserved from prior state, got %q", got)
	}
	if got := model.Name.ValueString(); got != "PROVIDER_OPENAI_CREDENTIAL" {
		t.Fatalf("name mismatch: got=%q", got)
	}
	if got := model.Type.ValueString(); got != "openai" {
		t.Fatalf("type mismatch: got=%q", got)
	}
}

func TestSetAISecretResourceModel_LeavesSecretNullWhenAPIOmitsItAndNoPriorState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := AISecretResourceModel{}
	aiSecret := &client.AISecret{
		ID:            "ai-secret-1",
		Name:          "PROVIDER_OPENAI_CREDENTIAL",
		Type:          "openai",
		OrgID:         "org-123",
		PreviewSecret: "********",
		Created:       "2026-03-16T18:00:00Z",
		UpdatedAt:     "2026-03-16T18:05:00Z",
	}

	diags := setAISecretResourceModel(ctx, &model, aiSecret)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.Secret.IsNull() {
		t.Fatalf("expected secret to remain null when API omits it and no prior state exists, got %q", model.Secret.ValueString())
	}
}

func TestSetAISecretResourceModel_PreservesEmptyMetadataWhenAPIOmitsIt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := AISecretResourceModel{
		Metadata: types.MapValueMust(types.StringType, map[string]attr.Value{}),
	}
	aiSecret := &client.AISecret{
		ID:            "ai-secret-1",
		Name:          "PROVIDER_OPENAI_CREDENTIAL",
		Type:          "openai",
		OrgID:         "org-123",
		PreviewSecret: "********",
		Created:       "2026-03-16T18:00:00Z",
		UpdatedAt:     "2026-03-16T18:05:00Z",
	}

	diags := setAISecretResourceModel(ctx, &model, aiSecret)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.Metadata.IsNull() {
		t.Fatal("expected empty metadata map to remain empty, got null")
	}
	if model.Metadata.IsUnknown() {
		t.Fatal("expected empty metadata map to remain known, got unknown")
	}
	if got := len(model.Metadata.Elements()); got != 0 {
		t.Fatalf("expected empty metadata map, got %d elements", got)
	}
}

func TestSetAISecretResourceModel_LeavesMetadataNullWhenAPIOmitsItAndNoPriorState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := AISecretResourceModel{
		Metadata: types.MapNull(types.StringType),
	}
	aiSecret := &client.AISecret{
		ID:            "ai-secret-1",
		Name:          "PROVIDER_OPENAI_CREDENTIAL",
		Type:          "openai",
		OrgID:         "org-123",
		PreviewSecret: "********",
		Created:       "2026-03-16T18:00:00Z",
		UpdatedAt:     "2026-03-16T18:05:00Z",
	}

	diags := setAISecretResourceModel(ctx, &model, aiSecret)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.Metadata.IsNull() {
		t.Fatal("expected null metadata to remain null when API omits it")
	}
}

func TestAISecretResourceSchema_SecretSensitive(t *testing.T) {
	t.Parallel()

	r := NewAISecretResource().(*AISecretResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	attrValue, ok := schemaResp.Schema.Attributes["secret"]
	if !ok {
		t.Fatal("expected secret attribute in schema")
	}

	secretAttr, ok := attrValue.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected secret to be schema.StringAttribute, got %T", attrValue)
	}

	if !secretAttr.IsSensitive() {
		t.Fatal("expected secret to be sensitive")
	}
}

func TestAISecretResourceSchema_SecretIsOptionalComputedWithPlanModifier(t *testing.T) {
	t.Parallel()

	r := NewAISecretResource().(*AISecretResource)
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	attrValue, ok := schemaResp.Schema.Attributes["secret"]
	if !ok {
		t.Fatal("expected secret attribute in schema")
	}

	secretAttr, ok := attrValue.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected secret to be schema.StringAttribute, got %T", attrValue)
	}

	if !secretAttr.Optional {
		t.Fatal("expected secret to be optional")
	}
	if !secretAttr.Computed {
		t.Fatal("expected secret to be computed")
	}
	if !secretAttr.IsSensitive() {
		t.Fatal("expected secret to be sensitive")
	}
	if len(secretAttr.PlanModifiers) == 0 {
		t.Fatal("expected secret to use a state-preserving plan modifier")
	}
}

func TestProviderResourcesIncludeAISecret(t *testing.T) {
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

	if _, ok := resourceNames["braintrustdata_ai_secret"]; !ok {
		t.Fatalf("expected braintrustdata_ai_secret to be registered")
	}
}
