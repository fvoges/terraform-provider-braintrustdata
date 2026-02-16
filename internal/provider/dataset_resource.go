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
var _ resource.Resource = &DatasetResource{}
var _ resource.ResourceWithImportState = &DatasetResource{}

// NewDatasetResource creates a new dataset resource instance.
func NewDatasetResource() resource.Resource {
	return &DatasetResource{}
}

// DatasetResource defines the resource implementation.
type DatasetResource struct {
	client *client.Client
}

// DatasetResourceModel describes the resource data model.
type DatasetResourceModel struct {
	Metadata    types.Map    `tfsdk:"metadata"`
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	UserID      types.String `tfsdk:"user_id"`
	OrgID       types.String `tfsdk:"org_id"`
}

// Metadata implements resource.Resource.
func (r *DatasetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

// Schema implements resource.Resource.
func (r *DatasetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust dataset. Datasets are collections of examples used for testing and evaluation.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the dataset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project this dataset belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the dataset.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the dataset.",
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Metadata associated with the dataset as key-value pairs.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the dataset was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the dataset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the organization this dataset belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *DatasetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create implements resource.Resource by creating a new dataset.
func (r *DatasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatasetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert metadata from Terraform Map to Go map
	var metadata map[string]interface{}
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metadata = make(map[string]interface{})
		metadataMap := make(map[string]string)
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadataMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range metadataMap {
			metadata[k] = v
		}
	}

	// Create dataset via API
	dataset, err := r.client.CreateDataset(ctx, &client.CreateDatasetRequest{
		ProjectID:   data.ProjectID.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Metadata:    metadata,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create dataset, got error: %s", err))
		return
	}

	// Read the dataset to get the complete state
	dataset, err = r.client.GetDataset(ctx, dataset.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read dataset after creation, got error: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(dataset.ID)
	data.ProjectID = types.StringValue(dataset.ProjectID)
	data.Name = types.StringValue(dataset.Name)
	if dataset.Description != "" {
		data.Description = types.StringValue(dataset.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.Created = types.StringValue(dataset.Created)
	if dataset.UserID != "" {
		data.UserID = types.StringValue(dataset.UserID)
	} else {
		data.UserID = types.StringNull()
	}
	data.OrgID = types.StringValue(dataset.OrgID)

	// Convert metadata from Go map to Terraform Map
	if len(dataset.Metadata) > 0 {
		metadataStrings := make(map[string]string)
		for k, v := range dataset.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource by reading a dataset.
func (r *DatasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatasetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get dataset from API
	dataset, err := r.client.GetDataset(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read dataset, got error: %s", err))
		return
	}

	// Check for soft delete
	if dataset.DeletedAt != "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model with response data
	data.Name = types.StringValue(dataset.Name)
	if dataset.Description != "" {
		data.Description = types.StringValue(dataset.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.ProjectID = types.StringValue(dataset.ProjectID)
	data.Created = types.StringValue(dataset.Created)
	if dataset.UserID != "" {
		data.UserID = types.StringValue(dataset.UserID)
	} else {
		data.UserID = types.StringNull()
	}
	data.OrgID = types.StringValue(dataset.OrgID)

	// Convert metadata from Go map to Terraform Map
	if len(dataset.Metadata) > 0 {
		metadataStrings := make(map[string]string)
		for k, v := range dataset.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource by updating an existing dataset.
func (r *DatasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatasetResourceModel
	var state DatasetResourceModel

	// Get current state to preserve fields not returned by update API
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields are not unknown or null
	if data.ID.IsNull() || data.ID.IsUnknown() || data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Plan",
			"Cannot update dataset because id is unknown or empty.",
		)
		return
	}
	if data.Name.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid Plan",
			"Cannot update dataset because name is unknown.",
		)
		return
	}

	// Convert metadata from Terraform Map to Go map
	var metadata map[string]interface{}
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metadata = make(map[string]interface{})
		metadataMap := make(map[string]string)
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadataMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range metadataMap {
			metadata[k] = v
		}
	} else if data.Metadata.IsNull() {
		// Explicitly clear metadata by sending empty map.
		// The omitempty JSON tag only omits nil values, not empty maps.
		metadata = make(map[string]interface{})
	}
	// If IsUnknown, metadata remains nil and field is omitted from request

	// Update dataset via API
	dataset, err := r.client.UpdateDataset(ctx, data.ID.ValueString(), &client.UpdateDatasetRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Metadata:    metadata,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update dataset, got error: %s", err))
		return
	}

	// Update model with response data
	data.Name = types.StringValue(dataset.Name)
	if dataset.Description != "" {
		data.Description = types.StringValue(dataset.Description)
	} else {
		data.Description = types.StringNull()
	}

	// Convert metadata from Go map to Terraform Map
	if len(dataset.Metadata) > 0 {
		metadataStrings := make(map[string]string)
		for k, v := range dataset.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// Preserve fields from state that aren't returned by update API
	data.Created = state.Created
	data.ProjectID = state.ProjectID
	data.ID = state.ID
	data.UserID = state.UserID
	data.OrgID = state.OrgID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete implements resource.Resource by deleting a dataset.
func (r *DatasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatasetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete dataset via API
	err := r.client.DeleteDataset(ctx, data.ID.ValueString())

	if err != nil {
		// Treat 404 as success (already deleted) for idempotency
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete dataset, got error: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState by importing a dataset by ID.
func (r *DatasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
