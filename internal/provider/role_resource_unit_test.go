package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestComputeStringSliceDiff(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		current []string
		desired []string
		wantAdd []string
		wantDel []string
	}{
		{
			name:    "no_changes",
			current: []string{"read", "update"},
			desired: []string{"read", "update"},
			wantAdd: nil,
			wantDel: nil,
		},
		{
			name:    "adds_and_removes",
			current: []string{"read", "delete"},
			desired: []string{"read", "update"},
			wantAdd: []string{"update"},
			wantDel: []string{"delete"},
		},
		{
			name:    "adds_all_when_current_empty",
			current: nil,
			desired: []string{"read", "update"},
			wantAdd: []string{"read", "update"},
			wantDel: nil,
		},
		{
			name:    "removes_all_when_desired_empty",
			current: []string{"read", "update"},
			desired: nil,
			wantAdd: nil,
			wantDel: []string{"read", "update"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotAdd, gotDel := computeStringSliceDiff(tc.current, tc.desired)
			if !reflect.DeepEqual(gotAdd, tc.wantAdd) {
				t.Fatalf("computeStringSliceDiff() add mismatch: got=%v want=%v", gotAdd, tc.wantAdd)
			}
			if !reflect.DeepEqual(gotDel, tc.wantDel) {
				t.Fatalf("computeStringSliceDiff() remove mismatch: got=%v want=%v", gotDel, tc.wantDel)
			}
		})
	}
}

func TestListToStringSlice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name   string
		input  types.List
		want   []string
		hasErr bool
	}{
		{
			name:  "null_list",
			input: types.ListNull(types.StringType),
			want:  nil,
		},
		{
			name: "known_values",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringValue("update"),
			}),
			want: []string{"read", "update"},
		},
		{
			name: "null_and_unknown_elements_are_filtered",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue("delete"),
			}),
			want: []string{"read", "delete"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, diags := listToStringSlice(ctx, tc.input)
			if tc.hasErr != diags.HasError() {
				t.Fatalf("listToStringSlice() diagnostics mismatch: hasErr=%v diags=%v", tc.hasErr, diags)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("listToStringSlice() mismatch: got=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestListToStringSliceWithState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name      string
		input     types.List
		want      []string
		wantState listValueState
		hasErr    bool
	}{
		{
			name:      "null_list",
			input:     types.ListNull(types.StringType),
			want:      nil,
			wantState: listValueStateNull,
		},
		{
			name:      "unknown_list",
			input:     types.ListUnknown(types.StringType),
			want:      nil,
			wantState: listValueStateUnknown,
		},
		{
			name: "known_values",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringValue("update"),
			}),
			want:      []string{"read", "update"},
			wantState: listValueStateKnown,
		},
		{
			name: "known_values_with_null_and_unknown_elements_filtered_without_diagnostics",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("read"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue("delete"),
			}),
			want:      []string{"read", "delete"},
			wantState: listValueStateKnown,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotState, diags := listToStringSliceWithState(ctx, tc.input)
			if tc.hasErr != diags.HasError() {
				t.Fatalf("listToStringSliceWithState() diagnostics mismatch: hasErr=%v diags=%v", tc.hasErr, diags)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("listToStringSliceWithState() mismatch: got=%v want=%v", got, tc.want)
			}
			if gotState != tc.wantState {
				t.Fatalf("listToStringSliceWithState() state mismatch: got=%v want=%v", gotState, tc.wantState)
			}
		})
	}
}

func TestComputeStringSliceDiffForDesiredState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		current      []string
		desired      []string
		wantAdd      []string
		wantDel      []string
		desiredState listValueState
	}{
		{
			name:         "desired_unknown_skips_diff",
			current:      []string{"read", "update"},
			desired:      nil,
			desiredState: listValueStateUnknown,
			wantAdd:      nil,
			wantDel:      nil,
		},
		{
			name:         "desired_null_removes_existing",
			current:      []string{"read", "update"},
			desired:      nil,
			desiredState: listValueStateNull,
			wantAdd:      nil,
			wantDel:      []string{"read", "update"},
		},
		{
			name:         "desired_known_uses_regular_diff",
			current:      []string{"read"},
			desired:      []string{"read", "update"},
			desiredState: listValueStateKnown,
			wantAdd:      []string{"update"},
			wantDel:      nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotAdd, gotDel := computeStringSliceDiffForDesiredState(tc.current, tc.desired, tc.desiredState)
			if !reflect.DeepEqual(gotAdd, tc.wantAdd) {
				t.Fatalf("computeStringSliceDiffForDesiredState() add mismatch: got=%v want=%v", gotAdd, tc.wantAdd)
			}
			if !reflect.DeepEqual(gotDel, tc.wantDel) {
				t.Fatalf("computeStringSliceDiffForDesiredState() remove mismatch: got=%v want=%v", gotDel, tc.wantDel)
			}
		})
	}
}

func TestBuildRoleCreateRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		want                *client.CreateRoleRequest
		data                RoleResourceModel
		name                string
		wantDiagnosticsFail bool
	}{
		{
			name: "known_memberships_are_sent_on_create_and_filtered",
			data: RoleResourceModel{
				Name:        types.StringValue("role-1"),
				Description: types.StringValue("desc"),
				MemberPermissions: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("read"),
					types.StringNull(),
					types.StringUnknown(),
					types.StringValue("update"),
				}),
				MemberRoles: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("role-a"),
				}),
			},
			want: &client.CreateRoleRequest{
				Name:        "role-1",
				Description: "desc",
				MemberPermissions: []client.RoleMemberPermission{
					{Permission: "read"},
					{Permission: "update"},
				},
				MemberRoles: []string{"role-a"},
			},
		},
		{
			name: "unknown_memberships_are_omitted",
			data: RoleResourceModel{
				Name:              types.StringValue("role-2"),
				Description:       types.StringValue("desc"),
				MemberPermissions: types.ListUnknown(types.StringType),
				MemberRoles:       types.ListUnknown(types.StringType),
			},
			want: &client.CreateRoleRequest{
				Name:        "role-2",
				Description: "desc",
			},
		},
		{
			name: "known_empty_memberships_are_sent_as_empty_lists",
			data: RoleResourceModel{
				Name:              types.StringValue("role-3"),
				Description:       types.StringValue("desc"),
				MemberPermissions: types.ListValueMust(types.StringType, []attr.Value{}),
				MemberRoles:       types.ListValueMust(types.StringType, []attr.Value{}),
			},
			want: &client.CreateRoleRequest{
				Name:              "role-3",
				Description:       "desc",
				MemberPermissions: []client.RoleMemberPermission{},
				MemberRoles:       []string{},
			},
		},
		{
			name: "null_memberships_are_omitted_but_known_empty_after_filtering_is_sent",
			data: RoleResourceModel{
				Name:        types.StringValue("role-4"),
				Description: types.StringValue("desc"),
				// Null list means "unset" and should be omitted from payload.
				MemberPermissions: types.ListNull(types.StringType),
				// Known list with only null/unknown elements is known-empty after filtering.
				MemberRoles: types.ListValueMust(types.StringType, []attr.Value{
					types.StringNull(),
					types.StringUnknown(),
				}),
			},
			want: &client.CreateRoleRequest{
				Name:        "role-4",
				Description: "desc",
				MemberRoles: []string{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, diags := buildRoleCreateRequest(ctx, tc.data)
			if tc.wantDiagnosticsFail != diags.HasError() {
				t.Fatalf("buildRoleCreateRequest() diagnostics mismatch: hasErr=%v diags=%v", tc.wantDiagnosticsFail, diags)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("buildRoleCreateRequest() mismatch: got=%#v want=%#v", got, tc.want)
			}
		})
	}
}

func TestRoleMemberPermissionConversions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     []string
		want      []client.RoleMemberPermission
		wantRound []string
	}{
		{
			name:  "empty",
			input: nil,
			want:  nil,
		},
		{
			name:  "values",
			input: []string{"read", "update"},
			want: []client.RoleMemberPermission{
				{Permission: "read"},
				{Permission: "update"},
			},
			wantRound: []string{"read", "update"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := roleMemberPermissionsFromStrings(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("roleMemberPermissionsFromStrings() mismatch: got=%#v want=%#v", got, tc.want)
			}

			gotRound := roleMemberPermissionStrings(got)
			if !reflect.DeepEqual(gotRound, tc.wantRound) {
				t.Fatalf("roleMemberPermissionStrings() mismatch: got=%#v want=%#v", gotRound, tc.wantRound)
			}
		})
	}
}

