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

var _ datasource.DataSource = &OrgsDataSource{}

// NewOrgsDataSource creates a new organizations data source instance.
func NewOrgsDataSource() datasource.DataSource {
	return &OrgsDataSource{}
}

// OrgsDataSource defines the data source implementation.
type OrgsDataSource struct {
	client *client.Client
}

// OrgsDataSourceModel describes the data source data model.
type OrgsDataSourceModel struct {
	OrgName       types.String        `tfsdk:"org_name"`
	StartingAfter types.String        `tfsdk:"starting_after"`
	EndingBefore  types.String        `tfsdk:"ending_before"`
	Orgs          []OrgsDataSourceOrg `tfsdk:"orgs"`
	IDs           []string            `tfsdk:"ids"`
	Limit         types.Int64         `tfsdk:"limit"`
}

// OrgsDataSourceOrg represents a single organization in the list.
type OrgsDataSourceOrg struct {
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
func (d *OrgsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_orgs"
}

// Schema implements datasource.DataSource.
func (d *OrgsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust organizations using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact organization name filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of organizations to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch organizations after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch organizations before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned organization IDs.",
			},
			"orgs": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of organizations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the organization.",
						},
						"org_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The organization ID.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The organization name.",
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
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *OrgsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListOrganizationsOptions(data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListOrganizations(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Organizations",
			fmt.Sprintf("Could not list organizations: %s", err.Error()),
		)
		return
	}

	data.Orgs = make([]OrgsDataSourceOrg, 0, len(listResp.Organizations))
	data.IDs = make([]string, 0, len(listResp.Organizations))

	for i := range listResp.Organizations {
		org := &listResp.Organizations[i]
		data.Orgs = append(data.Orgs, orgsDataSourceOrgFromOrganization(org))
		data.IDs = append(data.IDs, org.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListOrganizationsOptions(data OrgsDataSourceModel) (*client.ListOrganizationsOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""
	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	listOpts := &client.ListOrganizationsOptions{}
	if !data.OrgName.IsNull() && data.OrgName.ValueString() != "" {
		listOpts.OrgName = data.OrgName.ValueString()
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

func orgsDataSourceOrgFromOrganization(org *client.Organization) OrgsDataSourceOrg {
	model := OrgsDataSourceOrg{
		Name: types.StringValue(org.Name),
	}

	if org.ID != "" {
		model.ID = types.StringValue(org.ID)
		model.OrgID = types.StringValue(org.ID)
	} else {
		model.ID = types.StringNull()
		model.OrgID = types.StringNull()
	}
	if org.APIURL != nil {
		model.APIURL = types.StringValue(*org.APIURL)
	} else {
		model.APIURL = types.StringNull()
	}
	if org.IsUniversalAPI != nil {
		model.IsUniversalAPI = types.BoolValue(*org.IsUniversalAPI)
	} else {
		model.IsUniversalAPI = types.BoolNull()
	}
	if org.IsDataplanePrivate != nil {
		model.IsDataplanePrivate = types.BoolValue(*org.IsDataplanePrivate)
	} else {
		model.IsDataplanePrivate = types.BoolNull()
	}
	if org.ProxyURL != nil {
		model.ProxyURL = types.StringValue(*org.ProxyURL)
	} else {
		model.ProxyURL = types.StringNull()
	}
	if org.RealtimeURL != nil {
		model.RealtimeURL = types.StringValue(*org.RealtimeURL)
	} else {
		model.RealtimeURL = types.StringNull()
	}
	if org.ImageRenderingMode != nil {
		model.ImageRenderingMode = types.StringValue(*org.ImageRenderingMode)
	} else {
		model.ImageRenderingMode = types.StringNull()
	}
	if org.Created != nil {
		model.Created = types.StringValue(*org.Created)
	} else {
		model.Created = types.StringNull()
	}

	return model
}
