package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EnvironmentVariablesDataSource{}

// NewEnvironmentVariablesDataSource creates a new environment variables data source instance.
func NewEnvironmentVariablesDataSource() datasource.DataSource {
	return &EnvironmentVariablesDataSource{}
}

// EnvironmentVariablesDataSource defines the data source implementation.
type EnvironmentVariablesDataSource struct {
	client *client.Client
}

// EnvironmentVariablesDataSourceModel describes the data source data model.
type EnvironmentVariablesDataSourceModel struct {
	ObjectType           types.String                           `tfsdk:"object_type"`
	ObjectID             types.String                           `tfsdk:"object_id"`
	Name                 types.String                           `tfsdk:"name"`
	StartingAfter        types.String                           `tfsdk:"starting_after"`
	EndingBefore         types.String                           `tfsdk:"ending_before"`
	EnvironmentVariables []EnvironmentVariablesDataSourceEnvVar `tfsdk:"environment_variables"`
	IDs                  []string                               `tfsdk:"ids"`
	Limit                types.Int64                            `tfsdk:"limit"`
}

// EnvironmentVariablesDataSourceEnvVar represents a single environment variable in the list.
type EnvironmentVariablesDataSourceEnvVar struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ObjectType  types.String `tfsdk:"object_type"`
	ObjectID    types.String `tfsdk:"object_id"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *EnvironmentVariablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variables"
}

// Schema implements datasource.DataSource.
func (d *EnvironmentVariablesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust environment variables for an object. Optionally filter by exact `name`.",
		Attributes: map[string]schema.Attribute{
			"object_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object type that owns the environment variables (for example `project` or `function`).",
			},
			"object_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object ID that owns the environment variables.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact-name filter applied after retrieval.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of environment variables to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch environment variables after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch environment variables before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned environment variable IDs.",
			},
			"environment_variables": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of environment variables.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the environment variable.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The environment variable name.",
						},
						"object_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The object type that owns this environment variable.",
						},
						"object_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The owning object ID for this environment variable.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Optional description associated with the environment variable.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the environment variable was created.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *EnvironmentVariablesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EnvironmentVariablesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentVariablesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListEnvironmentVariablesOptions(data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListEnvironmentVariables(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Environment Variables",
			fmt.Sprintf("Could not list environment variables: %s", err.Error()),
		)
		return
	}

	nameFilter := ""
	if !data.Name.IsNull() && data.Name.ValueString() != "" {
		nameFilter = data.Name.ValueString()
	}

	data.EnvironmentVariables = make([]EnvironmentVariablesDataSourceEnvVar, 0, len(listResp.EnvironmentVariables))
	data.IDs = make([]string, 0, len(listResp.EnvironmentVariables))

	for i := range listResp.EnvironmentVariables {
		envVar := &listResp.EnvironmentVariables[i]
		if nameFilter != "" && envVar.Name != nameFilter {
			continue
		}

		envVarModel := environmentVariablesDataSourceEnvironmentVariableFromEnvironmentVariable(envVar)

		data.EnvironmentVariables = append(data.EnvironmentVariables, envVarModel)
		data.IDs = append(data.IDs, envVar.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListEnvironmentVariablesOptions(data EnvironmentVariablesDataSourceModel) (*client.ListEnvironmentVariablesOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	objectType := strings.TrimSpace(data.ObjectType.ValueString())
	if objectType == "" {
		diags.AddError("Invalid object_type", "'object_type' must be provided and non-empty.")
		return nil, diags
	}

	objectID := strings.TrimSpace(data.ObjectID.ValueString())
	if objectID == "" {
		diags.AddError("Invalid object_id", "'object_id' must be provided and non-empty.")
		return nil, diags
	}

	listOpts := &client.ListEnvironmentVariablesOptions{
		ObjectType: objectType,
		ObjectID:   objectID,
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

func environmentVariablesDataSourceEnvironmentVariableFromEnvironmentVariable(envVar *client.EnvironmentVariable) EnvironmentVariablesDataSourceEnvVar {
	envVarModel := EnvironmentVariablesDataSourceEnvVar{
		ID:   types.StringValue(envVar.ID),
		Name: types.StringValue(envVar.Name),
	}

	envVarModel.ObjectType = stringOrNull(envVar.ObjectType)
	envVarModel.ObjectID = stringOrNull(envVar.ObjectID)
	envVarModel.Description = stringOrNull(envVar.Description)
	envVarModel.Created = stringOrNull(envVar.Created)

	return envVarModel
}
