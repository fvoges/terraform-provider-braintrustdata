package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListRolesOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListRolesOptions
		model       RolesDataSourceModel
	}{
		"builds all supported API-native filters": {
			model: RolesDataSourceModel{
				OrgName:       types.StringValue("example-org"),
				RoleName:      types.StringValue("admin"),
				StartingAfter: types.StringValue("role-1"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListRolesOptions{
				OrgName:       "example-org",
				RoleName:      "admin",
				StartingAfter: "role-1",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: RolesDataSourceModel{
				StartingAfter: types.StringValue("role-1"),
				EndingBefore:  types.StringValue("role-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: RolesDataSourceModel{
				Limit: types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListRolesOptions(tc.model)
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
				opts.RoleName != tc.want.RoleName ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestProviderDataSourcesIncludeRolePair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_role"]; !ok {
		t.Fatalf("expected braintrustdata_role to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_roles"]; !ok {
		t.Fatalf("expected braintrustdata_roles to be registered")
	}
}
