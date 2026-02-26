package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &APIKeyDataSource{}

var (
	errAPIKeyNotFoundByName       = errors.New("api key not found by name")
	errMultipleAPIKeysFoundByName = errors.New("multiple api keys found by name")
)

// NewAPIKeyDataSource creates a new API key data source instance.
func NewAPIKeyDataSource() datasource.DataSource {
	return &APIKeyDataSource{}
}

// APIKeyDataSource defines the data source implementation.
type APIKeyDataSource struct {
	client *client.Client
}

// APIKeyDataSourceModel describes the data source data model.
type APIKeyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgName     types.String `tfsdk:"org_name"`
	OrgID       types.String `tfsdk:"org_id"`
	PreviewName types.String `tfsdk:"preview_name"`
	UserID      types.String `tfsdk:"user_id"`
	UserEmail   types.String `tfsdk:"user_email"`
	Created     types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *APIKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// Schema implements datasource.DataSource.
func (d *APIKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust API key by `id` or by API-native searchable attributes (`name`, optionally `org_name`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the API key. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The API key name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter applied during searchable lookups.",
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
	}
}

// Configure implements datasource.DataSource.
func (d *APIKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""
	hasOrgName := !data.OrgName.IsNull() && data.OrgName.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the API key.",
		)
		return
	}

	if hasID && (hasName || hasOrgName) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'org_name').",
		)
		return
	}

	var apiKey *client.APIKey
	if hasID {
		fetchedAPIKey, err := d.client.GetAPIKey(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading API Key",
				fmt.Sprintf("Could not read API key ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
		apiKey = fetchedAPIKey
	} else {
		listOpts := &client.ListAPIKeysOptions{
			APIKeyName: data.Name.ValueString(),
			Limit:      2,
		}
		if hasOrgName {
			listOpts.OrgName = data.OrgName.ValueString()
		}

		listResp, err := d.client.ListAPIKeys(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing API Keys",
				fmt.Sprintf("Could not list API keys using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedAPIKey, err := selectSingleAPIKeyByName(listResp.APIKeys, data.Name.ValueString())
		if errors.Is(err, errAPIKeyNotFoundByName) {
			resp.Diagnostics.AddError(
				"API Key Not Found",
				fmt.Sprintf("No API key found with name: %s", data.Name.ValueString()),
			)
			return
		}
		if errors.Is(err, errMultipleAPIKeysFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple API Keys Found",
				"Searchable attributes matched multiple API keys. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing API Keys",
				fmt.Sprintf("Could not resolve API key using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		apiKey = selectedAPIKey
	}

	populateAPIKeyDataSourceModel(&data, apiKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func selectSingleAPIKeyByName(apiKeys []client.APIKey, apiKeyName string) (*client.APIKey, error) {
	var selected *client.APIKey

	for i := range apiKeys {
		apiKey := &apiKeys[i]
		if apiKey.Name != apiKeyName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleAPIKeysFoundByName, apiKeyName)
		}
		selected = apiKey
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errAPIKeyNotFoundByName, apiKeyName)
	}

	return selected, nil
}

func populateAPIKeyDataSourceModel(data *APIKeyDataSourceModel, apiKey *client.APIKey) {
	data.ID = types.StringValue(apiKey.ID)
	data.Name = types.StringValue(apiKey.Name)

	if apiKey.OrgID != "" {
		data.OrgID = types.StringValue(apiKey.OrgID)
	} else {
		data.OrgID = types.StringNull()
	}
	if apiKey.PreviewName != "" {
		data.PreviewName = types.StringValue(apiKey.PreviewName)
	} else {
		data.PreviewName = types.StringNull()
	}
	if apiKey.Created != "" {
		data.Created = types.StringValue(apiKey.Created)
	} else {
		data.Created = types.StringNull()
	}
	if apiKey.UserID != "" {
		data.UserID = types.StringValue(apiKey.UserID)
	} else {
		data.UserID = types.StringNull()
	}
	if apiKey.UserEmail != "" {
		data.UserEmail = types.StringValue(apiKey.UserEmail)
	} else {
		data.UserEmail = types.StringNull()
	}
}
