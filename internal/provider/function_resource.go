package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FunctionResource{}
var _ resource.ResourceWithImportState = &FunctionResource{}

// NewFunctionResource creates a new function resource instance.
func NewFunctionResource() resource.Resource {
	return &FunctionResource{}
}

// FunctionResource defines the resource implementation.
type FunctionResource struct {
	client *client.Client
}

// FunctionResourceModel describes the resource data model.
type FunctionResourceModel struct {
	Metadata       types.Map    `tfsdk:"metadata"`
	Tags           types.Set    `tfsdk:"tags"`
	XactID         types.String `tfsdk:"xact_id"`
	Created        types.String `tfsdk:"created"`
	Description    types.String `tfsdk:"description"`
	FunctionData   types.String `tfsdk:"function_data"`
	FunctionSchema types.String `tfsdk:"function_schema"`
	FunctionType   types.String `tfsdk:"function_type"`
	ID             types.String `tfsdk:"id"`
	LogID          types.String `tfsdk:"log_id"`
	Name           types.String `tfsdk:"name"`
	OrgID          types.String `tfsdk:"org_id"`
	Origin         types.String `tfsdk:"origin"`
	ProjectID      types.String `tfsdk:"project_id"`
	PromptData     types.String `tfsdk:"prompt_data"`
	Slug           types.String `tfsdk:"slug"`
}

