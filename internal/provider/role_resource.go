package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

// NewRoleResource creates a new role resource instance.
func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

// RoleResource defines the resource implementation.
type RoleResource struct {
	client *client.Client
}

// RoleResourceModel describes the resource data model.
type RoleResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	OrgID             types.String `tfsdk:"org_id"`
	Description       types.String `tfsdk:"description"`
	MemberPermissions types.List   `tfsdk:"member_permissions"`
	MemberRoles       types.List   `tfsdk:"member_roles"`
	Created           types.String `tfsdk:"created"`
	UserID            types.String `tfsdk:"user_id"`
}

type listValueState int

const (
	listValueStateKnown listValueState = iota
	listValueStateNull
	listValueStateUnknown
)

// Metadata implements resource.Resource.
func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema implements resource.Resource.
func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust role. Roles define permission levels and can be assigned to users and groups for access control.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the role.",
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID that the role belongs to.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the role.",
			},
			"member_permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of permissions assigned to members of this role.",
			},
			"member_roles": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of role IDs assigned to members of this role.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the role was created.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the role.",
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create implements resource.Resource by creating a new role.
func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RoleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildRoleCreateRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create role via API
	role, err := r.client.CreateRole(ctx, createReq)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create role, got error: %s", err))
		return
	}

	// Read role after creation to ensure membership fields are populated.
	role, err = r.client.GetRole(ctx, role.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read role after creation, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(populateRoleState(ctx, &data, role)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource by reading a role.
func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get role from API
	role, err := r.client.GetRole(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read role, got error: %s", err))
		return
	}

	// Check for soft delete
	if role.DeletedAt != "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(populateRoleState(ctx, &data, role)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource by updating an existing role.
func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RoleResourceModel
	var state RoleResourceModel

	// Get current state to preserve computed fields
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentMemberPermissions, diags := listToStringSlice(ctx, state.MemberPermissions)
	resp.Diagnostics.Append(diags...)
	currentMemberRoles, diags := listToStringSlice(ctx, state.MemberRoles)
	resp.Diagnostics.Append(diags...)
	desiredMemberPermissions, desiredMemberPermissionsState, diags := listToStringSliceWithState(ctx, data.MemberPermissions)
	resp.Diagnostics.Append(diags...)
	desiredMemberRoles, desiredMemberRolesState, diags := listToStringSliceWithState(ctx, data.MemberRoles)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addMemberPermissions, removeMemberPermissions := computeStringSliceDiffForDesiredState(
		currentMemberPermissions,
		desiredMemberPermissions,
		desiredMemberPermissionsState,
	)
	addMemberRoles, removeMemberRoles := computeStringSliceDiffForDesiredState(
		currentMemberRoles,
		desiredMemberRoles,
		desiredMemberRolesState,
	)

	// Update role via API
	role, err := r.client.UpdateRole(ctx, data.ID.ValueString(), &client.UpdateRoleRequest{
		Name:                    data.Name.ValueString(),
		Description:             data.Description.ValueString(),
		AddMemberPermissions:    roleMemberPermissionsFromStrings(addMemberPermissions),
		RemoveMemberPermissions: roleMemberPermissionsFromStrings(removeMemberPermissions),
		AddMemberRoles:          addMemberRoles,
		RemoveMemberRoles:       removeMemberRoles,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update role, got error: %s", err))
		return
	}

	// Read role after update to ensure membership fields are populated.
	role, err = r.client.GetRole(ctx, role.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read role after update, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(populateRoleState(ctx, &data, role)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete implements resource.Resource by deleting a role.
func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete role via API (soft delete)
	err := r.client.DeleteRole(ctx, data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete role, got error: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState by importing a role by ID.
func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func listToStringSlice(ctx context.Context, values types.List) ([]string, diag.Diagnostics) {
	result, _, diags := listToStringSliceWithState(ctx, values)
	return result, diags
}

func buildRoleCreateRequest(ctx context.Context, data RoleResourceModel) (*client.CreateRoleRequest, diag.Diagnostics) {
	createReq := &client.CreateRoleRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	memberPermissions, memberPermissionsState, diags := listToStringSliceWithState(ctx, data.MemberPermissions)
	if diags.HasError() {
		return nil, diags
	}

	memberRoles, memberRolesState, memberRoleDiags := listToStringSliceWithState(ctx, data.MemberRoles)
	diags.Append(memberRoleDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if memberPermissionsState == listValueStateKnown {
		createReq.MemberPermissions = roleMemberPermissionsFromStrings(memberPermissions)
	}
	if memberRolesState == listValueStateKnown {
		createReq.MemberRoles = memberRoles
	}

	return createReq, diags
}

func listToStringSliceWithState(ctx context.Context, values types.List) ([]string, listValueState, diag.Diagnostics) {
	if values.IsNull() {
		return nil, listValueStateNull, nil
	}
	if values.IsUnknown() {
		return nil, listValueStateUnknown, nil
	}

	var elements []types.String
	// `allowUnhandled=true` is intentional: a known list can still include null/unknown
	// elements, and we need to filter those values out below rather than erroring.
	diags := values.ElementsAs(ctx, &elements, true)
	if diags.HasError() {
		return nil, listValueStateKnown, diags
	}

	result := make([]string, 0, len(elements))
	for _, element := range elements {
		if element.IsNull() || element.IsUnknown() {
			continue
		}
		result = append(result, element.ValueString())
	}

	return result, listValueStateKnown, diags
}

func listFromStringSlice(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValueFrom(ctx, types.StringType, values)
}

func populateRoleState(ctx context.Context, data *RoleResourceModel, role *client.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(role.ID)
	data.Name = types.StringValue(role.Name)
	data.Description = types.StringValue(role.Description)
	data.OrgID = types.StringValue(role.OrgID)
	data.Created = types.StringValue(role.Created)

	if role.UserID != "" {
		data.UserID = types.StringValue(role.UserID)
	} else {
		data.UserID = types.StringNull()
	}

	if role.MemberPermissions != nil {
		memberPermissions, listDiags := listFromStringSlice(ctx, roleMemberPermissionStrings(role.MemberPermissions))
		diags.Append(listDiags...)
		if diags.HasError() {
			return diags
		}
		data.MemberPermissions = memberPermissions
	}

	if role.MemberRoles != nil {
		memberRoles, listDiags := listFromStringSlice(ctx, role.MemberRoles)
		diags.Append(listDiags...)
		if diags.HasError() {
			return diags
		}
		data.MemberRoles = memberRoles
	}

	return diags
}

func computeStringSliceDiff(current []string, desired []string) ([]string, []string) {
	currentSet := make(map[string]struct{}, len(current))
	for _, value := range current {
		currentSet[value] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(desired))
	for _, value := range desired {
		desiredSet[value] = struct{}{}
	}

	var additions []string
	for _, value := range desired {
		if _, exists := currentSet[value]; !exists {
			additions = append(additions, value)
		}
	}

	var removals []string
	for _, value := range current {
		if _, exists := desiredSet[value]; !exists {
			removals = append(removals, value)
		}
	}

	return additions, removals
}

func computeStringSliceDiffForDesiredState(current []string, desired []string, desiredState listValueState) ([]string, []string) {
	if desiredState == listValueStateUnknown {
		return nil, nil
	}

	return computeStringSliceDiff(current, desired)
}

func roleMemberPermissionsFromStrings(permissions []string) []client.RoleMemberPermission {
	if permissions == nil {
		return nil
	}

	memberPermissions := make([]client.RoleMemberPermission, 0, len(permissions))
	for _, permission := range permissions {
		memberPermissions = append(memberPermissions, client.RoleMemberPermission{
			Permission: permission,
		})
	}

	return memberPermissions
}

func roleMemberPermissionStrings(memberPermissions []client.RoleMemberPermission) []string {
	if len(memberPermissions) == 0 {
		return nil
	}

	permissions := make([]string, 0, len(memberPermissions))
	for _, memberPermission := range memberPermissions {
		if memberPermission.Permission == "" {
			continue
		}
		permissions = append(permissions, memberPermission.Permission)
	}

	return permissions
}
