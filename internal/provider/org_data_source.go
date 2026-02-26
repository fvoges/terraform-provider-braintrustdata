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

var _ datasource.DataSource = &OrgDataSource{}

var (
	errOrganizationNotFoundByName       = errors.New("organization not found by name")
	errMultipleOrganizationsFoundByName = errors.New("multiple organizations found by name")
)

// NewOrgDataSource creates a new organization data source instance.
func NewOrgDataSource() datasource.DataSource {
	return &OrgDataSource{}
}

// OrgDataSource defines the data source implementation.
type OrgDataSource struct {
	client *client.Client
}

// OrgDataSourceModel describes the data source data model.
type OrgDataSourceModel struct {
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

// Metadata implements datasource.DataSource.
func (d *OrgDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

// Schema implements datasource.DataSource.
func (d *OrgDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust organization by `id` or by exact `name`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the organization. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The organization name. Used as an exact-match lookup when `id` is not provided.",
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID.",
			},
			"api_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The API URL for organization-scoped endpoints.",
			},
			"is_universal_api": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the organization uses universal API routing.",
			},
			"is_dataplane_private": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether dataplane access is private for this organization.",
			},
			"proxy_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The proxy URL used by this organization.",
			},
			"realtime_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The realtime websocket URL used by this organization.",
			},
			"image_rendering_mode": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Image rendering behavior in Braintrust UI.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the organization was created.",
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *OrgDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the organization.",
		)
		return
	}

	if hasID && hasName {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with 'name'.",
		)
		return
	}

	var org *client.Organization
	if hasID {
		fetchedOrg, err := d.client.GetOrganization(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Organization",
				fmt.Sprintf("Could not read organization ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
		org = fetchedOrg
	} else {
		listResp, err := d.client.ListOrganizations(ctx, &client.ListOrganizationsOptions{
			OrgName: data.Name.ValueString(),
			Limit:   2,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Organizations",
				fmt.Sprintf("Could not list organizations by name: %s", err.Error()),
			)
			return
		}

		selectedOrg, err := selectSingleOrganizationByName(listResp.Organizations, data.Name.ValueString())
		if errors.Is(err, errOrganizationNotFoundByName) {
			resp.Diagnostics.AddError(
				"Organization Not Found",
				fmt.Sprintf("No organization found with name: %s", data.Name.ValueString()),
			)
			return
		}
		if errors.Is(err, errMultipleOrganizationsFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple Organizations Found",
				"Searchable attributes matched multiple organizations. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Organizations",
				fmt.Sprintf("Could not resolve organization by name: %s", err.Error()),
			)
			return
		}

		org = selectedOrg
	}

	populateOrgDataSourceModel(&data, org)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func selectSingleOrganizationByName(orgs []client.Organization, orgName string) (*client.Organization, error) {
	var selected *client.Organization

	for i := range orgs {
		org := &orgs[i]
		if org.Name != orgName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleOrganizationsFoundByName, orgName)
		}
		selected = org
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errOrganizationNotFoundByName, orgName)
	}

	return selected, nil
}

func populateOrgDataSourceModel(data *OrgDataSourceModel, org *client.Organization) {
	if org.ID != "" {
		data.ID = types.StringValue(org.ID)
		data.OrgID = types.StringValue(org.ID)
	} else {
		data.ID = types.StringNull()
		data.OrgID = types.StringNull()
	}

	data.Name = types.StringValue(org.Name)

	if org.APIURL != nil {
		data.APIURL = types.StringValue(*org.APIURL)
	} else {
		data.APIURL = types.StringNull()
	}
	if org.IsUniversalAPI != nil {
		data.IsUniversalAPI = types.BoolValue(*org.IsUniversalAPI)
	} else {
		data.IsUniversalAPI = types.BoolNull()
	}
	if org.IsDataplanePrivate != nil {
		data.IsDataplanePrivate = types.BoolValue(*org.IsDataplanePrivate)
	} else {
		data.IsDataplanePrivate = types.BoolNull()
	}
	if org.ProxyURL != nil {
		data.ProxyURL = types.StringValue(*org.ProxyURL)
	} else {
		data.ProxyURL = types.StringNull()
	}
	if org.RealtimeURL != nil {
		data.RealtimeURL = types.StringValue(*org.RealtimeURL)
	} else {
		data.RealtimeURL = types.StringNull()
	}
	if org.ImageRenderingMode != nil {
		data.ImageRenderingMode = types.StringValue(*org.ImageRenderingMode)
	} else {
		data.ImageRenderingMode = types.StringNull()
	}
	if org.Created != nil {
		data.Created = types.StringValue(*org.Created)
	} else {
		data.Created = types.StringNull()
	}
}
