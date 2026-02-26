package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &APIKeysDataSource{}

// NewAPIKeysDataSource creates a new API keys data source instance.
func NewAPIKeysDataSource() datasource.DataSource {
	return &APIKeysDataSource{}
}

// APIKeysDataSource defines the data source implementation.
type APIKeysDataSource struct {
	client *client.Client
}

// APIKeysDataSourceModel describes the data source data model.
type APIKeysDataSourceModel struct {
	OrgName       types.String              `tfsdk:"org_name"`
	APIKeyName    types.String              `tfsdk:"api_key_name"`
	StartingAfter types.String              `tfsdk:"starting_after"`
	EndingBefore  types.String              `tfsdk:"ending_before"`
	APIKeys       []APIKeysDataSourceAPIKey `tfsdk:"api_keys"`
	IDs           []string                  `tfsdk:"ids"`
	Limit         types.Int64               `tfsdk:"limit"`
}

// APIKeysDataSourceAPIKey represents a single API key in the list.
type APIKeysDataSourceAPIKey struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgID       types.String `tfsdk:"org_id"`
	PreviewName types.String `tfsdk:"preview_name"`
	UserID      types.String `tfsdk:"user_id"`
	UserEmail   types.String `tfsdk:"user_email"`
	Created     types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *APIKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

// Schema implements datasource.DataSource.
func (d *APIKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust API keys using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter.",
			},
			"api_key_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact API key name filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of API keys to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch API keys after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch API keys before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned API key IDs.",
			},
			"api_keys": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of API keys.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the API key.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the API key.",
						},
						"org_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The organization ID that the API key belongs to.",
						},
						"preview_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The preview name of the API key.",
						},
						"user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the user who created the API key.",
						},
						"user_email": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The email of the user who created the API key.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the API key was created.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *APIKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListAPIKeysOptions(data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListAPIKeys(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing API Keys",
			fmt.Sprintf("Could not list API keys: %s", err.Error()),
		)
		return
	}

	data.APIKeys = make([]APIKeysDataSourceAPIKey, 0, len(listResp.APIKeys))
	data.IDs = make([]string, 0, len(listResp.APIKeys))

	for i := range listResp.APIKeys {
		apiKey := &listResp.APIKeys[i]

		apiKeyModel := apiKeysDataSourceAPIKeyFromAPIKey(apiKey)

		data.APIKeys = append(data.APIKeys, apiKeyModel)
		data.IDs = append(data.IDs, apiKey.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListAPIKeysOptions(data APIKeysDataSourceModel) (*client.ListAPIKeysOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	listOpts := &client.ListAPIKeysOptions{}

	if !data.OrgName.IsNull() && data.OrgName.ValueString() != "" {
		listOpts.OrgName = data.OrgName.ValueString()
	}
	if !data.APIKeyName.IsNull() && data.APIKeyName.ValueString() != "" {
		listOpts.APIKeyName = data.APIKeyName.ValueString()
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

func apiKeysDataSourceAPIKeyFromAPIKey(apiKey *client.APIKey) APIKeysDataSourceAPIKey {
	apiKeyModel := APIKeysDataSourceAPIKey{
		ID:   types.StringValue(apiKey.ID),
		Name: types.StringValue(apiKey.Name),
	}

	if apiKey.OrgID != "" {
		apiKeyModel.OrgID = types.StringValue(apiKey.OrgID)
	} else {
		apiKeyModel.OrgID = types.StringNull()
	}
	if apiKey.PreviewName != "" {
		apiKeyModel.PreviewName = types.StringValue(apiKey.PreviewName)
	} else {
		apiKeyModel.PreviewName = types.StringNull()
	}
	if apiKey.Created != "" {
		apiKeyModel.Created = types.StringValue(apiKey.Created)
	} else {
		apiKeyModel.Created = types.StringNull()
	}
	if apiKey.UserID != "" {
		apiKeyModel.UserID = types.StringValue(apiKey.UserID)
	} else {
		apiKeyModel.UserID = types.StringNull()
	}
	if apiKey.UserEmail != "" {
		apiKeyModel.UserEmail = types.StringValue(apiKey.UserEmail)
	} else {
		apiKeyModel.UserEmail = types.StringNull()
	}

	return apiKeyModel
}
