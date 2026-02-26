package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

	acl, err := d.client.GetACL(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"ACL Not Found",
				fmt.Sprintf("No ACL found with ID: %s", data.ID.ValueString()),
			)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading ACL",
			fmt.Sprintf("Could not read ACL ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	populateACLDataSourceModel(&data, acl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func populateACLDataSourceModel(data *ACLDataSourceModel, acl *client.ACL) {
	data.ID = types.StringValue(acl.ID)

	if acl.ObjectOrgID != "" {
		data.ObjectOrgID = types.StringValue(acl.ObjectOrgID)
	} else {
		data.ObjectOrgID = types.StringNull()
	}
	if acl.ObjectID != "" {
		data.ObjectID = types.StringValue(acl.ObjectID)
	} else {
		data.ObjectID = types.StringNull()
	}
	if acl.ObjectType != "" {
		data.ObjectType = types.StringValue(string(acl.ObjectType))
	} else {
		data.ObjectType = types.StringNull()
	}
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
	if acl.Created != "" {
		data.Created = types.StringValue(acl.Created)
	} else {
		data.Created = types.StringNull()
	}
}
