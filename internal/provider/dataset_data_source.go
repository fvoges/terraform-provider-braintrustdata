package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DatasetDataSource{}

// NewDatasetDataSource creates a new dataset data source instance.
func NewDatasetDataSource() datasource.DataSource {
	return &DatasetDataSource{}
}

// DatasetDataSource defines the data source implementation.
type DatasetDataSource struct {
	client *client.Client
}

// DatasetDataSourceModel describes the data source data model.
type DatasetDataSourceModel struct {
	Tags        types.Set    `tfsdk:"tags"`
	Metadata    types.Map    `tfsdk:"metadata"`
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	UserID      types.String `tfsdk:"user_id"`
	OrgID       types.String `tfsdk:"org_id"`
	Public      types.Bool   `tfsdk:"public"`
}

// Metadata implements datasource.DataSource.
func (d *DatasetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

// Schema implements datasource.DataSource.
func (d *DatasetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust dataset by ID or by name and project_id. Specify either `id` or both `name` and `project_id`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the dataset. Specify either `id` or both `name` and `project_id`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the dataset. Must be specified with `project_id` when not using `id`.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the dataset belongs to. Must be specified with `name` when not using `id`.",
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
			"public": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the dataset is publicly accessible.",
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
	}
}

// Configure implements datasource.DataSource.
func (d *DatasetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatasetDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""
	hasProjectID := !data.ProjectID.IsNull() && data.ProjectID.ValueString() != ""

	if !hasID && (!hasName || !hasProjectID) {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or both 'name' and 'project_id' to look up the dataset.",
		)
		return
	}

	if hasID && (hasName || hasProjectID) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot specify both 'id' and 'name'",
		)
		return
	}

	var dataset *client.Dataset
	var err error

	if hasID {
		dataset, err = d.client.GetDataset(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Dataset",
				fmt.Sprintf("Could not read dataset ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
	} else {
		// Paginate through all datasets to find the one with matching name
		var found *client.Dataset
		cursor := ""
		for {
			listResp, err := d.client.ListDatasets(ctx, &client.ListDatasetsOptions{
				ProjectID: data.ProjectID.ValueString(),
				Cursor:    cursor,
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Listing Datasets",
					fmt.Sprintf("Could not list datasets to find name %s: %s", data.Name.ValueString(), err.Error()),
				)
				return
			}

			// Scan current page for matching dataset
			for i := range listResp.Datasets {
				if listResp.Datasets[i].Name == data.Name.ValueString() {
					if listResp.Datasets[i].DeletedAt == "" {
						found = &listResp.Datasets[i]
						break
					}
				}
			}

			// Exit loop if found or no more pages
			if found != nil || listResp.Cursor == "" {
				break
			}
			cursor = listResp.Cursor
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Dataset Not Found",
				fmt.Sprintf("No dataset found with name: %s in project: %s", data.Name.ValueString(), data.ProjectID.ValueString()),
			)
			return
		}

		dataset = found
	}

	if dataset.DeletedAt != "" {
		resp.Diagnostics.AddError(
			"Dataset Deleted",
			fmt.Sprintf("Dataset %s has been deleted", dataset.ID),
		)
		return
	}

	data.ID = types.StringValue(dataset.ID)
	data.Name = types.StringValue(dataset.Name)
	data.ProjectID = types.StringValue(dataset.ProjectID)
	data.Description = types.StringValue(dataset.Description)
	data.Created = types.StringValue(dataset.Created)
	data.UserID = types.StringValue(dataset.UserID)
	data.OrgID = types.StringValue(dataset.OrgID)
	data.Public = types.BoolValue(dataset.Public)

	if len(dataset.Metadata) > 0 {
		metadataMap := make(map[string]string)
		for k, v := range dataset.Metadata {
			metadataMap[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	if len(dataset.Tags) > 0 {
		tagsSet, diags := types.SetValueFrom(ctx, types.StringType, dataset.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tagsSet
	} else {
		data.Tags = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
