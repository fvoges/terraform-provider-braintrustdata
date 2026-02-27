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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PromptDataSource{}

var (
	errPromptNotFoundByName       = errors.New("prompt not found by name")
	errMultiplePromptsFoundByName = errors.New("multiple prompts found by name")
)

// NewPromptDataSource creates a new prompt data source instance.
func NewPromptDataSource() datasource.DataSource {
	return &PromptDataSource{}
}

// PromptDataSource defines the data source implementation.
type PromptDataSource struct {
	client *client.Client
}

// PromptDataSourceModel describes the data source data model.
type PromptDataSourceModel struct {
	Metadata     types.Map    `tfsdk:"metadata"`
	Tags         types.Set    `tfsdk:"tags"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProjectID    types.String `tfsdk:"project_id"`
	Slug         types.String `tfsdk:"slug"`
	Description  types.String `tfsdk:"description"`
	FunctionType types.String `tfsdk:"function_type"`
	Created      types.String `tfsdk:"created"`
	UserID       types.String `tfsdk:"user_id"`
	OrgID        types.String `tfsdk:"org_id"`
}

// Metadata implements datasource.DataSource.
func (d *PromptDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

// Schema implements datasource.DataSource.
func (d *PromptDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust prompt by `id` or by searchable attributes (`name` and `project_id`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the prompt. Specify either `id` or both `name` and `project_id`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The prompt name. Must be specified with `project_id` when `id` is not provided.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID that scopes prompt lookup by name.",
			},
			"slug": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The prompt slug.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of the prompt.",
			},
			"function_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The function type associated with the prompt.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the prompt was created.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the prompt.",
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the organization the prompt belongs to.",
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Metadata associated with the prompt as key-value pairs.",
			},
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Tags associated with the prompt.",
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *PromptDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PromptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PromptDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, name, projectID, hasID, hasName, hasProjectID := normalizedPromptLookupInput(data)

	if !hasID && (!hasName || !hasProjectID) {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or both 'name' and 'project_id' to look up the prompt.",
		)
		return
	}

	if hasID && (hasName || hasProjectID) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'project_id').",
		)
		return
	}

	var prompt *client.Prompt
	if hasID {
		fetchedPrompt, err := d.client.GetPrompt(ctx, id)
		if err != nil {
			if client.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"Prompt Not Found",
					fmt.Sprintf("No prompt found with ID: %s", id),
				)
				return
			}

			resp.Diagnostics.AddError(
				"Error Reading Prompt",
				fmt.Sprintf("Could not read prompt ID %s: %s", id, err.Error()),
			)
			return
		}
		prompt = fetchedPrompt
	} else {
		listResp, err := d.client.ListPrompts(ctx, &client.ListPromptsOptions{
			ProjectID:  projectID,
			PromptName: name,
			Limit:      2,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Prompts",
				fmt.Sprintf("Could not list prompts using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedPrompt, err := selectSinglePromptByName(listResp.Prompts, name)
		if errors.Is(err, errPromptNotFoundByName) {
			resp.Diagnostics.AddError(
				"Prompt Not Found",
				fmt.Sprintf("No prompt found with name: %s in project: %s", name, projectID),
			)
			return
		}
		if errors.Is(err, errMultiplePromptsFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple Prompts Found",
				"Searchable attributes matched multiple prompts. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Prompts",
				fmt.Sprintf("Could not resolve prompt using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		prompt = selectedPrompt
	}

	if prompt.DeletedAt != "" {
		resp.Diagnostics.AddError(
			"Prompt Deleted",
			fmt.Sprintf("Prompt %s has been deleted", prompt.ID),
		)
		return
	}

	resp.Diagnostics.Append(populatePromptDataSourceModel(ctx, &data, prompt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func normalizedPromptLookupInput(data PromptDataSourceModel) (string, string, string, bool, bool, bool) {
	id := strings.TrimSpace(data.ID.ValueString())
	name := strings.TrimSpace(data.Name.ValueString())
	projectID := strings.TrimSpace(data.ProjectID.ValueString())

	hasID := !data.ID.IsNull() && id != ""
	hasName := !data.Name.IsNull() && name != ""
	hasProjectID := !data.ProjectID.IsNull() && projectID != ""

	return id, name, projectID, hasID, hasName, hasProjectID
}

func selectSinglePromptByName(prompts []client.Prompt, promptName string) (*client.Prompt, error) {
	var selected *client.Prompt

	for i := range prompts {
		prompt := &prompts[i]
		if prompt.DeletedAt != "" || prompt.Name != promptName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultiplePromptsFoundByName, promptName)
		}
		selected = prompt
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errPromptNotFoundByName, promptName)
	}

	return selected, nil
}

func populatePromptDataSourceModel(ctx context.Context, data *PromptDataSourceModel, prompt *client.Prompt) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = stringOrNull(prompt.ID)
	data.Name = stringOrNull(prompt.Name)
	data.ProjectID = stringOrNull(prompt.ProjectID)
	data.Slug = stringOrNull(prompt.Slug)
	data.Description = stringOrNull(prompt.Description)
	data.FunctionType = stringOrNull(prompt.FunctionType)
	data.Created = stringOrNull(prompt.Created)
	data.UserID = stringOrNull(prompt.UserID)
	data.OrgID = stringOrNull(prompt.OrgID)

	if len(prompt.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range prompt.Metadata {
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

	if len(prompt.Tags) > 0 {
		tagsSet, tagsDiags := types.SetValueFrom(ctx, types.StringType, prompt.Tags)
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
