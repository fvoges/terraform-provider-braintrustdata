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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PromptResource{}
var _ resource.ResourceWithImportState = &PromptResource{}

// NewPromptResource creates a new prompt resource instance.
func NewPromptResource() resource.Resource {
	return &PromptResource{}
}

// PromptResource defines the resource implementation.
type PromptResource struct {
	client *client.Client
}

// PromptResourceModel describes the resource data model.
type PromptResourceModel struct {
	Tags         types.Set    `tfsdk:"tags"`
	Metadata     types.Map    `tfsdk:"metadata"`
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	Description  types.String `tfsdk:"description"`
	FunctionType types.String `tfsdk:"function_type"`
	PromptData   types.String `tfsdk:"prompt_data"`
	Created      types.String `tfsdk:"created"`
	UserID       types.String `tfsdk:"user_id"`
	OrgID        types.String `tfsdk:"org_id"`
}

// Metadata implements resource.Resource.
func (r *PromptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

// Schema implements resource.Resource.
func (r *PromptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust prompt. Prompts are versioned templates used to query AI models.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the prompt.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project this prompt belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the prompt.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A URL-safe identifier for the prompt. Defaults to the name if not set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the prompt.",
			},
			"function_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The function type associated with the prompt.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prompt_data": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The prompt data as a JSON-encoded string. Use `jsonencode()` to supply structured prompt content.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Metadata associated with the prompt as key-value pairs.",
			},
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Tags associated with the prompt.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the prompt was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the prompt.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the organization this prompt belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *PromptResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create implements resource.Resource by creating a new prompt.
func (r *PromptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PromptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreatePromptRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	prompt, err := r.client.CreatePrompt(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create prompt, got error: %s", err))
		return
	}

	// Read back to get complete state.
	prompt, err = r.client.GetPrompt(ctx, prompt.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prompt after creation, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(setPromptResourceModel(ctx, &data, prompt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource by reading a prompt.
func (r *PromptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PromptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	prompt, err := r.client.GetPrompt(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prompt, got error: %s", err))
		return
	}

	// Treat soft deletes as removed.
	if prompt.DeletedAt != "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(setPromptResourceModel(ctx, &data, prompt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource by updating an existing prompt.
func (r *PromptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PromptResourceModel
	var state PromptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() || data.ID.IsUnknown() || data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Plan",
			"Cannot update prompt because id is unknown or empty.",
		)
		return
	}

	updateReq, diags := buildUpdatePromptRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	prompt, err := r.client.UpdatePrompt(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update prompt, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(setPromptResourceModel(ctx, &data, prompt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve fields not returned by update API.
	data.ID = state.ID
	data.ProjectID = state.ProjectID
	data.Created = state.Created
	if data.OrgID.IsNull() || data.OrgID.IsUnknown() {
		data.OrgID = state.OrgID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete implements resource.Resource by deleting a prompt.
func (r *PromptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PromptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePrompt(ctx, data.ID.ValueString())

	if err != nil {
		// Treat 404 as success (already deleted) for idempotency.
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete prompt, got error: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState by importing a prompt by ID.
func (r *PromptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildCreatePromptRequest converts a Terraform model to a CreatePromptRequest.
func buildCreatePromptRequest(ctx context.Context, data PromptResourceModel) (*client.CreatePromptRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

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

	promptData, err := decodePromptData(data.PromptData)
	if err != nil {
		diags.AddError("Invalid prompt_data", fmt.Sprintf("prompt_data must be valid JSON: %s", err))
		return nil, diags
	}

	slug := data.Slug.ValueString()
	if slug == "" {
		slug = slugify(data.Name.ValueString())
	}

	return &client.CreatePromptRequest{
		ProjectID:    data.ProjectID.ValueString(),
		Name:         data.Name.ValueString(),
		Slug:         slug,
		Description:  data.Description.ValueString(),
		FunctionType: data.FunctionType.ValueString(),
		Metadata:     metadata,
		Tags:         tags,
		PromptData:   promptData,
	}, diags
}

// slugify converts a name to a URL-safe slug: lowercase, spaces to hyphens,
// non-alphanumeric-hyphen characters removed.
func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// buildUpdatePromptRequest converts a Terraform model to an UpdatePromptRequest.
func buildUpdatePromptRequest(ctx context.Context, data PromptResourceModel) (*client.UpdatePromptRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var metadata map[string]interface{}
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		m, metaDiags := extractMetadata(ctx, data.Metadata)
		diags.Append(metaDiags...)
		if diags.HasError() {
			return nil, diags
		}
		metadata = m
	} else if data.Metadata.IsNull() {
		// Explicitly clear metadata.
		metadata = make(map[string]interface{})
	}

	tags, tagDiags := extractTags(ctx, data.Tags)
	diags.Append(tagDiags...)
	if diags.HasError() {
		return nil, diags
	}

	promptData, err := decodePromptData(data.PromptData)
	if err != nil {
		diags.AddError("Invalid prompt_data", fmt.Sprintf("prompt_data must be valid JSON: %s", err))
		return nil, diags
	}

	return &client.UpdatePromptRequest{
		Name:         data.Name.ValueString(),
		Slug:         data.Slug.ValueString(),
		Description:  data.Description.ValueString(),
		FunctionType: data.FunctionType.ValueString(),
		Metadata:     metadata,
		Tags:         tags,
		PromptData:   promptData,
	}, diags
}

// setPromptResourceModel populates the resource model from an API prompt response.
func setPromptResourceModel(ctx context.Context, data *PromptResourceModel, prompt *client.Prompt) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(prompt.ID)
	data.ProjectID = types.StringValue(prompt.ProjectID)
	data.Name = types.StringValue(prompt.Name)
	data.Slug = stringOrNull(prompt.Slug)
	data.Description = stringOrNull(prompt.Description)
	data.FunctionType = stringOrNull(prompt.FunctionType)
	data.Created = stringOrNull(prompt.Created)
	data.OrgID = stringOrNull(prompt.OrgID)
	if prompt.UserID != "" {
		data.UserID = types.StringValue(prompt.UserID)
	} else {
		data.UserID = types.StringNull()
	}

	// prompt_data: normalize to canonical JSON to avoid perpetual diffs.
	if prompt.PromptData != nil {
		encoded, err := json.Marshal(prompt.PromptData)
		if err != nil {
			diags.AddError(
				"Error Encoding prompt_data",
				fmt.Sprintf("Unable to encode prompt_data as JSON: %s", err),
			)
			return diags
		}
		data.PromptData = types.StringValue(string(encoded))
	} else {
		data.PromptData = types.StringNull()
	}

	// metadata
	if len(prompt.Metadata) > 0 {
		metadataStrings := make(map[string]string)
		for k, v := range prompt.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, metaDiags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		diags.Append(metaDiags...)
		if diags.HasError() {
			return diags
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// tags: preserve plan/state intent for null vs empty set.
	// - Non-empty API tags → store as set.
	// - Empty/nil API tags + plan had null (tags omitted) → keep null.
	// - Empty/nil API tags + plan had empty set (tags = []) → store empty set.
	if len(prompt.Tags) > 0 {
		tagsSet, tagDiags := types.SetValueFrom(ctx, types.StringType, prompt.Tags)
		diags.Append(tagDiags...)
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

// extractMetadata converts a Terraform Map to a Go map.
func extractMetadata(ctx context.Context, m types.Map) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.IsNull() || m.IsUnknown() {
		return nil, diags
	}
	metadata := make(map[string]interface{})
	metadataMap := make(map[string]string)
	diags.Append(m.ElementsAs(ctx, &metadataMap, false)...)
	if diags.HasError() {
		return nil, diags
	}
	for k, v := range metadataMap {
		metadata[k] = v
	}
	return metadata, diags
}

// extractTags converts a Terraform Set to a Go string slice.
func extractTags(ctx context.Context, s types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if s.IsNull() || s.IsUnknown() {
		return nil, diags
	}
	var tags []string
	diags.Append(s.ElementsAs(ctx, &tags, false)...)
	return tags, diags
}

// decodePromptData unmarshals a JSON string into an interface{}.
// Returns nil (no error) when value is null/unknown.
func decodePromptData(v types.String) (interface{}, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	var result interface{}
	if err := json.Unmarshal([]byte(v.ValueString()), &result); err != nil {
		return nil, err
	}
	return result, nil
}
