package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ACLResource{}
var _ resource.ResourceWithImportState = &ACLResource{}

// NewACLResource creates a new ACL resource instance.
func NewACLResource() resource.Resource {
	return &ACLResource{}
}

// ACLResource defines the resource implementation.
type ACLResource struct {
	client *client.Client
}

// ACLResourceModel describes the resource data model.
type ACLResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ObjectID           types.String `tfsdk:"object_id"`
	ObjectType         types.String `tfsdk:"object_type"`
	UserID             types.String `tfsdk:"user_id"`
	GroupID            types.String `tfsdk:"group_id"`
	RoleID             types.String `tfsdk:"role_id"`
	Permission         types.String `tfsdk:"permission"`
	RestrictObjectType types.String `tfsdk:"restrict_object_type"`
	Created            types.String `tfsdk:"created"`
}

// Metadata implements resource.Resource.
func (r *ACLResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl"
}

// Schema implements resource.Resource.
func (r *ACLResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust ACL (Access Control List) entry. ACLs grant specific permissions to users, groups, or roles on Braintrust objects.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the ACL.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the object to grant access to (e.g., project ID, dataset ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of object. Valid values: organization, project, experiment, dataset, prompt, prompt_session, group, role, org_member, project_log, org_project.",
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
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the user to grant access to. Exactly one of user_id, group_id, or role_id must be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the group to grant access to. Exactly one of user_id, group_id, or role_id must be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the role to grant access to. Exactly one of user_id, group_id, or role_id must be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The permission level to grant. Valid values: create, read, update, delete, create_acls, read_acls, update_acls, delete_acls.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"create",
						"read",
						"update",
						"delete",
						"create_acls",
						"read_acls",
						"update_acls",
						"delete_acls",
					),
				},
			},
			"restrict_object_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "When specified, restricts the ACL to only apply to objects of this type.",
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
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the ACL was created.",
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *ACLResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create implements resource.Resource by creating a new ACL.
func (r *ACLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ACLResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one of user_id, group_id, or role_id is specified
	count := 0
	if !data.UserID.IsNull() {
		count++
	}
	if !data.GroupID.IsNull() {
		count++
	}
	if !data.RoleID.IsNull() {
		count++
	}
	if count != 1 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Exactly one of user_id, group_id, or role_id must be specified.",
		)
		return
	}

	// Create ACL via API
	createReq := &client.CreateACLRequest{
		ObjectID:   data.ObjectID.ValueString(),
		ObjectType: client.ACLObjectType(data.ObjectType.ValueString()),
	}

	if !data.UserID.IsNull() {
		createReq.UserID = data.UserID.ValueString()
	}
	if !data.GroupID.IsNull() {
		createReq.GroupID = data.GroupID.ValueString()
	}
	if !data.RoleID.IsNull() {
		createReq.RoleID = data.RoleID.ValueString()
	}
	if !data.Permission.IsNull() {
		createReq.Permission = client.Permission(data.Permission.ValueString())
	}
	if !data.RestrictObjectType.IsNull() {
		createReq.RestrictObjectType = client.ACLObjectType(data.RestrictObjectType.ValueString())
	}

	acl, err := r.client.CreateACL(ctx, createReq)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ACL, got error: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(acl.ID)
	data.Created = types.StringValue(acl.Created)

	// Set optional fields from response
	if acl.UserID != "" {
		data.UserID = types.StringValue(acl.UserID)
	}
	if acl.GroupID != "" {
		data.GroupID = types.StringValue(acl.GroupID)
	}
	if acl.RoleID != "" {
		data.RoleID = types.StringValue(acl.RoleID)
	}
	if acl.Permission != "" {
		data.Permission = types.StringValue(string(acl.Permission))
	}
	if acl.RestrictObjectType != "" {
		data.RestrictObjectType = types.StringValue(string(acl.RestrictObjectType))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource by reading an ACL.
func (r *ACLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ACLResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get ACL from API
	acl, err := r.client.GetACL(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ACL, got error: %s", err))
		return
	}

	// Update model with response data
	data.ObjectID = types.StringValue(acl.ObjectID)
	data.ObjectType = types.StringValue(string(acl.ObjectType))
	data.Created = types.StringValue(acl.Created)

	// Set optional fields from response
	if acl.UserID != "" {
		data.UserID = types.StringValue(acl.UserID)
	} else {
		data.UserID = types.StringNull()
	}
	if acl.GroupID != "" {
		data.GroupID = types.StringValue(acl.GroupID)
	} else {
		data.GroupID = types.StringNull()
	}
	if acl.RoleID != "" {
		data.RoleID = types.StringValue(acl.RoleID)
	} else {
		data.RoleID = types.StringNull()
	}
	if acl.Permission != "" {
		data.Permission = types.StringValue(string(acl.Permission))
	} else {
		data.Permission = types.StringNull()
	}
	if acl.RestrictObjectType != "" {
		data.RestrictObjectType = types.StringValue(string(acl.RestrictObjectType))
	} else {
		data.RestrictObjectType = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource by updating an ACL.
// Note: ACLs are immutable in the Braintrust API, so this will trigger replacement.
func (r *ACLResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// ACLs are immutable - all changes require replacement
	// This should never be called due to RequiresReplace plan modifiers
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"ACLs are immutable and cannot be updated. All changes require replacement.",
	)
}

// Delete implements resource.Resource by deleting an ACL.
func (r *ACLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ACLResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ACL via API
	err := r.client.DeleteACL(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete ACL, got error: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *ACLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
