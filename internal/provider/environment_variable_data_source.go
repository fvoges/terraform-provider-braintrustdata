package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EnvironmentVariableDataSource{}

var (
	errEnvironmentVariableNotFoundByName       = errors.New("environment variable not found by name")
	errMultipleEnvironmentVariablesFoundByName = errors.New("multiple environment variables found by name")
)

// NewEnvironmentVariableDataSource creates a new environment variable data source instance.
func NewEnvironmentVariableDataSource() datasource.DataSource {
	return &EnvironmentVariableDataSource{}
}

// EnvironmentVariableDataSource defines the data source implementation.
type EnvironmentVariableDataSource struct {
	client *client.Client
}

// EnvironmentVariableDataSourceModel describes the data source data model.
type EnvironmentVariableDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ObjectType  types.String `tfsdk:"object_type"`
	ObjectID    types.String `tfsdk:"object_id"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *EnvironmentVariableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

// Schema implements datasource.DataSource.
func (d *EnvironmentVariableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust environment variable by `id` or by (`name`, `object_type`, `object_id`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the environment variable. Specify either `id` or (`name`, `object_type`, `object_id`).",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The environment variable name. Used for lookup when `id` is not provided.",
			},
			"object_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The object type that owns this environment variable (for example `project` or `function`).",
			},
			"object_id": schema.StringAttribute{
				Optional:            true,
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
	}
}

// Configure implements datasource.DataSource.
func (d *EnvironmentVariableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EnvironmentVariableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentVariableDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, name, objectType, objectID := trimEnvironmentVariableLookupInputs(data)

	hasID := !data.ID.IsNull() && id != ""
	hasName := !data.Name.IsNull() && name != ""
	hasObjectType := !data.ObjectType.IsNull() && objectType != ""
	hasObjectID := !data.ObjectID.IsNull() && objectID != ""

	if hasID && (hasName || hasObjectType || hasObjectID) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with lookup attributes ('name', 'object_type', 'object_id').",
		)
		return
	}

	var envVar *client.EnvironmentVariable
	if hasID {
		fetchedEnvVar, err := d.client.GetEnvironmentVariable(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Environment Variable",
				fmt.Sprintf("Could not read environment variable ID %s: %s", id, err.Error()),
			)
			return
		}

		envVar = fetchedEnvVar
	} else {
		lookupDiags := validateEnvironmentVariableLookupAttributes(name, objectType, objectID)
		resp.Diagnostics.Append(lookupDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		listResp, err := d.client.ListEnvironmentVariables(ctx, &client.ListEnvironmentVariablesOptions{
			ObjectType: objectType,
			ObjectID:   objectID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Environment Variables",
				fmt.Sprintf("Could not list environment variables using the provided lookup attributes: %s", err.Error()),
			)
			return
		}

		selectedEnvVar, err := selectSingleEnvironmentVariableByName(listResp.EnvironmentVariables, name)
		if errors.Is(err, errEnvironmentVariableNotFoundByName) {
			resp.Diagnostics.AddError(
				"Environment Variable Not Found",
				fmt.Sprintf("No environment variable found with name: %s", name),
			)
			return
		}
		if errors.Is(err, errMultipleEnvironmentVariablesFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple Environment Variables Found",
				"Lookup attributes matched multiple environment variables. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Environment Variables",
				fmt.Sprintf("Could not resolve environment variable using the provided lookup attributes: %s", err.Error()),
			)
			return
		}

		envVar = selectedEnvVar
	}

	populateEnvironmentVariableDataSourceModel(&data, envVar)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func trimEnvironmentVariableLookupInputs(data EnvironmentVariableDataSourceModel) (id, name, objectType, objectID string) {
	id = strings.TrimSpace(data.ID.ValueString())
	name = strings.TrimSpace(data.Name.ValueString())
	objectType = strings.TrimSpace(data.ObjectType.ValueString())
	objectID = strings.TrimSpace(data.ObjectID.ValueString())

	return id, name, objectType, objectID
}

func validateEnvironmentVariableLookupAttributes(name, objectType, objectID string) diag.Diagnostics {
	var diags diag.Diagnostics

	if name == "" {
		diags.AddAttributeError(
			path.Root("name"),
			"Invalid name",
			"'name' must be provided and non-empty when using lookup mode.",
		)
	}
	if objectType == "" {
		diags.AddAttributeError(
			path.Root("object_type"),
			"Invalid object_type",
			"'object_type' must be provided and non-empty when using lookup mode.",
		)
	}
	if objectID == "" {
		diags.AddAttributeError(
			path.Root("object_id"),
			"Invalid object_id",
			"'object_id' must be provided and non-empty when using lookup mode.",
		)
	}

	return diags
}

func selectSingleEnvironmentVariableByName(envVars []client.EnvironmentVariable, envVarName string) (*client.EnvironmentVariable, error) {
	var selected *client.EnvironmentVariable

	for i := range envVars {
		envVar := &envVars[i]
		if envVar.Name != envVarName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleEnvironmentVariablesFoundByName, envVarName)
		}
		selected = envVar
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errEnvironmentVariableNotFoundByName, envVarName)
	}

	return selected, nil
}

func populateEnvironmentVariableDataSourceModel(data *EnvironmentVariableDataSourceModel, envVar *client.EnvironmentVariable) {
	data.ID = types.StringValue(envVar.ID)
	data.Name = types.StringValue(envVar.Name)
	data.ObjectType = stringOrNull(envVar.ObjectType)
	data.ObjectID = stringOrNull(envVar.ObjectID)
	data.Description = stringOrNull(envVar.Description)
	data.Created = stringOrNull(envVar.Created)
}
