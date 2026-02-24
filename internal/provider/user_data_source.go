package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserDataSource{}

// NewUserDataSource creates a new user data source instance.
func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *client.Client
}

// UserDataSourceModel describes the data source data model.
type UserDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	GivenName  types.String `tfsdk:"given_name"`
	FamilyName types.String `tfsdk:"family_name"`
	Email      types.String `tfsdk:"email"`
	OrgName    types.String `tfsdk:"org_name"`
	AvatarURL  types.String `tfsdk:"avatar_url"`
	Created    types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema implements datasource.DataSource.
func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust user by `id` or by API-native searchable attributes (`email`, `given_name`, `family_name`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the user. Specify either `id` or at least one searchable attribute.",
			},
			"given_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The user's given name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"family_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The user's family name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The user's email. Can be used as a searchable attribute when `id` is not provided.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter applied during searchable lookups.",
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
	}
}

// Configure implements datasource.DataSource.
func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasEmail := !data.Email.IsNull() && data.Email.ValueString() != ""
	hasGivenName := !data.GivenName.IsNull() && data.GivenName.ValueString() != ""
	hasFamilyName := !data.FamilyName.IsNull() && data.FamilyName.ValueString() != ""
	hasOrgName := !data.OrgName.IsNull() && data.OrgName.ValueString() != ""

	searchableCount := 0
	if hasEmail {
		searchableCount++
	}
	if hasGivenName {
		searchableCount++
	}
	if hasFamilyName {
		searchableCount++
	}

	if !hasID && searchableCount == 0 {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or at least one searchable attribute ('email', 'given_name', 'family_name').",
		)
		return
	}

	if hasID && searchableCount > 0 {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('email', 'given_name', 'family_name').",
		)
		return
	}

	var user *client.User
	if hasID {
		fetchedUser, err := d.client.GetUser(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading User",
				fmt.Sprintf("Could not read user ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
		user = fetchedUser
	} else {
		listOpts := &client.ListUsersOptions{Limit: 2}
		if hasEmail {
			listOpts.Emails = []string{data.Email.ValueString()}
		}
		if hasGivenName {
			listOpts.GivenNames = []string{data.GivenName.ValueString()}
		}
		if hasFamilyName {
			listOpts.FamilyNames = []string{data.FamilyName.ValueString()}
		}
		if hasOrgName {
			listOpts.OrgName = data.OrgName.ValueString()
		}

		listResp, err := d.client.ListUsers(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Users",
				fmt.Sprintf("Could not list users using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		if len(listResp.Users) == 0 {
			resp.Diagnostics.AddError(
				"User Not Found",
				"No user found using the provided searchable attributes.",
			)
			return
		}

		if len(listResp.Users) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Users Found",
				"Searchable attributes matched multiple users. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}

		user = &listResp.Users[0]
	}

	data.ID = types.StringValue(user.ID)
	data.GivenName = types.StringValue(user.GivenName)
	data.FamilyName = types.StringValue(user.FamilyName)
	data.Email = types.StringValue(user.Email)
	data.AvatarURL = types.StringValue(user.AvatarURL)
	data.Created = types.StringValue(user.Created)
	if hasOrgName {
		data.OrgName = types.StringValue(data.OrgName.ValueString())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
