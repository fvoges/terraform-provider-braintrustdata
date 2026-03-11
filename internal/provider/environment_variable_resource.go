package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EnvironmentVariableResource{}
var _ resource.ResourceWithImportState = &EnvironmentVariableResource{}

// NewEnvironmentVariableResource creates a new environment variable resource instance.
func NewEnvironmentVariableResource() resource.Resource {
	return &EnvironmentVariableResource{}
}

// EnvironmentVariableResource defines the resource implementation.
type EnvironmentVariableResource struct {
	client *client.Client
}

// EnvironmentVariableResourceModel describes the resource data model.
type EnvironmentVariableResourceModel struct {
	Metadata       types.Map    `tfsdk:"metadata"`
	ID             types.String `tfsdk:"id"`
	ObjectType     types.String `tfsdk:"object_type"`
	ObjectID       types.String `tfsdk:"object_id"`
	Name           types.String `tfsdk:"name"`
	Value          types.String `tfsdk:"value"`
	Description    types.String `tfsdk:"description"`
	SecretType     types.String `tfsdk:"secret_type"`
	SecretCategory types.String `tfsdk:"secret_category"`
	Created        types.String `tfsdk:"created"`
	Used           types.Bool   `tfsdk:"used"`
}

// Metadata implements resource.Resource.
func (r *EnvironmentVariableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

// Schema implements resource.Resource.
func (r *EnvironmentVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust environment variable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the environment variable.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object type that owns the environment variable (for example `project` or `function`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object ID that owns the environment variable.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The environment variable name.",
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The environment variable value. Required on create, optional afterwards for import and drift-safe reads. Omitting `value` after creation preserves the prior state value and does not clear the remote secret.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description associated with the environment variable, returned by Braintrust.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional metadata associated with the environment variable as key-value pairs.",
			},
			"secret_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional secret type hint returned by Braintrust.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_category": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional secret category hint returned by Braintrust.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the environment variable was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"used": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the environment variable has been used.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *EnvironmentVariableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *EnvironmentVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EnvironmentVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateEnvironmentVariableRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdEnvVar, err := r.client.CreateEnvironmentVariable(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create environment variable, got error: %s", err),
		)
		return
	}

	fetchedEnvVar, err := r.client.GetEnvironmentVariable(ctx, createdEnvVar.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read environment variable after creation, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setEnvironmentVariableResourceModel(ctx, &data, fetchedEnvVar)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *EnvironmentVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnvironmentVariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envVar, err := r.client.GetEnvironmentVariable(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read environment variable, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setEnvironmentVariableResourceModel(ctx, &data, envVar)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *EnvironmentVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentVariableResourceModel
	var state EnvironmentVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateEnvironmentVariableRequest(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasEnvironmentVariableUpdateChanges(updateReq) {
		if _, err := r.client.UpdateEnvironmentVariable(ctx, state.ID.ValueString(), updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update environment variable, got error: %s", err),
			)
			return
		}
	}

	updatedEnvVar, err := r.client.GetEnvironmentVariable(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read environment variable after update, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setEnvironmentVariableResourceModel(ctx, &plan, updatedEnvVar)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource.
func (r *EnvironmentVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EnvironmentVariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEnvironmentVariable(ctx, data.ID.ValueString())
	if err != nil {
		// Treat 404 as success for idempotent destroy.
		if client.IsNotFound(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete environment variable, got error: %s", err),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState by importing an environment variable by ID.
func (r *EnvironmentVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildCreateEnvironmentVariableRequest(ctx context.Context, data EnvironmentVariableResourceModel) (*client.CreateEnvironmentVariableRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := strings.TrimSpace(data.Value.ValueString())
	// Validate emptiness using trimmed whitespace, but preserve the original secret bytes in the API request.
	if data.Value.IsNull() || data.Value.IsUnknown() || value == "" {
		diags.AddAttributeError(
			path.Root("value"),
			"Invalid value",
			"'value' must be provided and non-empty when creating an environment variable.",
		)
		return nil, diags
	}

	metadata, metadataDiags := environmentVariableMetadataFromTerraformMap(ctx, data.Metadata)
	diags.Append(metadataDiags...)
	if diags.HasError() {
		return nil, diags
	}

	createReq := &client.CreateEnvironmentVariableRequest{
		ObjectType: strings.TrimSpace(data.ObjectType.ValueString()),
		ObjectID:   strings.TrimSpace(data.ObjectID.ValueString()),
		Name:       strings.TrimSpace(data.Name.ValueString()),
		Value:      data.Value.ValueString(),
	}

	if metadata != nil {
		createReq.Metadata = metadata
	}
	if !data.SecretType.IsNull() && !data.SecretType.IsUnknown() {
		createReq.SecretType = data.SecretType.ValueString()
	}
	if !data.SecretCategory.IsNull() && !data.SecretCategory.IsUnknown() {
		createReq.SecretCategory = data.SecretCategory.ValueString()
	}

	return createReq, diags
}

func buildUpdateEnvironmentVariableRequest(
	ctx context.Context,
	plan EnvironmentVariableResourceModel,
	state EnvironmentVariableResourceModel,
) (*client.UpdateEnvironmentVariableRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	updateReq := &client.UpdateEnvironmentVariableRequest{}

	if !plan.Name.Equal(state.Name) {
		updateReq.Name = environmentVariableStringPtr(plan.Name.ValueString())
	}

	if !plan.Value.IsUnknown() && !plan.Value.Equal(state.Value) {
		if !plan.Value.IsNull() {
			updateReq.Value = environmentVariableStringPtr(plan.Value.ValueString())
		}
	}

	if !plan.Metadata.IsUnknown() && !plan.Metadata.Equal(state.Metadata) {
		if plan.Metadata.IsNull() {
			updateReq.Metadata = environmentVariableMapPtr(map[string]interface{}{})
		} else {
			metadata, metadataDiags := environmentVariableMetadataFromTerraformMap(ctx, plan.Metadata)
			diags.Append(metadataDiags...)
			if diags.HasError() {
				return nil, diags
			}
			updateReq.Metadata = environmentVariableMapPtr(metadata)
		}
	}

	if !plan.SecretType.IsUnknown() && !plan.SecretType.Equal(state.SecretType) {
		secretType := ""
		if !plan.SecretType.IsNull() {
			secretType = plan.SecretType.ValueString()
		}
		updateReq.SecretType = environmentVariableStringPtr(secretType)
	}

	if !plan.SecretCategory.IsUnknown() && !plan.SecretCategory.Equal(state.SecretCategory) {
		secretCategory := ""
		if !plan.SecretCategory.IsNull() {
			secretCategory = plan.SecretCategory.ValueString()
		}
		updateReq.SecretCategory = environmentVariableStringPtr(secretCategory)
	}

	return updateReq, diags
}

func hasEnvironmentVariableUpdateChanges(req *client.UpdateEnvironmentVariableRequest) bool {
	return req.Name != nil ||
		req.Value != nil ||
		req.Metadata != nil ||
		req.SecretType != nil ||
		req.SecretCategory != nil
}

func setEnvironmentVariableResourceModel(
	ctx context.Context,
	data *EnvironmentVariableResourceModel,
	envVar *client.EnvironmentVariable,
) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(envVar.ID)
	data.ObjectType = stringOrNull(envVar.ObjectType)
	data.ObjectID = stringOrNull(envVar.ObjectID)
	data.Name = types.StringValue(envVar.Name)
	data.Used = types.BoolValue(bool(envVar.Used))
	data.Created = stringOrNull(envVar.Created)

	if envVar.Description != "" {
		data.Description = types.StringValue(envVar.Description)
	} else if data.Description.IsNull() || data.Description.IsUnknown() {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue("")
	}

	if envVar.SecretType != "" {
		data.SecretType = types.StringValue(envVar.SecretType)
	} else if data.SecretType.IsNull() || data.SecretType.IsUnknown() {
		data.SecretType = types.StringNull()
	} else {
		data.SecretType = types.StringValue("")
	}

	if envVar.SecretCategory != "" {
		data.SecretCategory = types.StringValue(envVar.SecretCategory)
	} else if data.SecretCategory.IsNull() || data.SecretCategory.IsUnknown() {
		data.SecretCategory = types.StringNull()
	} else {
		data.SecretCategory = types.StringValue("")
	}

	metadataValue, metadataDiags := environmentVariableMetadataToTerraformMap(ctx, envVar.Metadata, data.Metadata)
	diags.Append(metadataDiags...)
	if diags.HasError() {
		return diags
	}
	data.Metadata = metadataValue

	// API does not return value on reads for security reasons.
	// Preserve existing state unless API explicitly returns a replacement value.
	if envVar.Value != "" {
		data.Value = types.StringValue(envVar.Value)
	} else if data.Value.IsNull() {
		data.Value = types.StringNull()
	}

	return diags
}

func environmentVariableMetadataFromTerraformMap(ctx context.Context, metadataMap types.Map) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if metadataMap.IsNull() || metadataMap.IsUnknown() {
		return nil, diags
	}

	metadataStrings := make(map[string]string)
	diags.Append(metadataMap.ElementsAs(ctx, &metadataStrings, false)...)
	if diags.HasError() {
		return nil, diags
	}

	metadata := make(map[string]interface{}, len(metadataStrings))
	for key, value := range metadataStrings {
		metadata[key] = value
	}

	return metadata, diags
}

func environmentVariableMetadataToTerraformMap(ctx context.Context, metadata map[string]interface{}, prior types.Map) (types.Map, diag.Diagnostics) {
	if len(metadata) > 0 {
		metadataStrings := make(map[string]string, len(metadata))
		for key, value := range metadata {
			metadataStrings[key] = fmt.Sprintf("%v", value)
		}

		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		if diags.HasError() {
			return types.MapNull(types.StringType), diags
		}

		return metadataValue, diags
	}

	if prior.IsNull() || prior.IsUnknown() {
		return types.MapNull(types.StringType), nil
	}

	return types.MapValueMust(types.StringType, map[string]attr.Value{}), nil
}

func environmentVariableStringPtr(value string) *string {
	return &value
}

func environmentVariableMapPtr(value map[string]interface{}) *map[string]interface{} {
	return &value
}
