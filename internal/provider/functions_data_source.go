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

var _ datasource.DataSource = &FunctionsDataSource{}

// NewFunctionsDataSource creates a new functions data source instance.
func NewFunctionsDataSource() datasource.DataSource {
	return &FunctionsDataSource{}
}

// FunctionsDataSource defines the data source implementation.
type FunctionsDataSource struct {
	client *client.Client
}

// FunctionsDataSourceModel describes the data source data model.
type FunctionsDataSourceModel struct {
	ProjectID     types.String              `tfsdk:"project_id"`
	Name          types.String              `tfsdk:"name"`
	Slug          types.String              `tfsdk:"slug"`
	StartingAfter types.String              `tfsdk:"starting_after"`
	EndingBefore  types.String              `tfsdk:"ending_before"`
	Functions     []FunctionsDataSourceItem `tfsdk:"functions"`
	IDs           []string                  `tfsdk:"ids"`
	Limit         types.Int64               `tfsdk:"limit"`
}

// FunctionsDataSourceItem represents a single function in the list.
type FunctionsDataSourceItem struct {
	Metadata       types.Map    `tfsdk:"metadata"`
	Tags           types.Set    `tfsdk:"tags"`
	XactID         types.String `tfsdk:"xact_id"`
	Created        types.String `tfsdk:"created"`
	Description    types.String `tfsdk:"description"`
	FunctionData   types.String `tfsdk:"function_data"`
	FunctionSchema types.String `tfsdk:"function_schema"`
	FunctionType   types.String `tfsdk:"function_type"`
	ID             types.String `tfsdk:"id"`
	LogID          types.String `tfsdk:"log_id"`
	Name           types.String `tfsdk:"name"`
	OrgID          types.String `tfsdk:"org_id"`
	Origin         types.String `tfsdk:"origin"`
	ProjectID      types.String `tfsdk:"project_id"`
	PromptData     types.String `tfsdk:"prompt_data"`
	Slug           types.String `tfsdk:"slug"`
}

// Metadata implements datasource.DataSource.
func (d *FunctionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_functions"
}

// Schema implements datasource.DataSource.
func (d *FunctionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust functions using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional project ID filter.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact function name filter. Maps to API query parameter `function_name`.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact function slug filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of functions to return. Supports `0`.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch functions after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch functions before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned function IDs.",
			},
			"functions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of functions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the function.",
						},
						"xact_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The transactional ID associated with the function.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the function was created.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of the function.",
						},
						"function_data": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The function data as a JSON-encoded string.",
						},
						"function_schema": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The function schema as a JSON-encoded string.",
						},
						"function_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The function type.",
						},
						"log_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The log ID associated with the function.",
						},
						"metadata": schema.MapAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Metadata associated with the function as key-value pairs.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the function.",
						},
						"org_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the organization the function belongs to.",
						},
						"origin": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The function origin as a JSON-encoded string.",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the project the function belongs to.",
						},
						"prompt_data": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The prompt data as a JSON-encoded string.",
						},
						"slug": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The slug of the function.",
						},
						"tags": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Tags associated with the function.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *FunctionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FunctionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FunctionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListFunctionsOptions(data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListFunctions(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Functions",
			fmt.Sprintf("Could not list functions: %s", err.Error()),
		)
		return
	}

	data.Functions = make([]FunctionsDataSourceItem, 0, len(listResp.Functions))
	data.IDs = make([]string, 0, len(listResp.Functions))

	for i := range listResp.Functions {
		function := &listResp.Functions[i]

		item, itemDiags := functionListItemFromFunction(ctx, function)
		resp.Diagnostics.Append(itemDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Functions = append(data.Functions, item)
		data.IDs = append(data.IDs, function.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListFunctionsOptions(data FunctionsDataSourceModel) (*client.ListFunctionsOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	projectID := strings.TrimSpace(data.ProjectID.ValueString())
	name := strings.TrimSpace(data.Name.ValueString())
	slug := strings.TrimSpace(data.Slug.ValueString())
	startingAfter := strings.TrimSpace(data.StartingAfter.ValueString())
	endingBefore := strings.TrimSpace(data.EndingBefore.ValueString())
	hasStartingAfter := !data.StartingAfter.IsNull() && startingAfter != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && endingBefore != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	listOpts := &client.ListFunctionsOptions{}
	if !data.ProjectID.IsNull() && projectID != "" {
		listOpts.ProjectID = projectID
	}
	if !data.Name.IsNull() && name != "" {
		listOpts.FunctionName = name
	}
	if !data.Slug.IsNull() && slug != "" {
		listOpts.Slug = slug
	}

	if !data.Limit.IsNull() {
		limit := data.Limit.ValueInt64()
		if limit < 0 {
			diags.AddError("Invalid Limit", "'limit' must be greater than or equal to 0.")
			return nil, diags
		}

		maxInt := int64(^uint(0) >> 1)
		if limit > maxInt {
			diags.AddError("Invalid Limit", "'limit' exceeds supported platform integer size.")
			return nil, diags
		}

		limitInt := int(limit)
		listOpts.Limit = &limitInt
	}
	if hasStartingAfter {
		listOpts.StartingAfter = startingAfter
	}
	if hasEndingBefore {
		listOpts.EndingBefore = endingBefore
	}

	return listOpts, diags
}

func functionListItemFromFunction(ctx context.Context, function *client.Function) (FunctionsDataSourceItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	item := FunctionsDataSourceItem{
		ID:           stringOrNull(function.ID),
		XactID:       stringOrNull(function.XactID),
		Created:      stringOrNull(function.Created),
		Description:  stringOrNull(function.Description),
		FunctionType: stringOrNull(function.FunctionType),
		LogID:        stringOrNull(function.LogID),
		Name:         stringOrNull(function.Name),
		OrgID:        stringOrNull(function.OrgID),
		ProjectID:    stringOrNull(function.ProjectID),
		Slug:         stringOrNull(function.Slug),
	}

	functionData, functionDataDiags := jsonEncodedOrNull("function_data", function.FunctionData)
	diags.Append(functionDataDiags...)
	item.FunctionData = functionData

	functionSchema, functionSchemaDiags := jsonEncodedOrNull("function_schema", function.FunctionSchema)
	diags.Append(functionSchemaDiags...)
	item.FunctionSchema = functionSchema

	origin, originDiags := jsonEncodedOrNull("origin", function.Origin)
	diags.Append(originDiags...)
	item.Origin = origin

	promptData, promptDataDiags := jsonEncodedOrNull("prompt_data", function.PromptData)
	diags.Append(promptDataDiags...)
	item.PromptData = promptData

	if diags.HasError() {
		return item, diags
	}

	if len(function.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range function.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		metadataMap, metadataDiags := types.MapValueFrom(ctx, types.StringType, metadata)
		diags.Append(metadataDiags...)
		if diags.HasError() {
			return item, diags
		}
		item.Metadata = metadataMap
	} else {
		item.Metadata = types.MapNull(types.StringType)
	}

	if len(function.Tags) > 0 {
		tagsSet, tagsDiags := types.SetValueFrom(ctx, types.StringType, function.Tags)
		diags.Append(tagsDiags...)
		if diags.HasError() {
			return item, diags
		}
		item.Tags = tagsSet
	} else {
		item.Tags = types.SetNull(types.StringType)
	}

	return item, diags
}
