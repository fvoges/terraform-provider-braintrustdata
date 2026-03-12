package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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

var _ resource.Resource = &ScoreResource{}
var _ resource.ResourceWithImportState = &ScoreResource{}

// NewScoreResource creates a new score resource instance.
func NewScoreResource() resource.Resource {
	return &ScoreResource{}
}

// ScoreResource defines the resource implementation.
type ScoreResource struct {
	client *client.Client
}

// ScoreResourceModel describes the resource data model.
type ScoreResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ScoreType   types.String `tfsdk:"score_type"`
	Categories  types.String `tfsdk:"categories"`
	Config      types.String `tfsdk:"config"`
	Position    types.String `tfsdk:"position"`
	UserID      types.String `tfsdk:"user_id"`
	Created     types.String `tfsdk:"created"`
}

// Metadata implements resource.Resource.
func (r *ScoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_score"
}

// Schema implements resource.Resource.
func (r *ScoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust project score.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the score.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project ID that owns the score.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The score name.",
			},
			"score_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The score type, such as `categorical`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The score description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"categories": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The score categories as a JSON-encoded string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"config": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The score configuration as a JSON-encoded string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"position": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "LexoRank position of the score within the project.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the score.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the score was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *ScoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ScoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateScoreRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdScore, err := r.client.CreateScore(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create score, got error: %s", err),
		)
		return
	}

	fetchedScore, err := r.client.GetScore(ctx, createdScore.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read score after creation, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setScoreResourceModel(ctx, &data, fetchedScore)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *ScoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	score, err := r.client.GetScore(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read score, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setScoreResourceModel(ctx, &data, score)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *ScoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScoreResourceModel
	var state ScoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateScoreRequest(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasScoreUpdateChanges(updateReq) {
		if _, err := r.client.UpdateScore(ctx, state.ID.ValueString(), updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update score, got error: %s", err),
			)
			return
		}
	}

	score, err := r.client.GetScore(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read score after update, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setScoreResourceModel(ctx, &plan, score)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource.
func (r *ScoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteScore(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete score, got error: %s", err),
		)
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *ScoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildCreateScoreRequest(_ context.Context, model ScoreResourceModel) (*client.CreateScoreRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	categories, fieldDiags := optionalScoreJSONField("categories", model.Categories)
	diags.Append(fieldDiags...)
	config, configDiags := optionalScoreJSONField("config", model.Config)
	diags.Append(configDiags...)
	if diags.HasError() {
		return nil, diags
	}

	req := &client.CreateScoreRequest{
		ProjectID: model.ProjectID.ValueString(),
		Name:      model.Name.ValueString(),
		ScoreType: model.ScoreType.ValueString(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		req.Description = model.Description.ValueString()
	}
	req.Categories = categories
	req.Config = config

	return req, diags
}

func buildUpdateScoreRequest(_ context.Context, plan, state ScoreResourceModel) (*client.UpdateScoreRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.UpdateScoreRequest{}

	if !plan.Name.Equal(state.Name) {
		req.Name = scoreStringPointer(plan.Name.ValueString())
	}

	if !plan.ScoreType.IsUnknown() && !plan.ScoreType.Equal(state.ScoreType) {
		req.ScoreType = scoreStringPointer(plan.ScoreType.ValueString())
	}

	if !plan.Description.IsUnknown() && !plan.Description.Equal(state.Description) {
		req.Description = scoreStringPointer(plan.Description.ValueString())
	}

	categoriesChanged, categories, fieldDiags := scoreJSONFieldChanged("categories", plan.Categories, state.Categories)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if categoriesChanged {
		req.Categories = scoreInterfacePointer(categories)
	}

	configChanged, config, fieldDiags := scoreJSONFieldChanged("config", plan.Config, state.Config)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if configChanged {
		req.Config = scoreInterfacePointer(config)
	}

	return req, diags
}

func scoreJSONFieldChanged(fieldName string, plan, state types.String) (bool, interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if plan.IsUnknown() {
		return false, nil, diags
	}
	if plan.IsNull() {
		return !state.IsNull(), nil, diags
	}

	planDecoded, fieldDiags := optionalScoreJSONField(fieldName, plan)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return false, nil, diags
	}

	if state.IsNull() || state.IsUnknown() {
		return true, planDecoded, diags
	}

	stateDecoded, err := decodeScoreJSONField(fieldName, state)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Invalid %s", fieldName),
			err.Error(),
		)
		return false, nil, diags
	}

	if reflect.DeepEqual(planDecoded, stateDecoded) {
		return false, nil, diags
	}

	return true, planDecoded, diags
}

func decodeScoreJSONField(fieldName string, value types.String) (interface{}, error) {
	if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, fmt.Errorf("%s must be valid JSON and cannot be empty", fieldName)
	}

	var decoded interface{}
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		return nil, fmt.Errorf("%s must be valid JSON: %w", fieldName, err)
	}

	return decoded, nil
}

func optionalScoreJSONField(fieldName string, value types.String) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	decoded, err := decodeScoreJSONField(fieldName, value)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Invalid %s", fieldName),
			err.Error(),
		)
		return nil, diags
	}

	return decoded, diags
}

func hasScoreUpdateChanges(req *client.UpdateScoreRequest) bool {
	return req.Name != nil ||
		req.ScoreType != nil ||
		req.Description != nil ||
		req.Categories != nil ||
		req.Config != nil
}

func setScoreResourceModel(ctx context.Context, model *ScoreResourceModel, score *client.ProjectScore) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = stringOrNull(score.ID)
	model.ProjectID = stringOrNull(score.ProjectID)
	model.Name = stringOrNull(score.Name)
	model.Description = stringOrNull(score.Description)
	model.ScoreType = stringOrNull(score.ScoreType)
	model.UserID = stringOrNull(score.UserID)
	model.Created = stringOrNull(score.Created)

	categories, fieldDiags := jsonEncodedOrNull("categories", score.Categories)
	diags.Append(fieldDiags...)
	model.Categories = categories

	config, configDiags := jsonEncodedOrNull("config", score.Config)
	diags.Append(configDiags...)
	model.Config = config

	if score.Position != nil {
		model.Position = stringOrNull(*score.Position)
	} else {
		model.Position = types.StringNull()
	}

	if diags.HasError() {
		return diags
	}

	_ = ctx
	return diags
}

func scoreStringPointer(v string) *string {
	return &v
}

func scoreInterfacePointer(v interface{}) *interface{} {
	return &v
}