func TestUpdateRoleRequestRoleDiffUsesStringSlices(t *testing.T) {
	t.Parallel()

	req := client.UpdateRoleRequest{
		AddMemberRoles:    []string{"role-a"},
		RemoveMemberRoles: []string{"role-b"},
	}

	if !reflect.DeepEqual(req.AddMemberRoles, []string{"role-a"}) {
		t.Fatalf("AddMemberRoles mismatch: got=%#v want=%#v", req.AddMemberRoles, []string{"role-a"})
	}
	if !reflect.DeepEqual(req.RemoveMemberRoles, []string{"role-b"}) {
		t.Fatalf("RemoveMemberRoles mismatch: got=%#v want=%#v", req.RemoveMemberRoles, []string{"role-b"})
	}
}

func TestPopulateRoleStatePreservesExistingMembershipWhenResponseOmitsFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := RoleResourceModel{
		MemberPermissions: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("perm.keep"),
		}),
		MemberRoles: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("role-keep"),
		}),
	}

	role := &client.Role{
		ID:          "role-id",
		Name:        "role-name",
		Description: "desc",
		OrgID:       "org-id",
		Created:     "created-at",
		// API omits membership fields (nil slices)
		MemberPermissions: nil,
		MemberRoles:       nil,
	}

	diags := populateRoleState(ctx, &data, role)
	if diags.HasError() {
		t.Fatalf("populateRoleState() unexpected diagnostics: %v", diags)
	}

	gotPermissions, permDiags := listToStringSlice(ctx, data.MemberPermissions)
	if permDiags.HasError() {
		t.Fatalf("listToStringSlice(member_permissions) unexpected diagnostics: %v", permDiags)
	}
	if !reflect.DeepEqual(gotPermissions, []string{"perm.keep"}) {
		t.Fatalf("member_permissions changed unexpectedly: got=%v want=%v", gotPermissions, []string{"perm.keep"})
	}

	gotRoles, roleDiags := listToStringSlice(ctx, data.MemberRoles)
	if roleDiags.HasError() {
		t.Fatalf("listToStringSlice(member_roles) unexpected diagnostics: %v", roleDiags)
	}
	if !reflect.DeepEqual(gotRoles, []string{"role-keep"}) {
		t.Fatalf("member_roles changed unexpectedly: got=%v want=%v", gotRoles, []string{"role-keep"})
	}
}

func TestPopulateRoleStateOverwritesMembershipWhenResponseIncludesFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := RoleResourceModel{
		MemberPermissions: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("perm.old"),
		}),
		MemberRoles: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("role-old"),
		}),
	}

	role := &client.Role{
		ID:          "role-id",
		Name:        "role-name",
		Description: "desc",
		OrgID:       "org-id",
		Created:     "created-at",
		MemberPermissions: []client.RoleMemberPermission{
			{Permission: "perm.new.a"},
			{Permission: "perm.new.b"},
		},
		MemberRoles: []string{"role-new"},
	}

	diags := populateRoleState(ctx, &data, role)
	if diags.HasError() {
		t.Fatalf("populateRoleState() unexpected diagnostics: %v", diags)
	}

	gotPermissions, permDiags := listToStringSlice(ctx, data.MemberPermissions)
	if permDiags.HasError() {
		t.Fatalf("listToStringSlice(member_permissions) unexpected diagnostics: %v", permDiags)
	}
	if !reflect.DeepEqual(gotPermissions, []string{"perm.new.a", "perm.new.b"}) {
		t.Fatalf("member_permissions mismatch: got=%v want=%v", gotPermissions, []string{"perm.new.a", "perm.new.b"})
	}

	gotRoles, roleDiags := listToStringSlice(ctx, data.MemberRoles)
	if roleDiags.HasError() {
		t.Fatalf("listToStringSlice(member_roles) unexpected diagnostics: %v", roleDiags)
	}
	if !reflect.DeepEqual(gotRoles, []string{"role-new"}) {
		t.Fatalf("member_roles mismatch: got=%v want=%v", gotRoles, []string{"role-new"})
	}
}
