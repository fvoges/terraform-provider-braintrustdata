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

var _ datasource.DataSource = &PromptsDataSource{}

// NewPromptsDataSource creates a new prompts data source instance.
func NewPromptsDataSource() datasource.DataSource {
	return &PromptsDataSource{}
}

// PromptsDataSource defines the data source implementation.
type PromptsDataSource struct {
	client *client.Client
}

// PromptsDataSourceModel describes the data source data model.
type PromptsDataSourceModel struct {
	ProjectID     types.String              `tfsdk:"project_id"`
	Name          types.String              `tfsdk:"name"`
	Slug          types.String              `tfsdk:"slug"`
	Version       types.String              `tfsdk:"version"`
	StartingAfter types.String              `tfsdk:"starting_after"`
	EndingBefore  types.String              `tfsdk:"ending_before"`
	Prompts       []PromptsDataSourcePrompt `tfsdk:"prompts"`
	IDs           []string                  `tfsdk:"ids"`
	Limit         types.Int64               `tfsdk:"limit"`
}

// PromptsDataSourcePrompt represents a single prompt in the list.
type PromptsDataSourcePrompt struct {
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
func (d *PromptsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompts"
}

// Schema implements datasource.DataSource.
func (d *PromptsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust prompts using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project ID to scope prompt listing.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact prompt name filter.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact prompt slug filter.",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional prompt version filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of prompts to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch prompts after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch prompts before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned prompt IDs.",
			},
			"prompts": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of prompts.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the prompt.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the prompt.",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project ID that owns the prompt.",
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
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *PromptsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PromptsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PromptsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListPromptsOptions(data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListPrompts(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Prompts",
			fmt.Sprintf("Could not list prompts: %s", err.Error()),
		)
		return
	}

	data.Prompts = make([]PromptsDataSourcePrompt, 0, len(listResp.Prompts))
	data.IDs = make([]string, 0, len(listResp.Prompts))

	for i := range listResp.Prompts {
		prompt := &listResp.Prompts[i]
		if prompt.DeletedAt != "" {
			continue
		}

		promptModel, promptDiags := promptsDataSourcePromptFromPrompt(ctx, prompt)
		resp.Diagnostics.Append(promptDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Prompts = append(data.Prompts, promptModel)
		data.IDs = append(data.IDs, prompt.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListPromptsOptions(data PromptsDataSourceModel) (*client.ListPromptsOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	projectID := strings.TrimSpace(data.ProjectID.ValueString())
	if projectID == "" {
		diags.AddError("Invalid project_id", "'project_id' must be provided and non-empty.")
		return nil, diags
	}

	listOpts := &client.ListPromptsOptions{ProjectID: projectID}

	name := strings.TrimSpace(data.Name.ValueString())
	slug := strings.TrimSpace(data.Slug.ValueString())
	version := strings.TrimSpace(data.Version.ValueString())

	if !data.Name.IsNull() && name != "" {
		listOpts.PromptName = name
	}
	if !data.Slug.IsNull() && slug != "" {
		listOpts.Slug = slug
	}
	if !data.Version.IsNull() && version != "" {
		listOpts.Version = version
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

func promptsDataSourcePromptFromPrompt(ctx context.Context, prompt *client.Prompt) (PromptsDataSourcePrompt, diag.Diagnostics) {
	var diags diag.Diagnostics

	promptModel := PromptsDataSourcePrompt{
		ID:           stringOrNull(prompt.ID),
		Name:         stringOrNull(prompt.Name),
		ProjectID:    stringOrNull(prompt.ProjectID),
		Slug:         stringOrNull(prompt.Slug),
		Description:  stringOrNull(prompt.Description),
		FunctionType: stringOrNull(prompt.FunctionType),
		Created:      stringOrNull(prompt.Created),
		UserID:       stringOrNull(prompt.UserID),
		OrgID:        stringOrNull(prompt.OrgID),
	}

	if len(prompt.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range prompt.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		metadataMap, metadataDiags := types.MapValueFrom(ctx, types.StringType, metadata)
		diags.Append(metadataDiags...)
		if diags.HasError() {
			return promptModel, diags
		}
		promptModel.Metadata = metadataMap
	} else {
		promptModel.Metadata = types.MapNull(types.StringType)
	}

	if len(prompt.Tags) > 0 {
		tagsSet, tagsDiags := types.SetValueFrom(ctx, types.StringType, prompt.Tags)
		diags.Append(tagsDiags...)
		if diags.HasError() {
			return promptModel, diags
		}
		promptModel.Tags = tagsSet
	} else {
		promptModel.Tags = types.SetNull(types.StringType)
	}

	return promptModel, diags
}
