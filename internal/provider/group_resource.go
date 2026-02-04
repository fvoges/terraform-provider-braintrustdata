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

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// GroupResource defines the resource implementation.
type GroupResource struct {
	client *client.Client
}

// GroupResourceModel describes the resource data model.
type GroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgID       types.String `tfsdk:"org_id"`
	Description types.String `tfsdk:"description"`
	MemberIDs   types.List   `tfsdk:"member_ids"`
	Created     types.String `tfsdk:"created"`
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"member_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of user IDs that are members of this group.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the group was created.",
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	// Convert member IDs from Terraform list to string slice
	var memberIDs []string
	if !data.MemberIDs.IsNull() {
		resp.Diagnostics.Append(data.MemberIDs.ElementsAs(ctx, &memberIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create group via API
	group, err := r.client.CreateGroup(ctx, &client.CreateGroupRequest{
		Name:        data.Name.ValueString(),
		OrgID:       orgID,
		Description: data.Description.ValueString(),
		MemberIDs:   memberIDs,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create group, got error: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(group.ID)
	data.OrgID = types.StringValue(group.OrgID)
	data.Created = types.StringValue(group.Created)

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

	// Convert member IDs to Terraform list
	if len(group.MemberIDs) > 0 {
		memberList, diags := types.ListValueFrom(ctx, types.StringType, group.MemberIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.MemberIDs = memberList
	} else {
		data.MemberIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	var state GroupResourceModel

	// Get current state to preserve fields not returned by update API
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert member IDs from Terraform list to string slice
	var memberIDs []string
	if !data.MemberIDs.IsNull() {
		resp.Diagnostics.Append(data.MemberIDs.ElementsAs(ctx, &memberIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update group via API
	group, err := r.client.UpdateGroup(ctx, data.ID.ValueString(), &client.UpdateGroupRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		MemberIDs:   memberIDs,
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

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
