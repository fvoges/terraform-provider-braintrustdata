package provider

import (
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
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
