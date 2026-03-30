package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AISecretResource{}
var _ resource.ResourceWithImportState = &AISecretResource{}

// NewAISecretResource creates a new AI secret resource instance.
func NewAISecretResource() resource.Resource {
	return &AISecretResource{}
}

// AISecretResource defines the resource implementation.
type AISecretResource struct {
	client *client.Client
}

// AISecretResourceModel describes the resource data model.
type AISecretResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	Metadata      types.Map    `tfsdk:"metadata"`
	Secret        types.String `tfsdk:"secret"`
	OrgName       types.String `tfsdk:"org_name"`
	OrgID         types.String `tfsdk:"org_id"`
	PreviewSecret types.String `tfsdk:"preview_secret"`
	Created       types.String `tfsdk:"created"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

// Metadata implements resource.Resource.
func (r *AISecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_secret"
}

// Schema implements resource.Resource.
func (r *AISecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust AI secret with a write-only secret contract.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the AI secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The AI secret name.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The AI secret type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional metadata associated with the AI secret as key-value pairs.",
			},
			"secret": schema.StringAttribute{
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The AI secret value. Required on create. Braintrust omits the raw secret on read and import, so the provider preserves any prior state value when available but cannot recover it from the API. To rotate the secret, set `secret` explicitly in configuration. Omitting it after creation leaves the existing remote secret unchanged; clearing or removing it is not supported. Leading/trailing whitespace on a non-empty secret is preserved, but whitespace-only values are rejected. Terraform stores this sensitive value in state (redacted in Terraform UI output).",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name used when creating the AI secret in multi-org contexts.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID that the AI secret belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"preview_secret": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A masked preview of the secret value returned by Braintrust.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the AI secret was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the AI secret was last updated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *AISecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AISecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AISecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateAISecretRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdAISecret, err := r.client.CreateAISecret(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create AI secret, got error: %s", err),
		)
		return
	}

	fetchedAISecret, err := r.client.GetAISecret(ctx, createdAISecret.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read AI secret after creation, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setAISecretResourceModel(ctx, &data, fetchedAISecret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *AISecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AISecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aiSecret, err := r.client.GetAISecret(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read AI secret, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setAISecretResourceModel(ctx, &data, aiSecret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *AISecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AISecretResourceModel
	var state AISecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateAISecretRequest(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasAISecretUpdateChanges(updateReq) {
		if _, err := r.client.UpdateAISecret(ctx, state.ID.ValueString(), updateReq); err != nil {
			if client.IsNotFound(err) {
				resp.State.RemoveResource(ctx)
				return
			}

			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update AI secret, got error: %s", err),
			)
			return
		}
	}

	updatedAISecret, err := r.client.GetAISecret(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read AI secret after update, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setAISecretResourceModel(ctx, &plan, updatedAISecret)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Secret = resolveAISecretSecretAfterUpdate(plan.Secret, state.Secret)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource.
func (r *AISecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AISecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAISecret(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete AI secret, got error: %s", err),
		)
	}
}

// ImportState implements resource.ResourceWithImportState by importing an AI secret by ID.
// Braintrust does not return the raw secret during import, so users must
// re-supply secret in configuration when they need future rotations.
func (r *AISecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildCreateAISecretRequest(ctx context.Context, data AISecretResourceModel) (*client.CreateAISecretRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	diags.Append(validateAISecretRequiredString(data.Name, "name", "creating")...)
	diags.Append(validateAISecretRequiredString(data.Secret, "secret", "creating")...)
	if diags.HasError() {
		return nil, diags
	}

	metadata, metadataDiags := aiSecretMetadataFromTerraformMap(ctx, data.Metadata)
	diags.Append(metadataDiags...)
	if diags.HasError() {
		return nil, diags
	}

	createReq := &client.CreateAISecretRequest{
		Name:   strings.TrimSpace(data.Name.ValueString()),
		Secret: data.Secret.ValueString(),
	}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		createReq.Type = strings.TrimSpace(data.Type.ValueString())
	}
	if !data.OrgName.IsNull() && !data.OrgName.IsUnknown() {
		createReq.OrgName = strings.TrimSpace(data.OrgName.ValueString())
	}
	if metadata != nil {
		createReq.Metadata = metadata
	}

	return createReq, diags
}

func buildUpdateAISecretRequest(
	ctx context.Context,
	plan AISecretResourceModel,
	state AISecretResourceModel,
) (*client.UpdateAISecretRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	updateReq := &client.UpdateAISecretRequest{}

	if !plan.Name.Equal(state.Name) {
		diags.Append(validateAISecretRequiredString(plan.Name, "name", "updating")...)
		if diags.HasError() {
			return nil, diags
		}
		updateReq.Name = aiSecretStringPtr(strings.TrimSpace(plan.Name.ValueString()))
	}

	if !plan.Type.IsUnknown() && !plan.Type.Equal(state.Type) {
		if !plan.Type.IsNull() {
			updateReq.Type = aiSecretStringPtr(strings.TrimSpace(plan.Type.ValueString()))
		}
	}

	if !plan.Secret.IsUnknown() && !plan.Secret.Equal(state.Secret) {
		if !plan.Secret.IsNull() {
			diags.Append(validateAISecretRequiredString(plan.Secret, "secret", "updating")...)
			if diags.HasError() {
				return nil, diags
			}
			updateReq.Secret = aiSecretStringPtr(plan.Secret.ValueString())
		}
	}

	if !plan.Metadata.IsUnknown() && !plan.Metadata.Equal(state.Metadata) {
		if plan.Metadata.IsNull() {
			updateReq.Metadata = aiSecretMapPtr(map[string]interface{}{})
		} else {
			metadata, metadataDiags := aiSecretMetadataFromTerraformMap(ctx, plan.Metadata)
			diags.Append(metadataDiags...)
			if diags.HasError() {
				return nil, diags
			}
			updateReq.Metadata = aiSecretMapPtr(metadata)
		}
	}

	return updateReq, diags
}

func hasAISecretUpdateChanges(req *client.UpdateAISecretRequest) bool {
	return req.Name != nil ||
		req.Type != nil ||
		req.Secret != nil ||
		req.Metadata != nil
}

func setAISecretResourceModel(ctx context.Context, data *AISecretResourceModel, aiSecret *client.AISecret) diag.Diagnostics {
	var diags diag.Diagnostics
	priorSecret := data.Secret

	data.ID = types.StringValue(aiSecret.ID)
	data.Name = types.StringValue(aiSecret.Name)
	data.Type = stringOrNull(aiSecret.Type)
	data.OrgID = stringOrNull(aiSecret.OrgID)
	data.PreviewSecret = stringOrNull(aiSecret.PreviewSecret)
	data.Created = stringOrNull(aiSecret.Created)
	data.UpdatedAt = stringOrNull(aiSecret.UpdatedAt)

	metadataValue, metadataDiags := aiSecretMetadataToTerraformMap(ctx, aiSecret.Metadata)
	diags.Append(metadataDiags...)
	if diags.HasError() {
		return diags
	}

	data.Metadata = metadataValue

	// API does not return the raw secret on reads for security reasons.
	// Preserve the prior state value when known.
	if !priorSecret.IsNull() && !priorSecret.IsUnknown() {
		data.Secret = priorSecret
	} else {
		data.Secret = types.StringNull()
	}

	return diags
}

func resolveAISecretSecretAfterUpdate(planValue, stateValue types.String) types.String {
	if !planValue.IsNull() && !planValue.IsUnknown() {
		return planValue
	}

	if !stateValue.IsNull() && !stateValue.IsUnknown() {
		return stateValue
	}

	return planValue
}

func aiSecretMetadataFromTerraformMap(ctx context.Context, metadataMap types.Map) (map[string]interface{}, diag.Diagnostics) {
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

func aiSecretStringPtr(value string) *string {
	return &value
}

func aiSecretMapPtr(value map[string]interface{}) *map[string]interface{} {
	return &value
}

func validateAISecretRequiredString(value types.String, fieldName, action string) diag.Diagnostics {
	var diags diag.Diagnostics

	if value.IsUnknown() || value.IsNull() {
		diags.AddAttributeError(
			path.Root(fieldName),
			"Invalid "+fieldName,
			fmt.Sprintf("'%s' must be provided and non-empty when %s an AI secret.", fieldName, action),
		)
		return diags
	}

	if strings.TrimSpace(value.ValueString()) == "" {
		diags.AddAttributeError(
			path.Root(fieldName),
			"Invalid "+fieldName,
			fmt.Sprintf("'%s' must be provided and non-empty when %s an AI secret.", fieldName, action),
		)
	}

	return diags
}
