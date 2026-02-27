package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListAISecretsOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListAISecretsOptions
		model       AISecretsDataSourceModel
	}{
		"builds all supported api-native filters": {
			model: AISecretsDataSourceModel{
				OrgName:       types.StringValue("example-org"),
				AISecretName:  types.StringValue("PROVIDER_OPENAI_CREDENTIAL"),
				StartingAfter: types.StringValue("ai-secret-1"),
				FilterIDs:     mustStringListValue(t, []string{"ai-secret-1", "ai-secret-2"}),
				AISecretTypes: mustStringListValue(t, []string{"openai", "anthropic"}),
				Limit:         types.Int64Value(10),
			},
			want: client.ListAISecretsOptions{
				OrgName:       "example-org",
				AISecretName:  "PROVIDER_OPENAI_CREDENTIAL",
				StartingAfter: "ai-secret-1",
				IDs:           []string{"ai-secret-1", "ai-secret-2"},
				AISecretTypes: []string{"openai", "anthropic"},
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: AISecretsDataSourceModel{
				StartingAfter: types.StringValue("ai-secret-1"),
				EndingBefore:  types.StringValue("ai-secret-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: AISecretsDataSourceModel{
				Limit: types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
		"rejects blank filter_ids entries": {
			model: AISecretsDataSourceModel{
				FilterIDs: mustStringListValue(t, []string{"ai-secret-1", " "}),
			},
			wantErrLike: "'filter_ids' cannot contain empty values",
		},
		"rejects blank ai_secret_types entries": {
			model: AISecretsDataSourceModel{
				AISecretTypes: mustStringListValue(t, []string{"openai", " "}),
			},
			wantErrLike: "'ai_secret_types' cannot contain empty values",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListAISecretsOptions(context.Background(), tc.model)
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
			if opts.OrgName != tc.want.OrgName ||
				opts.AISecretName != tc.want.AISecretName ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
			assertStringSlicesEqual(t, opts.IDs, tc.want.IDs)
			assertStringSlicesEqual(t, opts.AISecretTypes, tc.want.AISecretTypes)
		})
	}
}

func TestAISecretsDataSourceAISecretFromAISecret(t *testing.T) {
	t.Parallel()

	aiSecretModel, diags := aiSecretsDataSourceAISecretFromAISecret(context.Background(), &client.AISecret{
		ID:            "ai-secret-1",
		Name:          "PROVIDER_OPENAI_CREDENTIAL",
		OrgID:         "org-1",
		Type:          "openai",
		Metadata:      map[string]interface{}{"provider": "openai", "rotation_days": 30},
		PreviewSecret: "sk-***1234",
		Created:       "2026-02-26T00:00:00Z",
		UpdatedAt:     "2026-02-26T01:00:00Z",
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if aiSecretModel.ID.ValueString() != "ai-secret-1" {
		t.Fatalf("id mismatch: got=%q", aiSecretModel.ID.ValueString())
	}
	if aiSecretModel.Name.ValueString() != "PROVIDER_OPENAI_CREDENTIAL" {
		t.Fatalf("name mismatch: got=%q", aiSecretModel.Name.ValueString())
	}
	if aiSecretModel.OrgID.ValueString() != "org-1" {
		t.Fatalf("org_id mismatch: got=%q", aiSecretModel.OrgID.ValueString())
	}
	if aiSecretModel.Type.ValueString() != "openai" {
		t.Fatalf("type mismatch: got=%q", aiSecretModel.Type.ValueString())
	}
	if aiSecretModel.PreviewSecret.ValueString() != "sk-***1234" {
		t.Fatalf("preview_secret mismatch: got=%q", aiSecretModel.PreviewSecret.ValueString())
	}
	if aiSecretModel.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", aiSecretModel.Created.ValueString())
	}
	if aiSecretModel.UpdatedAt.ValueString() != "2026-02-26T01:00:00Z" {
		t.Fatalf("updated_at mismatch: got=%q", aiSecretModel.UpdatedAt.ValueString())
	}
}

func TestProviderDataSourcesIncludeAISecretPair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_ai_secret"]; !ok {
		t.Fatalf("expected braintrustdata_ai_secret to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_ai_secrets"]; !ok {
		t.Fatalf("expected braintrustdata_ai_secrets to be registered")
	}
}

func mustStringListValue(t *testing.T, values []string) types.List {
	t.Helper()

	listValue, diags := types.ListValueFrom(context.Background(), types.StringType, values)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating list value: %v", diags)
	}

	return listValue
}

func assertStringSlicesEqual(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice length mismatch: got=%v want=%v", got, want)
	}

	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("slice content mismatch: got=%v want=%v", got, want)
		}
	}
}
