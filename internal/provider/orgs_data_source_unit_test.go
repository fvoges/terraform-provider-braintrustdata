package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListOrganizationsOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListOrganizationsOptions
		model       OrgsDataSourceModel
	}{
		"builds all supported API-native filters": {
			model: OrgsDataSourceModel{
				OrgName:       types.StringValue("example-org"),
				StartingAfter: types.StringValue("org-1"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListOrganizationsOptions{
				OrgName:       "example-org",
				StartingAfter: "org-1",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: OrgsDataSourceModel{
				StartingAfter: types.StringValue("org-1"),
				EndingBefore:  types.StringValue("org-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: OrgsDataSourceModel{
				Limit: types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListOrganizationsOptions(tc.model)
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
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestOrgsDataSourceOrgFromOrganization(t *testing.T) {
	t.Parallel()

	orgModel := orgsDataSourceOrgFromOrganization(&client.Organization{
		ID:                 "org-1",
		Name:               "Acme",
		APIURL:             stringPtr("https://api.acme.dev"),
		IsUniversalAPI:     boolPtr(true),
		IsDataplanePrivate: boolPtr(false),
		ProxyURL:           stringPtr("https://proxy.acme.dev"),
		RealtimeURL:        stringPtr("wss://realtime.acme.dev"),
		Created:            stringPtr("2026-02-26T00:00:00Z"),
		ImageRenderingMode: stringPtr("auto"),
	})

	if orgModel.ID.ValueString() != "org-1" {
		t.Fatalf("id mismatch: got=%q", orgModel.ID.ValueString())
	}
	if orgModel.Name.ValueString() != "Acme" {
		t.Fatalf("name mismatch: got=%q", orgModel.Name.ValueString())
	}
	if orgModel.APIURL.ValueString() != "https://api.acme.dev" {
		t.Fatalf("api_url mismatch: got=%q", orgModel.APIURL.ValueString())
	}
	if !orgModel.IsUniversalAPI.ValueBool() {
		t.Fatalf("is_universal_api mismatch: got=%v", orgModel.IsUniversalAPI.ValueBool())
	}
	if orgModel.IsDataplanePrivate.ValueBool() {
		t.Fatalf("is_dataplane_private mismatch: got=%v", orgModel.IsDataplanePrivate.ValueBool())
	}
	if orgModel.ProxyURL.ValueString() != "https://proxy.acme.dev" {
		t.Fatalf("proxy_url mismatch: got=%q", orgModel.ProxyURL.ValueString())
	}
	if orgModel.RealtimeURL.ValueString() != "wss://realtime.acme.dev" {
		t.Fatalf("realtime_url mismatch: got=%q", orgModel.RealtimeURL.ValueString())
	}
	if orgModel.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", orgModel.Created.ValueString())
	}
	if orgModel.ImageRenderingMode.ValueString() != "auto" {
		t.Fatalf("image_rendering_mode mismatch: got=%q", orgModel.ImageRenderingMode.ValueString())
	}
}

func TestProviderDataSourcesIncludeOrganizationPair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_org"]; !ok {
		t.Fatalf("expected braintrustdata_org to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_orgs"]; !ok {
		t.Fatalf("expected braintrustdata_orgs to be registered")
	}
}
