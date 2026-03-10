package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FunctionDataSource{}

var (
	errFunctionNotFoundByField       = errors.New("function not found by field")
	errMultipleFunctionsFoundByField = errors.New("multiple functions found by field")
)

// NewFunctionDataSource creates a new function data source instance.
func NewFunctionDataSource() datasource.DataSource {
	return &FunctionDataSource{}
}

// FunctionDataSource defines the data source implementation.
type FunctionDataSource struct {
	client *client.Client
}

// FunctionDataSourceModel describes the data source data model.
type FunctionDataSourceModel struct {
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
func (d *FunctionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

// Schema implements datasource.DataSource.
func (d *FunctionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust function by `id` or by searchable attributes (`project_id` + `name` or `project_id` + `slug`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the function. Specify either `id` or one searchable pair (`project_id` + `name`, `project_id` + `slug`).",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID that scopes function lookup by searchable attributes.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The function name. Must be specified with `project_id` when `id` is not provided and `slug` is not used.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The function slug. Must be specified with `project_id` when `id` is not provided and `name` is not used.",
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
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the organization the function belongs to.",
			},
			"origin": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The function origin as a JSON-encoded string.",
			},
			"prompt_data": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The prompt data as a JSON-encoded string.",
			},
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Tags associated with the function.",
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *FunctionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FunctionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, projectID, name, slug, hasID, hasProjectID, hasName, hasSlug := normalizedFunctionLookupInput(data)
	resp.Diagnostics.Append(validateFunctionLookupInput(hasID, hasProjectID, hasName, hasSlug)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var function *client.Function

	if hasID {
		fetchedFunction, err := d.client.GetFunction(ctx, id)
		if err != nil {
			if client.IsFunctionNotFound(err) {
				resp.Diagnostics.AddError(
					"Function Not Found",
					fmt.Sprintf("No function found with ID: %s", id),
				)
				return
			}

			resp.Diagnostics.AddError(
				"Error Reading Function",
				fmt.Sprintf("Could not read function ID %s: %s", id, err.Error()),
			)
			return
		}
		function = fetchedFunction
	} else {
		listOpts := &client.ListFunctionsOptions{ProjectID: projectID}
		lookupFieldName := "name"
		lookupFieldValue := name
		if hasName {
			listOpts.FunctionName = name
		} else {
			lookupFieldName = "slug"
			lookupFieldValue = slug
			listOpts.Slug = slug
		}

		lookupLimit := 2
		listOpts.Limit = &lookupLimit

		listResp, err := d.client.ListFunctions(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Functions",
				fmt.Sprintf("Could not list functions using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedFunction, err := selectSingleFunctionByField(listResp.Functions, lookupFieldName, lookupFieldValue)
		if errors.Is(err, errFunctionNotFoundByField) {
			resp.Diagnostics.AddError(
				"Function Not Found",
				fmt.Sprintf("No function found with %s: %s in project: %s", lookupFieldName, lookupFieldValue, projectID),
			)
			return
		}
		if errors.Is(err, errMultipleFunctionsFoundByField) {
			resp.Diagnostics.AddError(
				"Multiple Functions Found",
				"Searchable attributes matched multiple functions. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Functions",
				fmt.Sprintf("Could not resolve function using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		function = selectedFunction
	}

	resp.Diagnostics.Append(populateFunctionDataSourceModel(ctx, &data, function)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func normalizedFunctionLookupInput(data FunctionDataSourceModel) (string, string, string, string, bool, bool, bool, bool) {
	id := strings.TrimSpace(data.ID.ValueString())
	projectID := strings.TrimSpace(data.ProjectID.ValueString())
	name := strings.TrimSpace(data.Name.ValueString())
	slug := strings.TrimSpace(data.Slug.ValueString())

	hasID := !data.ID.IsNull() && id != ""
	hasProjectID := !data.ProjectID.IsNull() && projectID != ""
	hasName := !data.Name.IsNull() && name != ""
	hasSlug := !data.Slug.IsNull() && slug != ""

	return id, projectID, name, slug, hasID, hasProjectID, hasName, hasSlug
}

func validateFunctionLookupInput(hasID, hasProjectID, hasName, hasSlug bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if hasID && (hasProjectID || hasName || hasSlug) {
		diags.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('project_id', 'name', 'slug').",
		)
		return diags
	}

	if hasName && hasSlug {
		diags.AddError(
			"Conflicting Attributes",
			"Cannot specify both 'name' and 'slug'.",
		)
		return diags
	}

	if (hasName || hasSlug) && !hasProjectID {
		diags.AddError(
			"Missing Required Attribute",
			"'project_id' must be provided when using 'name' or 'slug'.",
		)
		return diags
	}

	if !hasID && (!hasProjectID || (!hasName && !hasSlug)) {
		diags.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or one searchable pair ('project_id' + 'name', 'project_id' + 'slug').",
		)
	}

	return diags
}

func selectSingleFunctionByField(functions []client.Function, fieldName, fieldValue string) (*client.Function, error) {
	var selected *client.Function

	for i := range functions {
		function := &functions[i]

		candidateValue := ""
		switch fieldName {
		case "name":
			candidateValue = function.Name
		case "slug":
			candidateValue = function.Slug
		default:
			return nil, fmt.Errorf("unsupported function lookup field: %s", fieldName)
		}

		if candidateValue != fieldValue {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s=%s", errMultipleFunctionsFoundByField, fieldName, fieldValue)
		}

		selected = function
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s=%s", errFunctionNotFoundByField, fieldName, fieldValue)
	}

	return selected, nil
}

func populateFunctionDataSourceModel(ctx context.Context, data *FunctionDataSourceModel, function *client.Function) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = stringOrNull(function.ID)
	data.XactID = stringOrNull(function.XactID)
	data.Created = stringOrNull(function.Created)
	data.Description = stringOrNull(function.Description)
	data.FunctionType = stringOrNull(function.FunctionType)
	data.LogID = stringOrNull(function.LogID)
	data.Name = stringOrNull(function.Name)
	data.OrgID = stringOrNull(function.OrgID)
	data.ProjectID = stringOrNull(function.ProjectID)
	data.Slug = stringOrNull(function.Slug)

	functionData, functionDataDiags := jsonEncodedOrNull("function_data", function.FunctionData)
	diags.Append(functionDataDiags...)
	data.FunctionData = functionData

	functionSchema, functionSchemaDiags := jsonEncodedOrNull("function_schema", function.FunctionSchema)
	diags.Append(functionSchemaDiags...)
	data.FunctionSchema = functionSchema

	origin, originDiags := jsonEncodedOrNull("origin", function.Origin)
	diags.Append(originDiags...)
	data.Origin = origin

	promptData, promptDataDiags := jsonEncodedOrNull("prompt_data", function.PromptData)
	diags.Append(promptDataDiags...)
	data.PromptData = promptData

	if diags.HasError() {
		return diags
	}

	if len(function.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range function.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		metadataMap, metadataDiags := types.MapValueFrom(ctx, types.StringType, metadata)
		diags.Append(metadataDiags...)
		if diags.HasError() {
			return diags
		}
		data.Metadata = metadataMap
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	if len(function.Tags) > 0 {
		tagsSet, tagsDiags := types.SetValueFrom(ctx, types.StringType, function.Tags)
		diags.Append(tagsDiags...)
		if diags.HasError() {
			return diags
		}
		data.Tags = tagsSet
	} else {
		data.Tags = types.SetNull(types.StringType)
	}

	return diags
}

func jsonEncodedOrNull(fieldName string, v interface{}) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v == nil {
		return types.StringNull(), diags
	}

	encoded, err := json.Marshal(v)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error Encoding %s", fieldName),
			fmt.Sprintf("Could not encode %s to JSON: %s", fieldName, err),
		)
		return types.StringNull(), diags
	}

	return types.StringValue(string(encoded)), diags
}
