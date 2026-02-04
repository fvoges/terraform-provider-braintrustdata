package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}

// NewGroupResource creates a new group resource instance.
func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// GroupResource defines the resource implementation.
type GroupResource struct {
	client *client.Client
}

// GroupResourceModel describes the resource data model.
type GroupResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	OrgID        types.String `tfsdk:"org_id"`
	Description  types.String `tfsdk:"description"`
	MemberUsers  types.List   `tfsdk:"member_users"`
	MemberGroups types.List   `tfsdk:"member_groups"`
	Created      types.String `tfsdk:"created"`
}

// Metadata implements resource.Resource.
func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema implements resource.Resource.
func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust group. Groups are collections of users that can be assigned permissions via ACLs.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the group.",
			},
			"org_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The organization ID. Defaults to the provider's organization_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the group.",
			},
			"member_users": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of user IDs that are members of this group.",
			},
			"member_groups": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of group IDs that are members of this group.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the group was created.",
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create implements resource.Resource by creating a new group.
func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use provider's org ID if not specified
	orgID := r.client.OrgID()
	if !data.OrgID.IsNull() {
		orgID = data.OrgID.ValueString()
	}

	// Extract member lists from plan
	var memberUsers []string
	var memberGroups []string
	if !data.MemberUsers.IsNull() {
		resp.Diagnostics.Append(data.MemberUsers.ElementsAs(ctx, &memberUsers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.MemberGroups.IsNull() {
		resp.Diagnostics.Append(data.MemberGroups.ElementsAs(ctx, &memberGroups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create group via API with members
	group, err := r.client.CreateGroup(ctx, &client.CreateGroupRequest{
		Name:         data.Name.ValueString(),
		OrgID:        orgID,
		Description:  data.Description.ValueString(),
		MemberUsers:  memberUsers,
		MemberGroups: memberGroups,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create group, got error: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(group.ID)
	data.OrgID = types.StringValue(group.OrgID)
	data.Created = types.StringValue(group.Created)

	// Update member lists from API response
	if len(group.MemberUsers) > 0 {
		memberUsersList, diags := types.ListValueFrom(ctx, types.StringType, group.MemberUsers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.MemberUsers = memberUsersList
	} else {
		data.MemberUsers = types.ListNull(types.StringType)
	}

	if len(group.MemberGroups) > 0 {
		memberGroupsList, diags := types.ListValueFrom(ctx, types.StringType, group.MemberGroups)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.MemberGroups = memberGroupsList
	} else {
		data.MemberGroups = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get group from API
	group, err := r.client.GetGroup(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	// Check for soft delete
	if group.DeletedAt != "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model with response data
	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)
	data.OrgID = types.StringValue(group.OrgID)
	data.Created = types.StringValue(group.Created)

	// Convert member lists to Terraform lists
	if len(group.MemberUsers) > 0 {
		memberUsersList, diags := types.ListValueFrom(ctx, types.StringType, group.MemberUsers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.MemberUsers = memberUsersList
	} else {
		data.MemberUsers = types.ListNull(types.StringType)
	}

	if len(group.MemberGroups) > 0 {
		memberGroupsList, diags := types.ListValueFrom(ctx, types.StringType, group.MemberGroups)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.MemberGroups = memberGroupsList
	} else {
		data.MemberGroups = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource by updating an existing group.
func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	var state GroupResourceModel

	// Get current state to preserve fields not returned by update API
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert member lists from Terraform to string slices
	var memberUsers []string
	var memberGroups []string
	if !data.MemberUsers.IsNull() {
		resp.Diagnostics.Append(data.MemberUsers.ElementsAs(ctx, &memberUsers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.MemberGroups.IsNull() {
		resp.Diagnostics.Append(data.MemberGroups.ElementsAs(ctx, &memberGroups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update group via API
	group, err := r.client.UpdateGroup(ctx, data.ID.ValueString(), &client.UpdateGroupRequest{
		Name:         data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		MemberUsers:  memberUsers,
		MemberGroups: memberGroups,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update group, got error: %s", err))
		return
	}

	// Update model with response data
	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)

	// Preserve fields from state that aren't returned by update API
	data.Created = state.Created
	data.OrgID = state.OrgID
	data.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete implements resource.Resource by deleting a group.
func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete group via API
	err := r.client.DeleteGroup(ctx, data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete group, got error: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState by importing a group by ID.
func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
