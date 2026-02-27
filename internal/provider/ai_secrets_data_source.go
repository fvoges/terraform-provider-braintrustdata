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

var _ datasource.DataSource = &AISecretsDataSource{}

// NewAISecretsDataSource creates a new AI secrets data source instance.
func NewAISecretsDataSource() datasource.DataSource {
	return &AISecretsDataSource{}
}

// AISecretsDataSource defines the data source implementation.
type AISecretsDataSource struct {
	client *client.Client
}

// AISecretsDataSourceModel describes the data source data model.
type AISecretsDataSourceModel struct {
	OrgName       types.String                `tfsdk:"org_name"`
	AISecretName  types.String                `tfsdk:"ai_secret_name"`
	AISecretTypes types.List                  `tfsdk:"ai_secret_types"`
	FilterIDs     types.List                  `tfsdk:"filter_ids"`
	StartingAfter types.String                `tfsdk:"starting_after"`
	EndingBefore  types.String                `tfsdk:"ending_before"`
	AISecrets     []AISecretsDataSourceSecret `tfsdk:"ai_secrets"`
	IDs           []string                    `tfsdk:"ids"`
	Limit         types.Int64                 `tfsdk:"limit"`
}

// AISecretsDataSourceSecret represents a single AI secret in the list.
type AISecretsDataSourceSecret struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	OrgID         types.String `tfsdk:"org_id"`
	Type          types.String `tfsdk:"type"`
	Metadata      types.Map    `tfsdk:"metadata"`
	PreviewSecret types.String `tfsdk:"preview_secret"`
	Created       types.String `tfsdk:"created"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

// Metadata implements datasource.DataSource.
func (d *AISecretsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_secrets"
}

// Schema implements datasource.DataSource.
func (d *AISecretsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust AI secrets using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter.",
			},
			"ai_secret_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact AI secret name filter.",
			},
			"ai_secret_types": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional AI secret type filters.",
			},
			"filter_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional list of AI secret IDs to filter by. Maps to repeated `ids` query parameters.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of AI secrets to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch AI secrets after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch AI secrets before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned AI secret IDs.",
			},
			"ai_secrets": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of AI secrets.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the AI secret.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The AI secret name.",
						},
						"org_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The organization ID that the AI secret belongs to.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The AI secret type.",
						},
						"metadata": schema.MapAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Metadata associated with the AI secret.",
						},
						"preview_secret": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Preview of the secret value.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the AI secret was created.",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the AI secret was last updated.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *AISecretsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AISecretsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AISecretsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListAISecretsOptions(ctx, data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListAISecrets(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing AI Secrets",
			fmt.Sprintf("Could not list AI secrets: %s", err.Error()),
		)
		return
	}

	data.AISecrets = make([]AISecretsDataSourceSecret, 0, len(listResp.AISecrets))
	data.IDs = make([]string, 0, len(listResp.AISecrets))

	for i := range listResp.AISecrets {
		aiSecret := &listResp.AISecrets[i]

		aiSecretModel, conversionDiags := aiSecretsDataSourceAISecretFromAISecret(ctx, aiSecret)
		resp.Diagnostics.Append(conversionDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.AISecrets = append(data.AISecrets, aiSecretModel)
		data.IDs = append(data.IDs, aiSecret.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListAISecretsOptions(ctx context.Context, data AISecretsDataSourceModel) (*client.ListAISecretsOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	listOpts := &client.ListAISecretsOptions{}

	if !data.OrgName.IsNull() && data.OrgName.ValueString() != "" {
		listOpts.OrgName = data.OrgName.ValueString()
	}
	if !data.AISecretName.IsNull() && data.AISecretName.ValueString() != "" {
		listOpts.AISecretName = data.AISecretName.ValueString()
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
	if !data.FilterIDs.IsNull() {
		var filterIDs []string
		diags.Append(data.FilterIDs.ElementsAs(ctx, &filterIDs, false)...)
		if diags.HasError() {
			return nil, diags
		}

		for _, id := range filterIDs {
			if strings.TrimSpace(id) == "" {
				diags.AddError("Invalid filter_ids", "'filter_ids' cannot contain empty values.")
				return nil, diags
			}
		}

		listOpts.IDs = filterIDs
	}
	if !data.AISecretTypes.IsNull() {
		var aiSecretTypes []string
		diags.Append(data.AISecretTypes.ElementsAs(ctx, &aiSecretTypes, false)...)
		if diags.HasError() {
			return nil, diags
		}

		for _, aiSecretType := range aiSecretTypes {
			if strings.TrimSpace(aiSecretType) == "" {
				diags.AddError("Invalid ai_secret_types", "'ai_secret_types' cannot contain empty values.")
				return nil, diags
			}
		}

		listOpts.AISecretTypes = aiSecretTypes
	}

	return listOpts, diags
}

func aiSecretsDataSourceAISecretFromAISecret(ctx context.Context, aiSecret *client.AISecret) (AISecretsDataSourceSecret, diag.Diagnostics) {
	var diags diag.Diagnostics

	metadataValue, metadataDiags := aiSecretMetadataToTerraformMap(ctx, aiSecret.Metadata)
	diags.Append(metadataDiags...)
	if diags.HasError() {
		return AISecretsDataSourceSecret{}, diags
	}

	return AISecretsDataSourceSecret{
		ID:            types.StringValue(aiSecret.ID),
		Name:          types.StringValue(aiSecret.Name),
		OrgID:         stringOrNull(aiSecret.OrgID),
		Type:          stringOrNull(aiSecret.Type),
		Metadata:      metadataValue,
		PreviewSecret: stringOrNull(aiSecret.PreviewSecret),
		Created:       stringOrNull(aiSecret.Created),
		UpdatedAt:     stringOrNull(aiSecret.UpdatedAt),
	}, diags
}