// Metadata implements resource.Resource.
func (r *FunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

// Schema implements resource.Resource.
func (r *FunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust function.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project ID that owns the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The function name.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The function slug. Defaults to a slugified form of `name` when omitted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A description of the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"function_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The function type, such as `tool`, `scorer`, or `workflow`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"function_data": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The function data as a JSON-encoded string. Use `jsonencode()` for structured content.",
			},
			"function_schema": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The function schema as a JSON-encoded string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prompt_data": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Prompt data for prompt-backed functions as a JSON-encoded string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Metadata associated with the function as key-value pairs.",
			},
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Tags associated with the function.",
			},
			"xact_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The transaction ID associated with the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the function was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"log_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The log ID associated with the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID associated with the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"origin": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The function origin as a JSON-encoded string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *FunctionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FunctionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateFunctionRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdFunction, err := r.client.CreateFunction(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create function, got error: %s", err),
		)
		return
	}

	fetchedFunction, err := r.client.GetFunction(ctx, createdFunction.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read function after creation, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setFunctionResourceModel(ctx, &data, fetchedFunction)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	function, err := r.client.GetFunction(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsFunctionNotFound(err) || client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read function, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setFunctionResourceModel(ctx, &data, function)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *FunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FunctionResourceModel
	var state FunctionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateFunctionRequest(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasFunctionUpdateChanges(updateReq) {
		if _, err := r.client.UpdateFunction(ctx, state.ID.ValueString(), updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update function, got error: %s", err),
			)
			return
		}
	}

	updatedFunction, err := r.client.GetFunction(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsFunctionNotFound(err) || client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read function after update, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setFunctionResourceModel(ctx, &plan, updatedFunction)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource.
func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFunction(ctx, data.ID.ValueString()); err != nil {
		if client.IsFunctionNotFound(err) || client.IsNotFound(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete function, got error: %s", err),
		)
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *FunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildCreateFunctionRequest(ctx context.Context, data FunctionResourceModel) (*client.CreateFunctionRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	functionData, err := decodeFunctionJSONField("function_data", data.FunctionData)
	if err != nil {
		diags.AddError("Invalid function_data", err.Error())
		return nil, diags
	}

	functionSchema, schemaDiags := optionalFunctionJSONField("function_schema", data.FunctionSchema)
	diags.Append(schemaDiags...)
	if diags.HasError() {
		return nil, diags
	}

	promptData, promptDiags := optionalFunctionJSONField("prompt_data", data.PromptData)
	diags.Append(promptDiags...)
	if diags.HasError() {
		return nil, diags
	}

	metadata, metaDiags := extractMetadata(ctx, data.Metadata)
	diags.Append(metaDiags...)
	if diags.HasError() {
		return nil, diags
	}

	tags, tagDiags := extractTags(ctx, data.Tags)
	diags.Append(tagDiags...)
	if diags.HasError() {
		return nil, diags
	}

	slug := strings.TrimSpace(data.Slug.ValueString())
	if slug == "" {
		slug = slugify(data.Name.ValueString())
	}

	createReq := &client.CreateFunctionRequest{
		ProjectID:      data.ProjectID.ValueString(),
		Name:           data.Name.ValueString(),
		Slug:           slug,
		Description:    data.Description.ValueString(),
		FunctionType:   data.FunctionType.ValueString(),
		FunctionData:   functionData,
		FunctionSchema: functionSchema,
		PromptData:     promptData,
		Metadata:       metadata,
		Tags:           tags,
	}

	return createReq, diags
}

func buildUpdateFunctionRequest(ctx context.Context, plan, state FunctionResourceModel) (*client.UpdateFunctionRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.UpdateFunctionRequest{}

	if !plan.Name.Equal(state.Name) {
		req.Name = functionStringPtr(plan.Name.ValueString())
	}

	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() && !plan.Slug.Equal(state.Slug) {
		req.Slug = functionStringPtr(plan.Slug.ValueString())
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && !plan.Description.Equal(state.Description) {
		req.Description = functionStringPtr(plan.Description.ValueString())
	}

	if !plan.FunctionType.IsNull() && !plan.FunctionType.IsUnknown() && !plan.FunctionType.Equal(state.FunctionType) {
		req.FunctionType = functionStringPtr(plan.FunctionType.ValueString())
	}

	if !plan.FunctionData.Equal(state.FunctionData) {
		functionData, err := decodeFunctionJSONField("function_data", plan.FunctionData)
		if err != nil {
			diags.AddError("Invalid function_data", err.Error())
			return nil, diags
		}
		req.FunctionData = functionInterfacePtr(functionData)
	}

	if !plan.FunctionSchema.IsNull() && !plan.FunctionSchema.IsUnknown() && !plan.FunctionSchema.Equal(state.FunctionSchema) {
		functionSchema, fieldDiags := optionalFunctionJSONField("function_schema", plan.FunctionSchema)
		diags.Append(fieldDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.FunctionSchema = functionInterfacePtr(functionSchema)
	}

	if !plan.PromptData.IsNull() && !plan.PromptData.IsUnknown() && !plan.PromptData.Equal(state.PromptData) {
		promptData, fieldDiags := optionalFunctionJSONField("prompt_data", plan.PromptData)
		diags.Append(fieldDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.PromptData = functionInterfacePtr(promptData)
	}

	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() && !plan.Metadata.Equal(state.Metadata) {
		metadata, metaDiags := extractMetadata(ctx, plan.Metadata)
		diags.Append(metaDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Metadata = functionMapPtr(metadata)
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() && !plan.Tags.Equal(state.Tags) {
		tags, tagDiags := extractTags(ctx, plan.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Tags = functionStringSlicePtr(tags)
	}

	return req, diags
}

func hasFunctionUpdateChanges(req *client.UpdateFunctionRequest) bool {
	return req.Name != nil ||
		req.Slug != nil ||
		req.Description != nil ||
		req.FunctionType != nil ||
		req.FunctionData != nil ||
		req.FunctionSchema != nil ||
		req.PromptData != nil ||
		req.Metadata != nil ||
		req.Tags != nil ||
		req.Origin != nil
}

func setFunctionResourceModel(ctx context.Context, data *FunctionResourceModel, function *client.Function) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = stringOrNull(function.ID)
	data.ProjectID = stringOrNull(function.ProjectID)
	data.Name = stringOrNull(function.Name)
	data.Slug = stringOrNull(function.Slug)
	data.Description = stringOrNull(function.Description)
	data.FunctionType = stringOrNull(function.FunctionType)
	data.XactID = stringOrNull(function.XactID)
	data.Created = stringOrNull(function.Created)
	data.LogID = stringOrNull(function.LogID)
	data.OrgID = stringOrNull(function.OrgID)

	functionData, functionDataDiags := jsonEncodedOrNull("function_data", function.FunctionData)
	diags.Append(functionDataDiags...)
	data.FunctionData = functionData

	functionSchema, functionSchemaDiags := jsonEncodedOrNull("function_schema", function.FunctionSchema)
	diags.Append(functionSchemaDiags...)
	data.FunctionSchema = functionSchema

	origin, originDiags := jsonEncodedOrNull("origin", function.Origin)
	diags.Append(originDiags...)
	data.Origin = origin

	promptData, promptDataDiags := jsonEncodedOrNull("prompt_data", function.PromptData)
	diags.Append(promptDataDiags...)
	data.PromptData = promptData

	if diags.HasError() {
		return diags
	}

	if len(function.Metadata) > 0 {
		metadata := make(map[string]string, len(function.Metadata))
		for k, v := range function.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}

		metadataMap, metadataDiags := types.MapValueFrom(ctx, types.StringType, metadata)
		diags.Append(metadataDiags...)
		if diags.HasError() {
			return diags
		}
		data.Metadata = metadataMap
	} else if data.Metadata.IsNull() {
		data.Metadata = types.MapNull(types.StringType)
	} else {
		data.Metadata = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	if len(function.Tags) > 0 {
		tagsSet, tagsDiags := types.SetValueFrom(ctx, types.StringType, function.Tags)
		diags.Append(tagsDiags...)
		if diags.HasError() {
			return diags
		}
		data.Tags = tagsSet
	} else if data.Tags.IsNull() {
		data.Tags = types.SetNull(types.StringType)
	} else {
		data.Tags = types.SetValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

func decodeFunctionJSONField(fieldName string, value types.String) (interface{}, error) {
	if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, fmt.Errorf("%s must be valid JSON and cannot be empty", fieldName)
	}

	var decoded interface{}
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		return nil, fmt.Errorf("%s must be valid JSON: %w", fieldName, err)
	}

	return decoded, nil
}

func optionalFunctionJSONField(fieldName string, value types.String) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	decoded, err := decodeFunctionJSONField(fieldName, value)
	if err != nil {
		diags.AddError(fmt.Sprintf("Invalid %s", fieldName), err.Error())
		return nil, diags
	}

	return decoded, diags
}

func functionStringPtr(value string) *string {
	return &value
}

func functionStringSlicePtr(values []string) *[]string {
	return &values
}

func functionMapPtr(values map[string]interface{}) *map[string]interface{} {
	return &values
}

func functionInterfacePtr(value interface{}) *interface{} {
	return &value
}
