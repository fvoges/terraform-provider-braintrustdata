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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &GroupsDataSource{}

// NewGroupsDataSource creates a new groups data source instance.
func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

// GroupsDataSource defines the data source implementation.
type GroupsDataSource struct {
	client *client.Client
}

// GroupsDataSourceModel describes the data source data model.
type GroupsDataSourceModel struct {
	OrgID  types.String            `tfsdk:"org_id"`
	Groups []GroupsDataSourceGroup `tfsdk:"groups"`
	IDs    []string                `tfsdk:"ids"`
}

// GroupsDataSourceGroup represents a single group in the list.
type GroupsDataSourceGroup struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	OrgID        types.String `tfsdk:"org_id"`
	Description  types.String `tfsdk:"description"`
	MemberUsers  types.List   `tfsdk:"member_users"`
	MemberGroups types.List   `tfsdk:"member_groups"`
	Created      types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *GroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

// Schema implements datasource.DataSource.
func (d *GroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Braintrust groups in an organization.",

		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The organization ID to filter groups. Defaults to the provider's organization_id.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of group IDs.",
			},
			"groups": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the group.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the group.",
						},
						"org_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The organization ID.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of the group.",
						},
						"member_users": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of user IDs that are members of this group.",
						},
						"member_groups": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of group IDs that are members of this group.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the group was created.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *GroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine org ID
	orgID := d.client.OrgID()
	if !data.OrgID.IsNull() && data.OrgID.ValueString() != "" {
		orgID = data.OrgID.ValueString()
	}

	// List groups from API
	listResp, err := d.client.ListGroups(ctx, &client.ListGroupsOptions{
		OrgID: orgID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Groups",
			fmt.Sprintf("Could not list groups: %s", err.Error()),
		)
		return
	}

	// Map response to data model
	data.OrgID = types.StringValue(orgID)
	data.Groups = make([]GroupsDataSourceGroup, 0, len(listResp.Groups))
	data.IDs = make([]string, 0, len(listResp.Groups))

	for _, group := range listResp.Groups {
		// Convert member lists
		var memberUsersList, memberGroupsList types.List
		if len(group.MemberUsers) > 0 {
			var diags diag.Diagnostics
			memberUsersList, diags = types.ListValueFrom(ctx, types.StringType, group.MemberUsers)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			memberUsersList = types.ListNull(types.StringType)
		}

		if len(group.MemberGroups) > 0 {
			var diags diag.Diagnostics
			memberGroupsList, diags = types.ListValueFrom(ctx, types.StringType, group.MemberGroups)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			memberGroupsList = types.ListNull(types.StringType)
		}

		groupModel := GroupsDataSourceGroup{
			ID:           types.StringValue(group.ID),
			Name:         types.StringValue(group.Name),
			OrgID:        types.StringValue(group.OrgID),
			Description:  types.StringValue(group.Description),
			MemberUsers:  memberUsersList,
			MemberGroups: memberGroupsList,
			Created:      types.StringValue(group.Created),
		}

		data.Groups = append(data.Groups, groupModel)
		data.IDs = append(data.IDs, group.ID)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
