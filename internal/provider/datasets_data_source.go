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

var _ datasource.DataSource = &DatasetsDataSource{}

// NewDatasetsDataSource creates a new datasets data source instance.
func NewDatasetsDataSource() datasource.DataSource {
	return &DatasetsDataSource{}
}

// DatasetsDataSource defines the data source implementation.
type DatasetsDataSource struct {
	client *client.Client
}

// DatasetsDataSourceModel describes the data source data model.
type DatasetsDataSourceModel struct {
	ProjectID types.String                `tfsdk:"project_id"`
	Name      types.String                `tfsdk:"name"`
	Datasets  []DatasetsDataSourceDataset `tfsdk:"datasets"`
	IDs       []string                    `tfsdk:"ids"`
}

// DatasetsDataSourceDataset represents a single dataset in the list.
type DatasetsDataSourceDataset struct {
	Tags        types.Set    `tfsdk:"tags"`
	Metadata    types.Map    `tfsdk:"metadata"`
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	UserID      types.String `tfsdk:"user_id"`
	OrgID       types.String `tfsdk:"org_id"`
}

// Metadata implements datasource.DataSource.
func (d *DatasetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasets"
}

// Schema implements datasource.DataSource.
func (d *DatasetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Braintrust datasets in a project. Optionally filter by dataset name.",

		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project ID to filter datasets.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional name filter to return only datasets with this exact name.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of dataset IDs.",
			},
			"datasets": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of datasets.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the dataset.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the dataset.",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project ID the dataset belongs to.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of the dataset.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the dataset was created.",
						},
						"user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the user who created the dataset.",
						},
						"org_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the organization this dataset belongs to.",
						},
						"metadata": schema.MapAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Metadata associated with the dataset as key-value pairs.",
						},
						"tags": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Tags associated with the dataset.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *DatasetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatasetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatasetsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"project_id is required to list datasets.",
		)
		return
	}

	nameFilter := ""
	if !data.Name.IsNull() && data.Name.ValueString() != "" {
		nameFilter = data.Name.ValueString()
	}

	data.Datasets = make([]DatasetsDataSourceDataset, 0)
	data.IDs = make([]string, 0)

	// Paginate through all datasets
	cursor := ""
	for {
		listResp, err := d.client.ListDatasets(ctx, &client.ListDatasetsOptions{
			ProjectID: projectID,
			Cursor:    cursor,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Datasets",
				fmt.Sprintf("Could not list datasets: %s", err.Error()),
			)
			return
		}

		for _, dataset := range listResp.Datasets {
			if dataset.DeletedAt != "" {
				continue
			}

			if nameFilter != "" && dataset.Name != nameFilter {
				continue
			}

			var metadataMap types.Map
			if len(dataset.Metadata) > 0 {
				metadata := make(map[string]string)
				for k, v := range dataset.Metadata {
					metadata[k] = fmt.Sprintf("%v", v)
				}
				var diags diag.Diagnostics
				metadataMap, diags = types.MapValueFrom(ctx, types.StringType, metadata)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			} else {
				metadataMap = types.MapNull(types.StringType)
			}

			var tagsSet types.Set
			if len(dataset.Tags) > 0 {
				var diags diag.Diagnostics
				tagsSet, diags = types.SetValueFrom(ctx, types.StringType, dataset.Tags)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			} else {
				tagsSet = types.SetNull(types.StringType)
			}

			datasetModel := DatasetsDataSourceDataset{
				ID:          types.StringValue(dataset.ID),
				Name:        types.StringValue(dataset.Name),
				ProjectID:   types.StringValue(dataset.ProjectID),
				Description: types.StringValue(dataset.Description),
				Created:     types.StringValue(dataset.Created),
				UserID:      types.StringValue(dataset.UserID),
				OrgID:       types.StringValue(dataset.OrgID),
				Metadata:    metadataMap,
				Tags:        tagsSet,
			}

			data.Datasets = append(data.Datasets, datasetModel)
			data.IDs = append(data.IDs, dataset.ID)
		}

		// Exit loop if no more pages
		if listResp.Cursor == "" {
			break
		}
		cursor = listResp.Cursor
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
