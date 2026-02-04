package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &GroupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

// GroupDataSource defines the data source implementation.
type GroupDataSource struct {
	client *client.Client
}

// GroupDataSourceModel describes the data source data model.
type GroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgID       types.String `tfsdk:"org_id"`
	Description types.String `tfsdk:"description"`
	MemberIDs   types.List   `tfsdk:"member_ids"`
	Created     types.String `tfsdk:"created"`
}

func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust group by ID or name. Specify either `id` or `name`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the group. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the group. Specify either `id` or `name`.",
			},
			"org_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The organization ID. Defaults to the provider's organization_id.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of the group.",
			},
			"member_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of user IDs that are members of this group.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the group was created.",
			},
		},
	}
}

func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or Name is provided (not both, not neither)
	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the group.",
		)
		return
	}

	if hasID && hasName {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot specify both 'id' and 'name'. Please specify only one.",
		)
		return
	}

	var group *client.Group
	var err error

	if hasID {
		// Lookup by ID
		group, err = d.client.GetGroup(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Group",
				fmt.Sprintf("Could not read group ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
	} else {
		// Lookup by name - list all groups and find matching name
		orgID := d.client.OrgID()
		if !data.OrgID.IsNull() {
			orgID = data.OrgID.ValueString()
		}

		listResp, err := d.client.ListGroups(ctx, &client.ListGroupsOptions{
			OrgID: orgID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Groups",
				fmt.Sprintf("Could not list groups to find name %s: %s", data.Name.ValueString(), err.Error()),
			)
			return
		}

		// Find group with matching name
		var found *client.Group
		for i := range listResp.Groups {
			if listResp.Groups[i].Name == data.Name.ValueString() {
				found = &listResp.Groups[i]
				break
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Group Not Found",
				fmt.Sprintf("No group found with name: %s", data.Name.ValueString()),
			)
			return
		}

		group = found
	}

	// Map response to data model
	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)
	data.OrgID = types.StringValue(group.OrgID)
	data.Description = types.StringValue(group.Description)
	data.Created = types.StringValue(group.Created)

	// Convert member IDs
	memberIDsList, diags := types.ListValueFrom(ctx, types.StringType, group.MemberIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.MemberIDs = memberIDsList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
