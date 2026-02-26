package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSelectSingleRoleByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		roleName    string
		wantID      string
		roles       []client.Role
	}{
		"finds exact active role": {
			roles: []client.Role{
				{ID: "role-a", Name: "other"},
				{ID: "role-b", Name: "target"},
			},
			roleName: "target",
			wantID:   "role-b",
		},
		"ignores deleted roles": {
			roles: []client.Role{
				{ID: "role-a", Name: "target", DeletedAt: "2026-01-01T00:00:00Z"},
				{ID: "role-b", Name: "target"},
			},
			roleName: "target",
			wantID:   "role-b",
		},
		"returns not found when no exact match": {
			roles: []client.Role{
				{ID: "role-a", Name: "admin"},
			},
			roleName:    "target",
			wantErrType: errRoleNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			roles: []client.Role{
				{ID: "role-a", Name: "target"},
				{ID: "role-b", Name: "target"},
			},
			roleName:    "target",
			wantErrType: errMultipleRolesFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			role, err := selectSingleRoleByName(tc.roles, tc.roleName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if role == nil {
				t.Fatalf("expected role, got nil")
			}
			if role.ID != tc.wantID {
				t.Fatalf("expected role ID %q, got %q", tc.wantID, role.ID)
			}
		})
	}
}

func TestPopulateRoleDataSourceModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := RoleDataSourceModel{}
	role := &client.Role{
		ID:          "role-1",
		Name:        "admin",
		OrgID:       "org-1",
		Description: "Admin role",
		Created:     "2026-02-26T00:00:00Z",
		UserID:      "user-1",
		MemberPermissions: []client.RoleMemberPermission{
			{Permission: "read"},
			{Permission: "update"},
		},
		MemberRoles: []string{"role-parent"},
	}

	diags := populateRoleDataSourceModel(ctx, &model, role)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "role-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "admin" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.OrgID.ValueString() != "org-1" {
		t.Fatalf("org_id mismatch: got=%q", model.OrgID.ValueString())
	}
	if model.Description.ValueString() != "Admin role" {
		t.Fatalf("description mismatch: got=%q", model.Description.ValueString())
	}
	if model.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}

	gotPermissions, permissionDiags := listToStringSlice(ctx, model.MemberPermissions)
	if permissionDiags.HasError() {
		t.Fatalf("unexpected permissions diagnostics: %v", permissionDiags)
	}
	if len(gotPermissions) != 2 || gotPermissions[0] != "read" || gotPermissions[1] != "update" {
		t.Fatalf("member_permissions mismatch: got=%v", gotPermissions)
	}

	gotRoles, roleDiags := listToStringSlice(ctx, model.MemberRoles)
	if roleDiags.HasError() {
		t.Fatalf("unexpected roles diagnostics: %v", roleDiags)
	}
	if len(gotRoles) != 1 || gotRoles[0] != "role-parent" {
		t.Fatalf("member_roles mismatch: got=%v", gotRoles)
	}
}

func TestPopulateRoleDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := RoleDataSourceModel{
		MemberPermissions: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("stale"),
		}),
		MemberRoles: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("stale"),
		}),
	}
	role := &client.Role{
		ID:   "role-2",
		Name: "viewer",
	}

	diags := populateRoleDataSourceModel(ctx, &model, role)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.OrgID.IsNull() {
		t.Fatalf("expected org_id to be null")
	}
	if !model.Description.IsNull() {
		t.Fatalf("expected description to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
	if !model.UserID.IsNull() {
		t.Fatalf("expected user_id to be null")
	}
	if !model.MemberPermissions.IsNull() {
		t.Fatalf("expected member_permissions to be null")
	}
	if !model.MemberRoles.IsNull() {
		t.Fatalf("expected member_roles to be null")
	}
}
