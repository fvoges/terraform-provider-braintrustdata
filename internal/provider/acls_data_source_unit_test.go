package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListACLsOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListACLsOptions
		model       ACLsDataSourceModel
	}{
		"builds all supported api-native filters": {
			model: ACLsDataSourceModel{
				ObjectID:      types.StringValue("project-1"),
				ObjectType:    types.StringValue("project"),
				StartingAfter: types.StringValue("acl-1"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListACLsOptions{
				ObjectID:      "project-1",
				ObjectType:    client.ACLObjectTypeProject,
				StartingAfter: "acl-1",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: ACLsDataSourceModel{
				ObjectID:      types.StringValue("project-1"),
				ObjectType:    types.StringValue("project"),
				StartingAfter: types.StringValue("acl-1"),
				EndingBefore:  types.StringValue("acl-2"),
			},
			wantErrLike: "cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: ACLsDataSourceModel{
				ObjectID:   types.StringValue("project-1"),
				ObjectType: types.StringValue("project"),
				Limit:      types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListACLsOptions(tc.model)
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
			if opts.ObjectID != tc.want.ObjectID ||
				opts.ObjectType != tc.want.ObjectType ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestACLsDataSourceACLFromACL(t *testing.T) {
	t.Parallel()

	aclModel := aclsDataSourceACLFromACL(&client.ACL{
		ID:                 "acl-1",
		ObjectOrgID:        "org-1",
		ObjectID:           "project-1",
		ObjectType:         client.ACLObjectTypeProject,
		UserID:             "user-1",
		GroupID:            "group-1",
		RoleID:             "role-1",
		Permission:         client.PermissionRead,
		RestrictObjectType: client.ACLObjectTypeDataset,
		Created:            "2026-02-26T00:00:00Z",
	})

	if aclModel.ID.ValueString() != "acl-1" {
		t.Fatalf("id mismatch: got=%q", aclModel.ID.ValueString())
	}
	if aclModel.ObjectOrgID.ValueString() != "org-1" {
		t.Fatalf("object_org_id mismatch: got=%q", aclModel.ObjectOrgID.ValueString())
	}
	if aclModel.ObjectID.ValueString() != "project-1" {
		t.Fatalf("object_id mismatch: got=%q", aclModel.ObjectID.ValueString())
	}
	if aclModel.ObjectType.ValueString() != "project" {
		t.Fatalf("object_type mismatch: got=%q", aclModel.ObjectType.ValueString())
	}
	if aclModel.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", aclModel.UserID.ValueString())
	}
	if aclModel.GroupID.ValueString() != "group-1" {
		t.Fatalf("group_id mismatch: got=%q", aclModel.GroupID.ValueString())
	}
	if aclModel.RoleID.ValueString() != "role-1" {
		t.Fatalf("role_id mismatch: got=%q", aclModel.RoleID.ValueString())
	}
	if aclModel.Permission.ValueString() != "read" {
		t.Fatalf("permission mismatch: got=%q", aclModel.Permission.ValueString())
	}
	if aclModel.RestrictObjectType.ValueString() != "dataset" {
		t.Fatalf("restrict_object_type mismatch: got=%q", aclModel.RestrictObjectType.ValueString())
	}
	if aclModel.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", aclModel.Created.ValueString())
	}
}

func TestProviderDataSourcesIncludeACLPair(t *testing.T) {
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

	if _, ok := dataSourceNames["braintrustdata_acl"]; !ok {
		t.Fatalf("expected braintrustdata_acl to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_acls"]; !ok {
		t.Fatalf("expected braintrustdata_acls to be registered")
	}
}
