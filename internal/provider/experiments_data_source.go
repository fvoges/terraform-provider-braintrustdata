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

var _ datasource.DataSource = &ExperimentsDataSource{}

// NewExperimentsDataSource creates a new experiments data source instance.
func NewExperimentsDataSource() datasource.DataSource {
	return &ExperimentsDataSource{}
}

// ExperimentsDataSource defines the data source implementation.
type ExperimentsDataSource struct {
	client *client.Client
}

// ExperimentsDataSourceModel describes the data source data model.
type ExperimentsDataSourceModel struct {
	ProjectID   types.String                     `tfsdk:"project_id"`
	Name        types.String                     `tfsdk:"name"`
	Experiments []ExperimentsDataSourceExperiment `tfsdk:"experiments"`
	IDs         []string                         `tfsdk:"ids"`
}

// ExperimentsDataSourceExperiment represents a single experiment in the list.
type ExperimentsDataSourceExperiment struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	Public      types.Bool   `tfsdk:"public"`
	Metadata    types.Map    `tfsdk:"metadata"`
	Tags        types.Set    `tfsdk:"tags"`
}

// Metadata implements datasource.DataSource.
func (d *ExperimentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_experiments"
}

// Schema implements datasource.DataSource.
func (d *ExperimentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Braintrust experiments in a project. Optionally filter by experiment name.",

		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project ID to filter experiments.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional name filter to return only experiments with this exact name.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of experiment IDs.",
			},
			"experiments": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of experiments.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the experiment.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the experiment.",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project ID the experiment belongs to.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of the experiment.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the experiment was created.",
						},
						"public": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the experiment is public.",
						},
						"metadata": schema.MapAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Metadata associated with the experiment.",
						},
						"tags": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Tags associated with the experiment.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *ExperimentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExperimentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExperimentsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"project_id is required to list experiments.",
		)
		return
	}

	listResp, err := d.client.ListExperiments(ctx, &client.ListExperimentsOptions{
		ProjectID: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Experiments",
			fmt.Sprintf("Could not list experiments: %s", err.Error()),
		)
		return
	}

	nameFilter := ""
	if !data.Name.IsNull() && data.Name.ValueString() != "" {
		nameFilter = data.Name.ValueString()
	}

	data.Experiments = make([]ExperimentsDataSourceExperiment, 0)
	data.IDs = make([]string, 0)

	for _, experiment := range listResp.Experiments {
		if experiment.DeletedAt != "" {
			continue
		}

		if nameFilter != "" && experiment.Name != nameFilter {
			continue
		}

		var metadataMap types.Map
		if len(experiment.Metadata) > 0 {
			metadata := make(map[string]string)
			for k, v := range experiment.Metadata {
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
		if len(experiment.Tags) > 0 {
			var diags diag.Diagnostics
			tagsSet, diags = types.SetValueFrom(ctx, types.StringType, experiment.Tags)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			tagsSet = types.SetNull(types.StringType)
		}

		experimentModel := ExperimentsDataSourceExperiment{
			ID:          types.StringValue(experiment.ID),
			Name:        types.StringValue(experiment.Name),
			ProjectID:   types.StringValue(experiment.ProjectID),
			Description: types.StringValue(experiment.Description),
			Created:     types.StringValue(experiment.Created),
			Public:      types.BoolValue(experiment.Public),
			Metadata:    metadataMap,
			Tags:        tagsSet,
		}

		data.Experiments = append(data.Experiments, experimentModel)
		data.IDs = append(data.IDs, experiment.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
