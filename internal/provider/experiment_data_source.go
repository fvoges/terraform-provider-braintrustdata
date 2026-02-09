package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ExperimentDataSource{}

// NewExperimentDataSource creates a new experiment data source instance.
func NewExperimentDataSource() datasource.DataSource {
	return &ExperimentDataSource{}
}

// ExperimentDataSource defines the data source implementation.
type ExperimentDataSource struct {
	client *client.Client
}

// ExperimentDataSourceModel describes the data source data model.
type ExperimentDataSourceModel struct {
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
func (d *ExperimentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_experiment"
}

// Schema implements datasource.DataSource.
func (d *ExperimentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust experiment by ID or by name and project_id. Specify either `id` or both `name` and `project_id`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the experiment. Specify either `id` or both `name` and `project_id`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the experiment. Must be specified with `project_id` when not using `id`.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the experiment belongs to. Must be specified with `name` when not using `id`.",
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
	}
}

// Configure implements datasource.DataSource.
func (d *ExperimentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExperimentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExperimentDataSourceModel

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
			"Must specify either 'id' or both 'name' and 'project_id' to look up the experiment.",
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

	var experiment *client.Experiment
	var err error

	if hasID {
		experiment, err = d.client.GetExperiment(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Experiment",
				fmt.Sprintf("Could not read experiment ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
	} else {
		// Paginate through all experiments to find the one with matching name
		var found *client.Experiment
		cursor := ""
		for {
			listResp, err := d.client.ListExperiments(ctx, &client.ListExperimentsOptions{
				ProjectID: data.ProjectID.ValueString(),
				Cursor:    cursor,
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Listing Experiments",
					fmt.Sprintf("Could not list experiments to find name %s: %s", data.Name.ValueString(), err.Error()),
				)
				return
			}

			// Scan current page for matching experiment
			for i := range listResp.Experiments {
				if listResp.Experiments[i].Name == data.Name.ValueString() {
					if listResp.Experiments[i].DeletedAt == "" {
						found = &listResp.Experiments[i]
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
				"Experiment Not Found",
				fmt.Sprintf("No experiment found with name: %s in project: %s", data.Name.ValueString(), data.ProjectID.ValueString()),
			)
			return
		}

		experiment = found
	}

	if experiment.DeletedAt != "" {
		resp.Diagnostics.AddError(
			"Experiment Deleted",
			fmt.Sprintf("Experiment %s has been deleted", experiment.ID),
		)
		return
	}

	data.ID = types.StringValue(experiment.ID)
	data.Name = types.StringValue(experiment.Name)
	data.ProjectID = types.StringValue(experiment.ProjectID)
	data.Description = types.StringValue(experiment.Description)
	data.Created = types.StringValue(experiment.Created)
	data.Public = types.BoolValue(experiment.Public)

	if len(experiment.Metadata) > 0 {
		metadataMap := make(map[string]string)
		for k, v := range experiment.Metadata {
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

	if len(experiment.Tags) > 0 {
		tagsSet, diags := types.SetValueFrom(ctx, types.StringType, experiment.Tags)
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
