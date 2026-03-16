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

var _ resource.Resource = &TagResource{}
var _ resource.ResourceWithImportState = &TagResource{}

// NewTagResource creates a new tag resource instance.
func NewTagResource() resource.Resource {
	return &TagResource{}
}

// TagResource defines the resource implementation.
type TagResource struct {
	client *client.Client
}

// TagResourceModel describes the resource data model.
type TagResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Color       types.String `tfsdk:"color"`
	UserID      types.String `tfsdk:"user_id"`
	Created     types.String `tfsdk:"created"`
	Position    types.String `tfsdk:"position"`
}

// Metadata implements resource.Resource.
func (r *TagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

// Schema implements resource.Resource.
func (r *TagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust project tag.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the tag.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project ID that owns the tag.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The tag name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The tag description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"color": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Color of the tag for UI display.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the tag.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the tag was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"position": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "LexoRank position of the tag within the project.",
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *TagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// Create implements resource.Resource.
func (r *TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateTagRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.client.CreateTag(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create tag, got error: %s", err),
		)
		return
	}

	setTagResourceModel(ctx, &data, tag)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.client.GetTag(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read tag, got error: %s", err),
		)
		return
	}

	setTagResourceModel(ctx, &data, tag)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TagResourceModel
	var state TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateTagRequest(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasTagUpdateChanges(updateReq) {
		if _, err := r.client.UpdateTag(ctx, state.ID.ValueString(), updateReq); err != nil {
			if client.IsNotFound(err) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update tag, got error: %s", err),
			)
			return
		}
	}

	tag, err := r.client.GetTag(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read tag after update, got error: %s", err),
		)
		return
	}

	setTagResourceModel(ctx, &plan, tag)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource.
func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTag(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete tag, got error: %s", err),
		)
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildCreateTagRequest(_ context.Context, model TagResourceModel) (*client.CreateTagRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := &client.CreateTagRequest{
		ProjectID: model.ProjectID.ValueString(),
		Name:      model.Name.ValueString(),
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		req.Description = model.Description.ValueString()
	}
	if !model.Color.IsNull() && !model.Color.IsUnknown() {
		req.Color = model.Color.ValueString()
	}

	return req, diags
}

func buildUpdateTagRequest(_ context.Context, plan, state TagResourceModel) (*client.UpdateTagRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.UpdateTagRequest{}

	if !plan.Name.IsUnknown() && !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		req.Name = &v
	}

	if !plan.Description.IsUnknown() {
		if state.Description.IsNull() {
			if !plan.Description.IsNull() {
				v := plan.Description.ValueString()
				req.Description = &v
			}
		} else {
			if plan.Description.IsNull() {
				diags.AddError("Cannot Clear Description", "Description cannot be cleared after it has been set.")
				return nil, diags
			}
			if !plan.Description.Equal(state.Description) {
				v := plan.Description.ValueString()
				req.Description = &v
			}
		}
	}

	if !plan.Color.IsUnknown() {
		if state.Color.IsNull() {
			if !plan.Color.IsNull() {
				v := plan.Color.ValueString()
				req.Color = &v
			}
		} else {
			if plan.Color.IsNull() {
				diags.AddError("Cannot Clear Color", "Color cannot be cleared after it has been set.")
				return nil, diags
			}
			if !plan.Color.Equal(state.Color) {
				v := plan.Color.ValueString()
				req.Color = &v
			}
		}
	}

	return req, diags
}

func hasTagUpdateChanges(req *client.UpdateTagRequest) bool {
	return req.Name != nil || req.Description != nil || req.Color != nil
}

func setTagResourceModel(_ context.Context, model *TagResourceModel, tag *client.Tag) {
	model.ID = stringOrNull(tag.ID)
	model.ProjectID = stringOrNull(tag.ProjectID)
	model.Name = stringOrNull(tag.Name)
	model.Description = stringOrNull(tag.Description)
	model.Color = stringOrNull(tag.Color)
	model.UserID = stringOrNull(tag.UserID)
	model.Created = stringOrNull(tag.Created)

	if tag.Position != nil {
		model.Position = stringOrNull(*tag.Position)
	} else {
		model.Position = types.StringNull()
	}
}
