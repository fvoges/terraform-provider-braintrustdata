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
var _ resource.Resource = &APIKeyResource{}
var _ resource.ResourceWithImportState = &APIKeyResource{}

// NewAPIKeyResource creates a new API key resource instance.
func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

// APIKeyResource defines the resource implementation.
type APIKeyResource struct {
	client *client.Client
}

// APIKeyResourceModel describes the resource data model.
type APIKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgID       types.String `tfsdk:"org_id"`
	PreviewName types.String `tfsdk:"preview_name"`
	UserID      types.String `tfsdk:"user_id"`
	UserEmail   types.String `tfsdk:"user_email"`
	Created     types.String `tfsdk:"created"`
	Key         types.String `tfsdk:"key"`
}

// Metadata implements resource.Resource.
func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// Schema implements resource.Resource.
func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust API key. API keys are used for authentication and inherit their user's permissions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the API key.",
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID that the API key belongs to.",
			},
			"preview_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The preview name of the API key.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the API key.",
			},
			"user_email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email of the user who created the API key.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the API key was created.",
			},
			"key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The actual API key value. This is only available when the key is first created and cannot be retrieved later.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create implements resource.Resource by creating a new API key.
func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API key via API
	apiKey, err := r.client.CreateAPIKey(ctx, &client.CreateAPIKeyRequest{
		Name: data.Name.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create API key, got error: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(apiKey.ID)
	data.OrgID = types.StringValue(apiKey.OrgID)
	data.PreviewName = types.StringValue(apiKey.PreviewName)
	data.Created = types.StringValue(apiKey.Created)
	if apiKey.UserID != "" {
		data.UserID = types.StringValue(apiKey.UserID)
	}
	if apiKey.UserEmail != "" {
		data.UserEmail = types.StringValue(apiKey.UserEmail)
	}
	if apiKey.Key != "" {
		data.Key = types.StringValue(apiKey.Key)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource by reading an API key.
func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data APIKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get API key from API
	apiKey, err := r.client.GetAPIKey(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API key, got error: %s", err))
		return
	}

	// Update model with response data
	data.Name = types.StringValue(apiKey.Name)
	data.PreviewName = types.StringValue(apiKey.PreviewName)
	data.OrgID = types.StringValue(apiKey.OrgID)
	data.Created = types.StringValue(apiKey.Created)
	if apiKey.UserID != "" {
		data.UserID = types.StringValue(apiKey.UserID)
	}
	if apiKey.UserEmail != "" {
		data.UserEmail = types.StringValue(apiKey.UserEmail)
	}
	// Note: Key is not returned by GET, preserve existing value in state

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource by updating an existing API key.
func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data APIKeyResourceModel
	var state APIKeyResourceModel

	// Get current state to preserve computed fields
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API key via API
	apiKey, err := r.client.UpdateAPIKey(ctx, data.ID.ValueString(), &client.UpdateAPIKeyRequest{
		Name: data.Name.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update API key, got error: %s", err))
		return
	}

	// Update model with response data
	data.Name = types.StringValue(apiKey.Name)
	data.PreviewName = types.StringValue(apiKey.PreviewName)

	// Preserve computed fields from state
	data.Created = state.Created
	data.OrgID = state.OrgID
	data.ID = state.ID
	data.UserID = state.UserID
	data.UserEmail = state.UserEmail
	data.Key = state.Key // Key is only available at creation time

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete implements resource.Resource by deleting an API key.
func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data APIKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API key via API
	err := r.client.DeleteAPIKey(ctx, data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete API key, got error: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState by importing an API key by ID.
func (r *APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
