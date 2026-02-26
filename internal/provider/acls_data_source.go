package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ACLsDataSource{}

// NewACLsDataSource creates a new ACLs data source instance.
func NewACLsDataSource() datasource.DataSource {
	return &ACLsDataSource{}
}

// ACLsDataSource defines the data source implementation.
type ACLsDataSource struct {
	client *client.Client
}

// ACLsDataSourceModel describes the data source data model.
type ACLsDataSourceModel struct {
	ObjectID      types.String        `tfsdk:"object_id"`
	ObjectType    types.String        `tfsdk:"object_type"`
	StartingAfter types.String        `tfsdk:"starting_after"`
	EndingBefore  types.String        `tfsdk:"ending_before"`
	ACLs          []ACLsDataSourceACL `tfsdk:"acls"`
	IDs           []string            `tfsdk:"ids"`
	Limit         types.Int64         `tfsdk:"limit"`
}

// ACLsDataSourceACL represents a single ACL in the list.
type ACLsDataSourceACL struct {
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
func (d *ACLsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acls"
}

// Schema implements datasource.DataSource.
func (d *ACLsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust ACLs for an object using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"object_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object ID to list ACLs for.",
			},
			"object_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object type to list ACLs for.",
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
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of ACLs to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch ACLs after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch ACLs before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned ACL IDs.",
			},
			"acls": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of ACLs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
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
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *ACLsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ACLsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ACLsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListACLsOptions(data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListACLs(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing ACLs",
			fmt.Sprintf("Could not list ACLs: %s", err.Error()),
		)
		return
	}

	data.ACLs = make([]ACLsDataSourceACL, 0, len(listResp.Objects))
	data.IDs = make([]string, 0, len(listResp.Objects))

	for i := range listResp.Objects {
		acl := &listResp.Objects[i]
		aclModel := aclsDataSourceACLFromACL(acl)

		data.ACLs = append(data.ACLs, aclModel)
		data.IDs = append(data.IDs, acl.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListACLsOptions(data ACLsDataSourceModel) (*client.ListACLsOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	objectID := strings.TrimSpace(data.ObjectID.ValueString())
	if objectID == "" {
		diags.AddError("Invalid object_id", "'object_id' must be provided and non-empty.")
		return nil, diags
	}

	objectTypeValue := strings.TrimSpace(data.ObjectType.ValueString())
	if objectTypeValue == "" {
		diags.AddError("Invalid object_type", "'object_type' must be provided and non-empty.")
		return nil, diags
	}

	listOpts := &client.ListACLsOptions{
		ObjectID:   objectID,
		ObjectType: client.ACLObjectType(objectTypeValue),
	}

	if !data.Limit.IsNull() {
		limit := data.Limit.ValueInt64()
		if limit < 1 {
			diags.AddError("Invalid Limit", "'limit' must be greater than or equal to 1.")
			return nil, diags
		}

		maxInt := int64(^uint(0) >> 1)
		if limit > maxInt {
			diags.AddError("Invalid Limit", "'limit' exceeds supported platform integer size.")
			return nil, diags
		}

		listOpts.Limit = int(limit)
	}
	if hasStartingAfter {
		listOpts.StartingAfter = data.StartingAfter.ValueString()
	}
	if hasEndingBefore {
		listOpts.EndingBefore = data.EndingBefore.ValueString()
	}

	return listOpts, diags
}

func aclsDataSourceACLFromACL(acl *client.ACL) ACLsDataSourceACL {
	aclModel := ACLsDataSourceACL{
		ID: types.StringValue(acl.ID),
	}

	fields := aclDataSourceFieldsFromACL(acl)
	aclModel.ObjectOrgID = fields.ObjectOrgID
	aclModel.ObjectID = fields.ObjectID
	aclModel.ObjectType = fields.ObjectType
	aclModel.UserID = fields.UserID
	aclModel.GroupID = fields.GroupID
	aclModel.RoleID = fields.RoleID
	aclModel.Permission = fields.Permission
	aclModel.RestrictObjectType = fields.RestrictObjectType
	aclModel.Created = fields.Created

	return aclModel
}
