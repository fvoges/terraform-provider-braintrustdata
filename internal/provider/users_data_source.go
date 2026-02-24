package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UsersDataSource{}

// NewUsersDataSource creates a new users data source instance.
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// UsersDataSource defines the data source implementation.
type UsersDataSource struct {
	client *client.Client
}

// UsersDataSourceModel describes the data source data model.
type UsersDataSourceModel struct {
	FilterIDs     types.List            `tfsdk:"filter_ids"`
	GivenName     types.String          `tfsdk:"given_name"`
	FamilyName    types.String          `tfsdk:"family_name"`
	Email         types.String          `tfsdk:"email"`
	OrgName       types.String          `tfsdk:"org_name"`
	StartingAfter types.String          `tfsdk:"starting_after"`
	EndingBefore  types.String          `tfsdk:"ending_before"`
	Users         []UsersDataSourceUser `tfsdk:"users"`
	IDs           []string              `tfsdk:"ids"`
	Limit         types.Int64           `tfsdk:"limit"`
}

// UsersDataSourceUser represents a single user in the list.
type UsersDataSourceUser struct {
	ID         types.String `tfsdk:"id"`
	GivenName  types.String `tfsdk:"given_name"`
	FamilyName types.String `tfsdk:"family_name"`
	Email      types.String `tfsdk:"email"`
	AvatarURL  types.String `tfsdk:"avatar_url"`
	Created    types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Schema implements datasource.DataSource.
func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust users with API-native searchable filters.",
		Attributes: map[string]schema.Attribute{
			"filter_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional list of user IDs to filter by. Maps to repeated `ids` query parameters.",
			},
			"given_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional given name filter.",
			},
			"family_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional family name filter.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional email filter.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of users to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch users after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch users before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned user IDs.",
			},
			"users": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the user.",
						},
						"given_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user's given name.",
						},
						"family_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user's family name.",
						},
						"email": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user's email.",
						},
						"avatar_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The URL of the user's avatar image.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the user was created.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""
	if hasStartingAfter && hasEndingBefore {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot specify both 'starting_after' and 'ending_before'.",
		)
		return
	}

	listOpts := &client.ListUsersOptions{}
	if !data.Limit.IsNull() {
		listOpts.Limit = int(data.Limit.ValueInt64())
	}
	if hasStartingAfter {
		listOpts.StartingAfter = data.StartingAfter.ValueString()
	}
	if hasEndingBefore {
		listOpts.EndingBefore = data.EndingBefore.ValueString()
	}
	if !data.GivenName.IsNull() && data.GivenName.ValueString() != "" {
		listOpts.GivenNames = []string{data.GivenName.ValueString()}
	}
	if !data.FamilyName.IsNull() && data.FamilyName.ValueString() != "" {
		listOpts.FamilyNames = []string{data.FamilyName.ValueString()}
	}
	if !data.Email.IsNull() && data.Email.ValueString() != "" {
		listOpts.Emails = []string{data.Email.ValueString()}
	}
	if !data.OrgName.IsNull() && data.OrgName.ValueString() != "" {
		listOpts.OrgName = data.OrgName.ValueString()
	}
	if !data.FilterIDs.IsNull() {
		var ids []string
		resp.Diagnostics.Append(data.FilterIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		listOpts.IDs = ids
	}

	listResp, err := d.client.ListUsers(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Users",
			fmt.Sprintf("Could not list users: %s", err.Error()),
		)
		return
	}

	data.Users = make([]UsersDataSourceUser, 0, len(listResp.Users))
	data.IDs = make([]string, 0, len(listResp.Users))

	for _, user := range listResp.Users {
		userModel := UsersDataSourceUser{
			ID:         types.StringValue(user.ID),
			GivenName:  types.StringValue(user.GivenName),
			FamilyName: types.StringValue(user.FamilyName),
			Email:      types.StringValue(user.Email),
			AvatarURL:  types.StringValue(user.AvatarURL),
			Created:    types.StringValue(user.Created),
		}

		data.Users = append(data.Users, userModel)
		data.IDs = append(data.IDs, user.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
