package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ViewResource{}
var _ resource.ResourceWithImportState = &ViewResource{}

// NewViewResource creates a new view resource instance.
func NewViewResource() resource.Resource {
	return &ViewResource{}
}

// ViewResource defines the resource implementation.
type ViewResource struct {
	client *client.Client
}

// ViewResourceModel describes the resource data model.
type ViewResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ObjectID   types.String `tfsdk:"object_id"`
	ObjectType types.String `tfsdk:"object_type"`
	ViewType   types.String `tfsdk:"view_type"`
	Name       types.String `tfsdk:"name"`
	Options    types.String `tfsdk:"options"`
	ViewData   types.String `tfsdk:"view_data"`
	UserID     types.String `tfsdk:"user_id"`
	Created    types.String `tfsdk:"created"`
	DeletedAt  types.String `tfsdk:"deleted_at"`
}

// Metadata implements resource.Resource.
func (r *ViewResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_view"
}

// Schema implements resource.Resource.
func (r *ViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust view.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the view.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object ID that owns the view.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object type that owns the view.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"organization",
						"project",
						"experiment",
						"dataset",
						"prompt",
						"prompt_session",
						"group",
						"role",
						"org_member",
						"project_log",
						"org_project",
					),
				},
			},
			"view_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The view type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"projects",
						"experiments",
						"experiment",
						"playgrounds",
						"playground",
						"datasets",
						"dataset",
						"prompts",
						"tools",
						"scorers",
						"logs",
					),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The view name.",
			},
			"options": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional view options as a JSON-encoded object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"view_data": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional view definition as a JSON-encoded object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the view.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the view was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deleted_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The deletion timestamp, if the view has been soft-deleted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *ViewResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateViewRequest(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdView, err := r.client.CreateView(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create view, got error: %s", err),
		)
		return
	}

	view, err := r.client.GetView(ctx, createdView.ID, &client.GetViewOptions{
		ObjectID:   data.ObjectID.ValueString(),
		ObjectType: client.ACLObjectType(data.ObjectType.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read view after creation, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setViewResourceModel(ctx, &data, view)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *ViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view, err := r.client.GetView(ctx, data.ID.ValueString(), &client.GetViewOptions{
		ObjectID:   data.ObjectID.ValueString(),
		ObjectType: client.ACLObjectType(data.ObjectType.ValueString()),
	})
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read view, got error: %s", err),
		)
		return
	}

	if view.DeletedAt != "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(setViewResourceModel(ctx, &data, view)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *ViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ViewResourceModel
	var state ViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateViewRequest(plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if hasViewUpdateChanges(updateReq) {
		if _, err := r.client.UpdateView(ctx, state.ID.ValueString(), updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update view, got error: %s", err),
			)
			return
		}
	}

	view, err := r.client.GetView(ctx, state.ID.ValueString(), &client.GetViewOptions{
		ObjectID:   state.ObjectID.ValueString(),
		ObjectType: client.ACLObjectType(state.ObjectType.ValueString()),
	})
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read view after update, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(setViewResourceModel(ctx, &plan, view)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource.
func (r *ViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteView(ctx, data.ID.ValueString(), &client.DeleteViewRequest{
		ObjectID:   data.ObjectID.ValueString(),
		ObjectType: client.ACLObjectType(data.ObjectType.ValueString()),
	})
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete view, got error: %s", err),
		)
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *ViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	viewID, objectID, objectType, err := parseViewImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), viewID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("object_id"), objectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("object_type"), objectType)...)
}

func buildCreateViewRequest(model ViewResourceModel) (*client.CreateViewRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	options, fieldDiags := optionalViewJSONObjectField("options", model.Options)
	diags.Append(fieldDiags...)
	viewData, fieldDiags := optionalViewJSONObjectField("view_data", model.ViewData)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return nil, diags
	}

	req := &client.CreateViewRequest{
		ObjectID:   model.ObjectID.ValueString(),
		ObjectType: client.ACLObjectType(model.ObjectType.ValueString()),
		ViewType:   client.ViewType(model.ViewType.ValueString()),
		Name:       model.Name.ValueString(),
		Options:    options,
		ViewData:   viewData,
	}

	return req, diags
}

