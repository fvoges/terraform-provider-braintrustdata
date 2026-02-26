package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ACLDataSource{}

// NewACLDataSource creates a new ACL data source instance.
func NewACLDataSource() datasource.DataSource {
	return &ACLDataSource{}
}

// ACLDataSource defines the data source implementation.
type ACLDataSource struct {
	client *client.Client
}

// ACLDataSourceModel describes the data source data model.
type ACLDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ObjectOrgID        types.String `tfsdk:"object_org_id"`
	ObjectID           types.String `tfsdk:"object_id"`
	ObjectType         types.String `tfsdk:"object_type"`
	UserID             types.String `tfsdk:"user_id"`
	GroupID            types.String `tfsdk:"group_id"`
	RoleID             types.String `tfsdk:"role_id"`
	Permission         types.String `tfsdk:"permission"`
	RestrictObjectType types.String `tfsdk:"restrict_object_type"`
	Created            types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *ACLDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl"
}

// Schema implements datasource.DataSource.
func (d *ACLDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust ACL by `id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the ACL.",
			},
			"object_org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID of the ACL object.",
			},
			"object_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the object the ACL applies to.",
			},
			"object_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of object the ACL applies to.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user ID subject for this ACL, when set.",
			},
			"group_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The group ID subject for this ACL, when set.",
			},
			"role_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The role ID subject for this ACL, when set.",
			},
			"permission": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The permission granted by this ACL.",
			},
			"restrict_object_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Optional object type restriction for this ACL.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the ACL was created.",
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *ACLDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ACLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ACLDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aclID, idDiags := validateACLID(data.ID)
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	acl, err := d.client.GetACL(ctx, aclID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"ACL Not Found",
				fmt.Sprintf("No ACL found with ID: %s", aclID),
			)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading ACL",
			fmt.Sprintf("Could not read ACL ID %s: %s", aclID, err.Error()),
		)
		return
	}

	populateACLDataSourceModel(&data, acl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func populateACLDataSourceModel(data *ACLDataSourceModel, acl *client.ACL) {
	data.ID = types.StringValue(acl.ID)

	fields := aclDataSourceFieldsFromACL(acl)
	data.ObjectOrgID = fields.ObjectOrgID
	data.ObjectID = fields.ObjectID
	data.ObjectType = fields.ObjectType
	data.UserID = fields.UserID
	data.GroupID = fields.GroupID
	data.RoleID = fields.RoleID
	data.Permission = fields.Permission
	data.RestrictObjectType = fields.RestrictObjectType
	data.Created = fields.Created
}

type aclDataSourceFields struct {
	ObjectOrgID        types.String
	ObjectID           types.String
	ObjectType         types.String
	UserID             types.String
	GroupID            types.String
	RoleID             types.String
	Permission         types.String
	RestrictObjectType types.String
	Created            types.String
}

func aclDataSourceFieldsFromACL(acl *client.ACL) aclDataSourceFields {
	return aclDataSourceFields{
		ObjectOrgID:        stringOrNull(acl.ObjectOrgID),
		ObjectID:           stringOrNull(acl.ObjectID),
		ObjectType:         stringOrNull(string(acl.ObjectType)),
		UserID:             stringOrNull(acl.UserID),
		GroupID:            stringOrNull(acl.GroupID),
		RoleID:             stringOrNull(acl.RoleID),
		Permission:         stringOrNull(string(acl.Permission)),
		RestrictObjectType: stringOrNull(string(acl.RestrictObjectType)),
		Created:            stringOrNull(acl.Created),
	}
}

func stringOrNull(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}

func validateACLID(id types.String) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if id.IsUnknown() {
		diags.AddError(
			"Invalid ACL ID",
			"ACL ID is unknown. Ensure the ID is set before reading this data source.",
		)
		return "", diags
	}

	if id.IsNull() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError(
			"Invalid ACL ID",
			"ACL ID must be provided and non-empty when reading this ACL data source.",
		)
		return "", diags
	}

	return id.ValueString(), diags
}
