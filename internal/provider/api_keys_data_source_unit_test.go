package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListAPIKeysOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListAPIKeysOptions
		model       APIKeysDataSourceModel
	}{
		"builds all supported api-native filters": {
			model: APIKeysDataSourceModel{
				OrgName:       types.StringValue("example-org"),
				APIKeyName:    types.StringValue("service-key"),
				StartingAfter: types.StringValue("api-key-1"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListAPIKeysOptions{
				OrgName:       "example-org",
				APIKeyName:    "service-key",
				StartingAfter: "api-key-1",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: APIKeysDataSourceModel{
				StartingAfter: types.StringValue("api-key-1"),
				EndingBefore:  types.StringValue("api-key-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: APIKeysDataSourceModel{
				Limit: types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListAPIKeysOptions(tc.model)
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
				opts.APIKeyName != tc.want.APIKeyName ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestAPIKeysDataSourceAPIKeyFromAPIKey(t *testing.T) {
	t.Parallel()

	apiKeyModel := apiKeysDataSourceAPIKeyFromAPIKey(&client.APIKey{
		ID:          "api-key-1",
		Name:        "service-key",
		OrgID:       "org-1",
		PreviewName: "sk-1234",
		Created:     "2026-02-26T00:00:00Z",
		UserID:      "user-1",
		UserEmail:   "user@example.com",
	})

	if apiKeyModel.ID.ValueString() != "api-key-1" {
		t.Fatalf("id mismatch: got=%q", apiKeyModel.ID.ValueString())
	}
	if apiKeyModel.Name.ValueString() != "service-key" {
		t.Fatalf("name mismatch: got=%q", apiKeyModel.Name.ValueString())
	}
	if apiKeyModel.OrgID.ValueString() != "org-1" {
		t.Fatalf("org_id mismatch: got=%q", apiKeyModel.OrgID.ValueString())
	}
	if apiKeyModel.PreviewName.ValueString() != "sk-1234" {
		t.Fatalf("preview_name mismatch: got=%q", apiKeyModel.PreviewName.ValueString())
	}
	if apiKeyModel.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", apiKeyModel.Created.ValueString())
	}
	if apiKeyModel.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", apiKeyModel.UserID.ValueString())
	}
	if apiKeyModel.UserEmail.ValueString() != "user@example.com" {
		t.Fatalf("user_email mismatch: got=%q", apiKeyModel.UserEmail.ValueString())
	}
}

func TestProviderDataSourcesIncludeAPIKeyPair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_api_key"]; !ok {
		t.Fatalf("expected braintrustdata_api_key to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_api_keys"]; !ok {
		t.Fatalf("expected braintrustdata_api_keys to be registered")
	}
}
