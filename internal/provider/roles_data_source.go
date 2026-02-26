package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RolesDataSource{}

// NewRolesDataSource creates a new roles data source instance.
func NewRolesDataSource() datasource.DataSource {
	return &RolesDataSource{}
}

// RolesDataSource defines the data source implementation.
type RolesDataSource struct {
	client *client.Client
}

// RolesDataSourceModel describes the data source data model.
type RolesDataSourceModel struct {
	OrgName       types.String          `tfsdk:"org_name"`
	RoleName      types.String          `tfsdk:"role_name"`
	StartingAfter types.String          `tfsdk:"starting_after"`
	EndingBefore  types.String          `tfsdk:"ending_before"`
	Roles         []RolesDataSourceRole `tfsdk:"roles"`
	IDs           []string              `tfsdk:"ids"`
	Limit         types.Int64           `tfsdk:"limit"`
}

// RolesDataSourceRole represents a single role in the list.
type RolesDataSourceRole struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	OrgID             types.String `tfsdk:"org_id"`
	Description       types.String `tfsdk:"description"`
	MemberPermissions types.List   `tfsdk:"member_permissions"`
	MemberRoles       types.List   `tfsdk:"member_roles"`
	Created           types.String `tfsdk:"created"`
	UserID            types.String `tfsdk:"user_id"`
}

// Metadata implements datasource.DataSource.
func (d *RolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

// Schema implements datasource.DataSource.
func (d *RolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust roles using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter.",
			},
			"role_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact role name filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of roles to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch roles after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch roles before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned role IDs.",
			},
			"roles": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the role.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the role.",
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
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *RolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, err := buildListRolesOptions(data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Filters", err.Error())
		return
	}

	listResp, err := d.client.ListRoles(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Roles",
			fmt.Sprintf("Could not list roles: %s", err.Error()),
		)
		return
	}

	data.Roles = make([]RolesDataSourceRole, 0, len(listResp.Roles))
	data.IDs = make([]string, 0, len(listResp.Roles))

	for i := range listResp.Roles {
		role := &listResp.Roles[i]
		if role.DeletedAt != "" {
			continue
		}

		roleModel, roleDiags := rolesDataSourceRoleFromRole(ctx, role)
		resp.Diagnostics.Append(roleDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Roles = append(data.Roles, roleModel)
		data.IDs = append(data.IDs, role.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListRolesOptions(data RolesDataSourceModel) (*client.ListRolesOptions, error) {
	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		return nil, fmt.Errorf("cannot specify both 'starting_after' and 'ending_before'")
	}

	listOpts := &client.ListRolesOptions{}

	if !data.OrgName.IsNull() && data.OrgName.ValueString() != "" {
		listOpts.OrgName = data.OrgName.ValueString()
	}
	if !data.RoleName.IsNull() && data.RoleName.ValueString() != "" {
		listOpts.RoleName = data.RoleName.ValueString()
	}
	if !data.Limit.IsNull() {
		limit := data.Limit.ValueInt64()
		if limit < 1 {
			return nil, fmt.Errorf("'limit' must be greater than or equal to 1")
		}

		maxInt := int64(^uint(0) >> 1)
		if limit > maxInt {
			return nil, fmt.Errorf("'limit' exceeds supported platform integer size")
		}

		listOpts.Limit = int(limit)
	}
	if hasStartingAfter {
		listOpts.StartingAfter = data.StartingAfter.ValueString()
	}
	if hasEndingBefore {
		listOpts.EndingBefore = data.EndingBefore.ValueString()
	}

	return listOpts, nil
}

func rolesDataSourceRoleFromRole(ctx context.Context, role *client.Role) (RolesDataSourceRole, diag.Diagnostics) {
	var diags diag.Diagnostics

	roleModel := RolesDataSourceRole{
		ID:   types.StringValue(role.ID),
		Name: types.StringValue(role.Name),
	}

	if role.OrgID != "" {
		roleModel.OrgID = types.StringValue(role.OrgID)
	} else {
		roleModel.OrgID = types.StringNull()
	}
	if role.Description != "" {
		roleModel.Description = types.StringValue(role.Description)
	} else {
		roleModel.Description = types.StringNull()
	}
	if role.Created != "" {
		roleModel.Created = types.StringValue(role.Created)
	} else {
		roleModel.Created = types.StringNull()
	}
	if role.UserID != "" {
		roleModel.UserID = types.StringValue(role.UserID)
	} else {
		roleModel.UserID = types.StringNull()
	}

	memberPermissions, memberPermissionDiags := listFromStringSlice(ctx, roleMemberPermissionStrings(role.MemberPermissions))
	diags.Append(memberPermissionDiags...)
	if diags.HasError() {
		return roleModel, diags
	}
	roleModel.MemberPermissions = memberPermissions

	memberRoles, memberRoleDiags := listFromStringSlice(ctx, role.MemberRoles)
	diags.Append(memberRoleDiags...)
	if diags.HasError() {
		return roleModel, diags
	}
	roleModel.MemberRoles = memberRoles

	return roleModel, diags
}
