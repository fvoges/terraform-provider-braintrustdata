package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	errMissingOrganizationID = errors.New("organization ID is required: set resource org_id or provider organization_id")
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OrgResource{}
var _ resource.ResourceWithImportState = &OrgResource{}

// NewOrgResource creates a new organization resource instance.
func NewOrgResource() resource.Resource {
	return &OrgResource{}
}

// OrgResource defines the resource implementation.
type OrgResource struct {
	client *client.Client
}

// OrgResourceModel describes the resource data model.
type OrgResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	OrgID              types.String `tfsdk:"org_id"`
	Name               types.String `tfsdk:"name"`
	APIURL             types.String `tfsdk:"api_url"`
	ProxyURL           types.String `tfsdk:"proxy_url"`
	RealtimeURL        types.String `tfsdk:"realtime_url"`
	ImageRenderingMode types.String `tfsdk:"image_rendering_mode"`
	Created            types.String `tfsdk:"created"`
	IsUniversalAPI     types.Bool   `tfsdk:"is_universal_api"`
	IsDataplanePrivate types.Bool   `tfsdk:"is_dataplane_private"`
}

// Metadata implements resource.Resource.
func (r *OrgResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

// Schema implements resource.Resource.
func (r *OrgResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust organization's settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The organization ID to manage. Defaults to the provider's organization_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The API URL for organization-scoped endpoints.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_universal_api": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the organization uses universal API routing.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_dataplane_private": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether dataplane access is private for this organization.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The proxy URL used by this organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"realtime_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The realtime websocket URL used by this organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_rendering_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Controls image rendering behavior in Braintrust UI. Valid values: `auto`, `click_to_load`, `blocked`.",
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "click_to_load", "blocked"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the organization was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *OrgResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create implements resource.Resource by creating/managing an existing organization settings object.
func (r *OrgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrgResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := resolveOrgID(data.OrgID, r.client.OrgID())
	if err != nil {
		resp.Diagnostics.AddError("Missing Organization ID", err.Error())
		return
	}

	patchReq, hasChanges := buildOrganizationPatchRequestFromPlan(data)
	var org *client.Organization
	if hasChanges {
		org, err = r.client.UpdateOrganization(ctx, orgID, patchReq)
	} else {
		org, err = r.client.GetOrganization(ctx, orgID)
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to initialize organization resource, got error: %s", err))
		return
	}

	populateOrgResourceModel(&data, org, orgID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *OrgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrgResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := ""
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		orgID = data.ID.ValueString()
	}
	if orgID == "" {
		var err error
		orgID, err = resolveOrgID(data.OrgID, r.client.OrgID())
		if err != nil {
			resp.Diagnostics.AddError("Missing Organization ID", err.Error())
			return
		}
	}

	org, err := r.client.GetOrganization(ctx, orgID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
		return
	}

	populateOrgResourceModel(&data, org, orgID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *OrgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrgResourceModel
	var state OrgResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := ""
	if !state.ID.IsNull() && !state.ID.IsUnknown() {
		orgID = state.ID.ValueString()
	}
	if orgID == "" {
		var err error
		orgID, err = resolveOrgID(plan.OrgID, r.client.OrgID())
		if err != nil {
			resp.Diagnostics.AddError("Missing Organization ID", err.Error())
			return
		}
	}

	patchReq, hasChanges := buildOrganizationPatchRequestFromDiff(plan, state)
	var (
		org *client.Organization
		err error
	)
	if hasChanges {
		org, err = r.client.UpdateOrganization(ctx, orgID, patchReq)
	} else {
		org, err = r.client.GetOrganization(ctx, orgID)
	}
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update organization, got error: %s", err))
		return
	}

	populateOrgResourceModel(&plan, org, orgID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource.
func (r *OrgResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Organization not deleted",
		"Braintrust organizations cannot be deleted through the API. The Terraform resource has been removed from state only.",
	)
}

// ImportState implements resource.ResourceWithImportState.
func (r *OrgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func resolveOrgID(planOrgID types.String, clientOrgID string) (string, error) {
	if !planOrgID.IsNull() && !planOrgID.IsUnknown() {
		id := strings.TrimSpace(planOrgID.ValueString())
		if id == "" {
			return "", errMissingOrganizationID
		}
		return id, nil
	}

	id := strings.TrimSpace(clientOrgID)
	if id == "" {
		return "", errMissingOrganizationID
	}

	return id, nil
}

func buildOrganizationPatchRequestFromPlan(plan OrgResourceModel) (*client.PatchOrganizationRequest, bool) {
	req := &client.PatchOrganizationRequest{}
	hasChanges := false

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		v := plan.Name.ValueString()
		req.Name = &v
		hasChanges = true
	}
	if !plan.APIURL.IsNull() && !plan.APIURL.IsUnknown() {
		v := plan.APIURL.ValueString()
		req.APIURL = &v
		hasChanges = true
	}
	if !plan.IsUniversalAPI.IsNull() && !plan.IsUniversalAPI.IsUnknown() {
		v := plan.IsUniversalAPI.ValueBool()
		req.IsUniversalAPI = &v
		hasChanges = true
	}
	if !plan.IsDataplanePrivate.IsNull() && !plan.IsDataplanePrivate.IsUnknown() {
		v := plan.IsDataplanePrivate.ValueBool()
		req.IsDataplanePrivate = &v
		hasChanges = true
	}
	if !plan.ProxyURL.IsNull() && !plan.ProxyURL.IsUnknown() {
		v := plan.ProxyURL.ValueString()
		req.ProxyURL = &v
		hasChanges = true
	}
	if !plan.RealtimeURL.IsNull() && !plan.RealtimeURL.IsUnknown() {
		v := plan.RealtimeURL.ValueString()
		req.RealtimeURL = &v
		hasChanges = true
	}
	if !plan.ImageRenderingMode.IsNull() && !plan.ImageRenderingMode.IsUnknown() {
		v := plan.ImageRenderingMode.ValueString()
		req.ImageRenderingMode = &v
		hasChanges = true
	}

	if !hasChanges {
		return nil, false
	}

	return req, true
}

func buildOrganizationPatchRequestFromDiff(plan, state OrgResourceModel) (*client.PatchOrganizationRequest, bool) {
	req := &client.PatchOrganizationRequest{}
	hasChanges := false

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() && !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		req.Name = &v
		hasChanges = true
	}
	if !plan.APIURL.IsNull() && !plan.APIURL.IsUnknown() && !plan.APIURL.Equal(state.APIURL) {
		v := plan.APIURL.ValueString()
		req.APIURL = &v
		hasChanges = true
	}
	if !plan.IsUniversalAPI.IsNull() && !plan.IsUniversalAPI.IsUnknown() && !plan.IsUniversalAPI.Equal(state.IsUniversalAPI) {
		v := plan.IsUniversalAPI.ValueBool()
		req.IsUniversalAPI = &v
		hasChanges = true
	}
	if !plan.IsDataplanePrivate.IsNull() && !plan.IsDataplanePrivate.IsUnknown() && !plan.IsDataplanePrivate.Equal(state.IsDataplanePrivate) {
		v := plan.IsDataplanePrivate.ValueBool()
		req.IsDataplanePrivate = &v
		hasChanges = true
	}
	if !plan.ProxyURL.IsNull() && !plan.ProxyURL.IsUnknown() && !plan.ProxyURL.Equal(state.ProxyURL) {
		v := plan.ProxyURL.ValueString()
		req.ProxyURL = &v
		hasChanges = true
	}
	if !plan.RealtimeURL.IsNull() && !plan.RealtimeURL.IsUnknown() && !plan.RealtimeURL.Equal(state.RealtimeURL) {
		v := plan.RealtimeURL.ValueString()
		req.RealtimeURL = &v
		hasChanges = true
	}
	if !plan.ImageRenderingMode.IsNull() && !plan.ImageRenderingMode.IsUnknown() && !plan.ImageRenderingMode.Equal(state.ImageRenderingMode) {
		v := plan.ImageRenderingMode.ValueString()
		req.ImageRenderingMode = &v
		hasChanges = true
	}

	if !hasChanges {
		return nil, false
	}

	return req, true
}

func populateOrgResourceModel(model *OrgResourceModel, org *client.Organization, orgID string) {
	id := orgID
	if org != nil && org.ID != "" {
		id = org.ID
	}

	model.ID = types.StringValue(id)
	model.OrgID = types.StringValue(id)

	if org == nil {
		return
	}

	model.Name = types.StringValue(org.Name)
	if org.APIURL != nil {
		model.APIURL = types.StringValue(*org.APIURL)
	} else {
		model.APIURL = types.StringNull()
	}
	if org.IsUniversalAPI != nil {
		model.IsUniversalAPI = types.BoolValue(*org.IsUniversalAPI)
	} else {
		model.IsUniversalAPI = types.BoolNull()
	}
	if org.IsDataplanePrivate != nil {
		model.IsDataplanePrivate = types.BoolValue(*org.IsDataplanePrivate)
	} else {
		model.IsDataplanePrivate = types.BoolNull()
	}
	if org.ProxyURL != nil {
		model.ProxyURL = types.StringValue(*org.ProxyURL)
	} else {
		model.ProxyURL = types.StringNull()
	}
	if org.RealtimeURL != nil {
		model.RealtimeURL = types.StringValue(*org.RealtimeURL)
	} else {
		model.RealtimeURL = types.StringNull()
	}
	if org.ImageRenderingMode != nil {
		model.ImageRenderingMode = types.StringValue(*org.ImageRenderingMode)
	} else {
		model.ImageRenderingMode = types.StringNull()
	}
	if org.Created != nil {
		model.Created = types.StringValue(*org.Created)
	} else {
		model.Created = types.StringNull()
	}
}
