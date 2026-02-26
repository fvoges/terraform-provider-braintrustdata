package provider

import (
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPopulateACLDataSourceModel(t *testing.T) {
	t.Parallel()

	model := ACLDataSourceModel{}
	acl := &client.ACL{
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
	}

	populateACLDataSourceModel(&model, acl)

	if model.ID.ValueString() != "acl-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.ObjectOrgID.ValueString() != "org-1" {
		t.Fatalf("object_org_id mismatch: got=%q", model.ObjectOrgID.ValueString())
	}
	if model.ObjectID.ValueString() != "project-1" {
		t.Fatalf("object_id mismatch: got=%q", model.ObjectID.ValueString())
	}
	if model.ObjectType.ValueString() != "project" {
		t.Fatalf("object_type mismatch: got=%q", model.ObjectType.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}
	if model.GroupID.ValueString() != "group-1" {
		t.Fatalf("group_id mismatch: got=%q", model.GroupID.ValueString())
	}
	if model.RoleID.ValueString() != "role-1" {
		t.Fatalf("role_id mismatch: got=%q", model.RoleID.ValueString())
	}
	if model.Permission.ValueString() != "read" {
		t.Fatalf("permission mismatch: got=%q", model.Permission.ValueString())
	}
	if model.RestrictObjectType.ValueString() != "dataset" {
		t.Fatalf("restrict_object_type mismatch: got=%q", model.RestrictObjectType.ValueString())
	}
	if model.Created.ValueString() != "2026-02-26T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
}

func TestPopulateACLDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := ACLDataSourceModel{}
	acl := &client.ACL{
		ID:         "acl-2",
		ObjectID:   "project-2",
		ObjectType: client.ACLObjectTypeProject,
	}

	populateACLDataSourceModel(&model, acl)

	if !model.ObjectOrgID.IsNull() {
		t.Fatalf("expected object_org_id to be null")
	}
	if !model.UserID.IsNull() {
		t.Fatalf("expected user_id to be null")
	}
	if !model.GroupID.IsNull() {
		t.Fatalf("expected group_id to be null")
	}
	if !model.RoleID.IsNull() {
		t.Fatalf("expected role_id to be null")
	}
	if !model.Permission.IsNull() {
		t.Fatalf("expected permission to be null")
	}
	if !model.RestrictObjectType.IsNull() {
		t.Fatalf("expected restrict_object_type to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
}

func TestValidateACLID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		id          types.String
		wantID      string
		wantErrLike string
	}{
		{
			name:   "valid",
			id:     types.StringValue("acl-1"),
			wantID: "acl-1",
		},
		{
			name:        "unknown",
			id:          types.StringUnknown(),
			wantErrLike: "ACL ID is unknown",
		},
		{
			name:        "null",
			id:          types.StringNull(),
			wantErrLike: "ACL ID must be provided and non-empty",
		},
		{
			name:        "empty",
			id:          types.StringValue(""),
			wantErrLike: "ACL ID must be provided and non-empty",
		},
		{
			name:        "whitespace",
			id:          types.StringValue("   "),
			wantErrLike: "ACL ID must be provided and non-empty",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotID, diags := validateACLID(tc.id)
			if tc.wantErrLike == "" {
				if diags.HasError() {
					t.Fatalf("unexpected diagnostics: %v", diags)
				}
				if gotID != tc.wantID {
					t.Fatalf("id mismatch: got=%q want=%q", gotID, tc.wantID)
				}
				return
			}

			if !diags.HasError() {
				t.Fatalf("expected diagnostic containing %q, got none", tc.wantErrLike)
			}
			found := false
			for _, d := range diags {
				if strings.Contains(d.Detail(), tc.wantErrLike) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected diagnostic containing %q, got %v", tc.wantErrLike, diags)
			}
		})
	}
}
