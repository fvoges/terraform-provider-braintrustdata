package provider

import (
	"context"
	"os"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ provider.Provider = &BraintrustProvider{}

// BraintrustProvider defines the provider implementation.
type BraintrustProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BraintrustProviderModel describes the provider data model.
type BraintrustProviderModel struct {
	APIKey         types.String `tfsdk:"api_key"`
	APIURL         types.String `tfsdk:"api_url"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

// Metadata returns the provider type name.
func (p *BraintrustProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "braintrustdata"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *BraintrustProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Braintrust (braintrust.dev), enabling infrastructure-as-code " +
			"management of projects, experiments, datasets, prompts, functions, and access control.",
		MarkdownDescription: "Terraform provider for [Braintrust](https://braintrust.dev), enabling infrastructure-as-code " +
			"management of projects, experiments, datasets, prompts, functions, and access control.\n\n" +
			"## Authentication\n\n" +
			"The provider requires a Braintrust API key (format: `sk-*`). " +
			"You can obtain an API key from your Braintrust organization settings.\n\n" +
			"## Example Usage\n\n" +
			"```terraform\n" +
			"provider \"braintrustdata\" {\n" +
			"  api_key         = \"sk-***\"\n" +
			"  organization_id = \"org-***\"\n" +
			"}\n" +
			"```\n\n" +
			"Alternatively, use environment variables:\n\n" +
			"```sh\n" +
			"export BRAINTRUST_API_KEY=\"sk-***\"\n" +
			"export BRAINTRUST_ORG_ID=\"org-***\"\n" +
			"```",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description:         "Braintrust API key (format: sk-*). Can also be set via BRAINTRUST_API_KEY environment variable.",
				MarkdownDescription: "Braintrust API key (format: `sk-*`). Can also be set via `BRAINTRUST_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				Description:         "Braintrust API base URL. Defaults to https://api.braintrust.dev. Can also be set via BRAINTRUST_API_URL environment variable.",
				MarkdownDescription: "Braintrust API base URL. Defaults to `https://api.braintrust.dev`. Can also be set via `BRAINTRUST_API_URL` environment variable.",
				Optional:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Default Braintrust organization ID. Can also be set via BRAINTRUST_ORG_ID environment variable.",
				MarkdownDescription: "Default Braintrust organization ID. Can also be set via `BRAINTRUST_ORG_ID` environment variable.",
				Optional:            true,
			},
		},
	}
}

// Configure prepares a Braintrust API client for data sources and resources.
func (p *BraintrustProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config BraintrustProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get API key from config or environment
	apiKey := os.Getenv("BRAINTRUST_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// Get API URL from config or environment, default to production
	apiURL := "https://api.braintrust.dev"
	if envURL := os.Getenv("BRAINTRUST_API_URL"); envURL != "" {
		apiURL = envURL
	}
	if !config.APIURL.IsNull() {
		apiURL = config.APIURL.ValueString()
	}

	// Get organization ID from config or environment
	orgID := os.Getenv("BRAINTRUST_ORG_ID")
	if !config.OrganizationID.IsNull() {
		orgID = config.OrganizationID.ValueString()
	}

	// Validate required fields
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"The provider requires a Braintrust API key. "+
				"Set the api_key attribute in the provider configuration or the BRAINTRUST_API_KEY environment variable. "+
				"You can obtain an API key from your Braintrust organization settings at https://www.braintrust.dev.",
		)
		return
	}

	// Create API client
	c := client.NewClient(apiKey, apiURL, orgID)

	// Make the client available to data sources and resources
	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources defines the resources implemented in the provider.
func (p *BraintrustProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewACLResource,
		NewGroupResource,
		NewProjectResource,
		NewRoleResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *BraintrustProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGroupDataSource,
		NewGroupsDataSource,
	}
}

// New returns a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BraintrustProvider{
			version: version,
		}
	}
}
