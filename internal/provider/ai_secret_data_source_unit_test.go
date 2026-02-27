package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSelectSingleAISecretByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType  error
		aiSecretName string
		wantID       string
		aiSecrets    []client.AISecret
	}{
		"finds exact ai secret": {
			aiSecrets: []client.AISecret{
				{ID: "ai-secret-a", Name: "OTHER_KEY"},
				{ID: "ai-secret-b", Name: "PROVIDER_OPENAI_CREDENTIAL"},
			},
			aiSecretName: "PROVIDER_OPENAI_CREDENTIAL",
			wantID:       "ai-secret-b",
		},
		"returns not found when no exact match": {
			aiSecrets: []client.AISecret{
				{ID: "ai-secret-a", Name: "OTHER_KEY"},
			},
			aiSecretName: "PROVIDER_OPENAI_CREDENTIAL",
			wantErrType:  errAISecretNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			aiSecrets: []client.AISecret{
				{ID: "ai-secret-a", Name: "PROVIDER_OPENAI_CREDENTIAL"},
				{ID: "ai-secret-b", Name: "PROVIDER_OPENAI_CREDENTIAL"},
			},
			aiSecretName: "PROVIDER_OPENAI_CREDENTIAL",
			wantErrType:  errMultipleAISecretsFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			aiSecret, err := selectSingleAISecretByName(tc.aiSecrets, tc.aiSecretName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if aiSecret == nil {
				t.Fatalf("expected ai secret, got nil")
			}
			if aiSecret.ID != tc.wantID {
				t.Fatalf("expected ai secret ID %q, got %q", tc.wantID, aiSecret.ID)
			}
		})
	}
}

func TestPopulateAISecretDataSourceModel(t *testing.T) {
	t.Parallel()

	model := AISecretDataSourceModel{}
	aiSecret := &client.AISecret{
		ID:            "ai-secret-1",
		Name:          "PROVIDER_OPENAI_CREDENTIAL",
		OrgID:         "org-1",
		Type:          "openai",
		Metadata:      map[string]interface{}{"provider": "openai", "rotation_days": 30},
		PreviewSecret: "sk-***1234",
		Created:       "2026-02-26T00:00:00Z",
		UpdatedAt:     "2026-02-26T01:00:00Z",
	}

	diags := populateAISecretDataSourceModel(context.Background(), &model, aiSecret)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "ai-secret-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "PROVIDER_OPENAI_CREDENTIAL" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.OrgID.ValueString() != "org-1" {
		t.Fatalf("org_id mismatch: got=%q", model.OrgID.ValueString())
	}
	if model.Type.ValueString() != "openai" {
		t.Fatalf("type mismatch: got=%q", model.Type.ValueString())
	}
	if model.PreviewSecret.ValueString() != "sk-***1234" {
		t.Fatalf("preview_secret mismatch: got=%q", model.PreviewSecret.ValueString())
	}
	if model.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2026-02-26T01:00:00Z" {
		t.Fatalf("updated_at mismatch: got=%q", model.UpdatedAt.ValueString())
	}

	var metadata map[string]string
	metadataDiags := model.Metadata.ElementsAs(context.Background(), &metadata, false)
	if metadataDiags.HasError() {
		t.Fatalf("unexpected metadata diagnostics: %v", metadataDiags)
	}
	if metadata["provider"] != "openai" {
		t.Fatalf("metadata.provider mismatch: got=%q", metadata["provider"])
	}
	if metadata["rotation_days"] != "30" {
		t.Fatalf("metadata.rotation_days mismatch: got=%q", metadata["rotation_days"])
	}
}

func TestPopulateAISecretDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := AISecretDataSourceModel{}
	aiSecret := &client.AISecret{
		ID:   "ai-secret-2",
		Name: "PROVIDER_ANTHROPIC_CREDENTIAL",
	}

	diags := populateAISecretDataSourceModel(context.Background(), &model, aiSecret)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.OrgID.IsNull() {
		t.Fatalf("expected org_id to be null")
	}
	if !model.Type.IsNull() {
		t.Fatalf("expected type to be null")
	}
	if !model.Metadata.IsNull() {
		t.Fatalf("expected metadata to be null")
	}
	if !model.PreviewSecret.IsNull() {
		t.Fatalf("expected preview_secret to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("expected updated_at to be null")
	}
}

func TestBuildAISecretLookupInputs_TrimsValues(t *testing.T) {
	t.Parallel()

	inputs := buildAISecretLookupInputs(AISecretDataSourceModel{
		ID:           types.StringValue("  ai-secret-1  "),
		Name:         types.StringValue("  PROVIDER_OPENAI_CREDENTIAL  "),
		OrgName:      types.StringValue("  example-org  "),
		AISecretType: types.StringValue("  openai  "),
	})

	if inputs.id != "ai-secret-1" {
		t.Fatalf("id mismatch: got=%q", inputs.id)
	}
	if inputs.name != "PROVIDER_OPENAI_CREDENTIAL" {
		t.Fatalf("name mismatch: got=%q", inputs.name)
	}
	if inputs.orgName != "example-org" {
		t.Fatalf("org_name mismatch: got=%q", inputs.orgName)
	}
	if inputs.aiSecretType != "openai" {
		t.Fatalf("ai_secret_type mismatch: got=%q", inputs.aiSecretType)
	}
	if !inputs.hasID || !inputs.hasName || !inputs.hasOrgName || !inputs.hasAISecretType {
		t.Fatalf("expected all has* flags to be true: %+v", inputs)
	}
}

func TestValidateAISecretLookupInputs_WhitespaceName(t *testing.T) {
	t.Parallel()

	inputs := buildAISecretLookupInputs(AISecretDataSourceModel{
		Name:         types.StringValue("   "),
		OrgName:      types.StringValue("example-org"),
		AISecretType: types.StringValue("openai"),
	})

	diags := validateAISecretLookupInputs(inputs)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics for whitespace name")
	}

	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail(), "'name' must be provided and non-empty") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected lookup-name validation diagnostic, got %v", diags)
	}
}