func buildUpdateViewRequest(plan, state ViewResourceModel) (*client.UpdateViewRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := &client.UpdateViewRequest{
		ObjectID:   plan.ObjectID.ValueString(),
		ObjectType: client.ACLObjectType(plan.ObjectType.ValueString()),
	}

	if !plan.Name.IsUnknown() && !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		req.Name = &name
	}

	optionsChanged, options, fieldDiags := viewJSONObjectFieldChanged("options", plan.Options, state.Options)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if optionsChanged {
		req.Options = options
	}

	viewDataChanged, viewData, fieldDiags := viewJSONObjectFieldChanged("view_data", plan.ViewData, state.ViewData)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if viewDataChanged {
		req.ViewData = viewData
	}

	return req, diags
}

func hasViewUpdateChanges(req *client.UpdateViewRequest) bool {
	return req.Name != nil || req.Options != nil || req.ViewData != nil
}

func viewJSONRawMessage(body []byte) *json.RawMessage {
	msg := json.RawMessage(body)
	return &msg
}

func viewJSONNull() *json.RawMessage {
	return viewJSONRawMessage([]byte("null"))
}

func decodeViewJSONObjectField(fieldName string, value types.String) (map[string]interface{}, error) {
	if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, fmt.Errorf("%s must be valid JSON and cannot be empty", fieldName)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		return nil, fmt.Errorf("%s must be valid JSON: %w", fieldName, err)
	}

	return decoded, nil
}

func optionalViewJSONObjectField(fieldName string, value types.String) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	decoded, err := decodeViewJSONObjectField(fieldName, value)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Invalid %s", fieldName),
			err.Error(),
		)
		return nil, diags
	}

	return decoded, diags
}

func viewJSONObjectFieldChanged(fieldName string, plan, state types.String) (bool, *json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics

	if plan.IsUnknown() {
		return false, nil, diags
	}
	if plan.IsNull() {
		if state.IsNull() || state.IsUnknown() {
			return false, nil, diags
		}

		return true, viewJSONNull(), diags
	}

	planDecoded, fieldDiags := optionalViewJSONObjectField(fieldName, plan)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return false, nil, diags
	}
	planEncoded, err := json.Marshal(planDecoded)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Invalid %s", fieldName),
			fmt.Sprintf("could not encode %s to JSON: %s", fieldName, err),
		)
		return false, nil, diags
	}

	if state.IsNull() || state.IsUnknown() {
		return true, viewJSONRawMessage(planEncoded), diags
	}

	stateDecoded, err := decodeViewJSONObjectField(fieldName, state)
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

	return true, viewJSONRawMessage(planEncoded), diags
}

func setViewResourceModel(ctx context.Context, model *ViewResourceModel, view *client.View) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = stringOrNull(view.ID)
	model.ObjectID = stringOrNull(view.ObjectID)
	model.ObjectType = stringOrNull(string(view.ObjectType))
	model.ViewType = stringOrNull(string(view.ViewType))
	model.Name = stringOrNull(view.Name)
	model.UserID = stringOrNull(view.UserID)
	model.Created = stringOrNull(view.Created)
	model.DeletedAt = stringOrNull(view.DeletedAt)

	options, fieldDiags := viewJSONValueOrPreserve("options", model.Options, view.Options)
	diags.Append(fieldDiags...)
	model.Options = options

	viewData, fieldDiags := viewJSONValueOrPreserve("view_data", model.ViewData, view.ViewData)
	diags.Append(fieldDiags...)
	model.ViewData = viewData

	_ = ctx
	return diags
}

func viewJSONValueOrPreserve(fieldName string, current types.String, apiValue interface{}) (types.String, diag.Diagnostics) {
	encoded, diags := jsonEncodedOrNull(fieldName, apiValue)
	if diags.HasError() || encoded.IsNull() || current.IsNull() || current.IsUnknown() {
		return encoded, diags
	}

	currentDecoded, err := decodeViewJSONObjectField(fieldName, current)
	if err != nil {
		return encoded, diags
	}

	apiDecoded, err := decodeViewJSONObjectField(fieldName, encoded)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Invalid %s", fieldName),
			err.Error(),
		)
		return types.StringNull(), diags
	}

	if reflect.DeepEqual(currentDecoded, apiDecoded) {
		return current, diags
	}

	return encoded, diags
}

func parseViewImportID(raw string) (string, string, string, error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("expected import ID in the format <view_id>,<object_id>,<object_type>")
	}

	viewID := strings.TrimSpace(parts[0])
	objectID := strings.TrimSpace(parts[1])
	objectType := strings.TrimSpace(parts[2])
	if viewID == "" || objectID == "" || objectType == "" {
		return "", "", "", fmt.Errorf("expected import ID in the format <view_id>,<object_id>,<object_type>")
	}

	return viewID, objectID, objectType, nil
}
