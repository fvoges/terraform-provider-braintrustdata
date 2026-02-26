package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RoleDataSource{}

var (
	errRoleNotFoundByName       = errors.New("role not found by name")
	errMultipleRolesFoundByName = errors.New("multiple roles found by name")
)

// NewRoleDataSource creates a new role data source instance.
func NewRoleDataSource() datasource.DataSource {
	return &RoleDataSource{}
}

// RoleDataSource defines the data source implementation.
type RoleDataSource struct {
	client *client.Client
}

// RoleDataSourceModel describes the data source data model.
type RoleDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	OrgName           types.String `tfsdk:"org_name"`
	OrgID             types.String `tfsdk:"org_id"`
	Description       types.String `tfsdk:"description"`
	MemberPermissions types.List   `tfsdk:"member_permissions"`
	MemberRoles       types.List   `tfsdk:"member_roles"`
	Created           types.String `tfsdk:"created"`
	UserID            types.String `tfsdk:"user_id"`
}

// Metadata implements datasource.DataSource.
func (d *RoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema implements datasource.DataSource.
func (d *RoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust role by `id` or by API-native searchable attributes (`name`, optionally `org_name`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the role. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The role name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter applied during searchable lookups.",
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID that the role belongs to.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of the role.",
			},
			"member_permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of permissions assigned to members of this role.",
			},
			"member_roles": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
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

// Configure implements datasource.DataSource.
func (d *RoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *RoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""
	hasOrgName := !data.OrgName.IsNull() && data.OrgName.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the role.",
		)
		return
	}

	if hasID && (hasName || hasOrgName) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'org_name').",
		)
		return
	}

	var role *client.Role
	if hasID {
		fetchedRole, err := d.client.GetRole(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Role",
				fmt.Sprintf("Could not read role ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
		role = fetchedRole
	} else {
		listOpts := &client.ListRolesOptions{
			RoleName: data.Name.ValueString(),
			Limit:    2,
		}
		if hasOrgName {
			listOpts.OrgName = data.OrgName.ValueString()
		}

		listResp, err := d.client.ListRoles(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Roles",
				fmt.Sprintf("Could not list roles using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedRole, err := selectSingleRoleByName(listResp.Roles, data.Name.ValueString())
		if errors.Is(err, errRoleNotFoundByName) {
			resp.Diagnostics.AddError(
				"Role Not Found",
				fmt.Sprintf("No role found with name: %s", data.Name.ValueString()),
			)
			return
		}
		if errors.Is(err, errMultipleRolesFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple Roles Found",
				"Searchable attributes matched multiple roles. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Roles",
				fmt.Sprintf("Could not resolve role using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		role = selectedRole
	}

	if role.DeletedAt != "" {
		resp.Diagnostics.AddError(
			"Role Not Found",
			"The requested role has been deleted.",
		)
		return
	}

	resp.Diagnostics.Append(populateRoleDataSourceModel(ctx, &data, role)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func selectSingleRoleByName(roles []client.Role, roleName string) (*client.Role, error) {
	var selected *client.Role

	for i := range roles {
		role := &roles[i]
		if role.DeletedAt != "" || role.Name != roleName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleRolesFoundByName, roleName)
		}
		selected = role
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errRoleNotFoundByName, roleName)
	}

	return selected, nil
}

func populateRoleDataSourceModel(ctx context.Context, data *RoleDataSourceModel, role *client.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(role.ID)
	data.Name = types.StringValue(role.Name)

	if role.OrgID != "" {
		data.OrgID = types.StringValue(role.OrgID)
	} else {
		data.OrgID = types.StringNull()
	}
	if role.Description != "" {
		data.Description = types.StringValue(role.Description)
	} else {
		data.Description = types.StringNull()
	}
	if role.Created != "" {
		data.Created = types.StringValue(role.Created)
	} else {
		data.Created = types.StringNull()
	}
	if role.UserID != "" {
		data.UserID = types.StringValue(role.UserID)
	} else {
		data.UserID = types.StringNull()
	}

	memberPermissions, memberPermissionDiags := listFromStringSlice(ctx, roleMemberPermissionStrings(role.MemberPermissions))
	diags.Append(memberPermissionDiags...)
	if diags.HasError() {
		return diags
	}
	data.MemberPermissions = memberPermissions

	memberRoles, memberRoleDiags := listFromStringSlice(ctx, role.MemberRoles)
	diags.Append(memberRoleDiags...)
	if diags.HasError() {
		return diags
	}
	data.MemberRoles = memberRoles

	return diags
}
