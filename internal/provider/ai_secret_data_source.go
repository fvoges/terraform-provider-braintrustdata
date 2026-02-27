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

var _ datasource.DataSource = &AISecretDataSource{}

var (
	errAISecretNotFoundByName       = errors.New("ai secret not found by name")
	errMultipleAISecretsFoundByName = errors.New("multiple ai secrets found by name")
)

// NewAISecretDataSource creates a new AI secret data source instance.
func NewAISecretDataSource() datasource.DataSource {
	return &AISecretDataSource{}
}

// AISecretDataSource defines the data source implementation.
type AISecretDataSource struct {
	client *client.Client
}

// AISecretDataSourceModel describes the data source data model.
type AISecretDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	OrgName       types.String `tfsdk:"org_name"`
	AISecretType  types.String `tfsdk:"ai_secret_type"`
	OrgID         types.String `tfsdk:"org_id"`
	Type          types.String `tfsdk:"type"`
	Metadata      types.Map    `tfsdk:"metadata"`
	PreviewSecret types.String `tfsdk:"preview_secret"`
	Created       types.String `tfsdk:"created"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

// Metadata implements datasource.DataSource.
func (d *AISecretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_secret"
}

// Schema implements datasource.DataSource.
func (d *AISecretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust AI secret by `id` or by API-native searchable attributes (`name`, optionally `org_name`, `ai_secret_type`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the AI secret. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The AI secret name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter applied during searchable lookups.",
			},
			"ai_secret_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional AI secret type filter applied during searchable lookups.",
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
	}
}

// Configure implements datasource.DataSource.
func (d *AISecretDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AISecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AISecretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lookupInputs := buildAISecretLookupInputs(data)
	resp.Diagnostics.Append(validateAISecretLookupInputs(lookupInputs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var aiSecret *client.AISecret
	if lookupInputs.hasID {
		fetchedAISecret, err := d.client.GetAISecret(ctx, lookupInputs.id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading AI Secret",
				fmt.Sprintf("Could not read AI secret ID %s: %s", lookupInputs.id, err.Error()),
			)
			return
		}

		aiSecret = fetchedAISecret
	} else {
		listOpts := &client.ListAISecretsOptions{
			AISecretName: lookupInputs.name,
			Limit:        2,
		}
		if lookupInputs.hasOrgName {
			listOpts.OrgName = lookupInputs.orgName
		}
		if lookupInputs.hasAISecretType {
			listOpts.AISecretTypes = []string{lookupInputs.aiSecretType}
		}

		listResp, err := d.client.ListAISecrets(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing AI Secrets",
				fmt.Sprintf("Could not list AI secrets using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedAISecret, err := selectSingleAISecretByName(listResp.AISecrets, lookupInputs.name)
		if errors.Is(err, errAISecretNotFoundByName) {
			resp.Diagnostics.AddError(
				"AI Secret Not Found",
				fmt.Sprintf("No AI secret found with name: %s", lookupInputs.name),
			)
			return
		}
		if errors.Is(err, errMultipleAISecretsFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple AI Secrets Found",
				"Searchable attributes matched multiple AI secrets. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing AI Secrets",
				fmt.Sprintf("Could not resolve AI secret using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		aiSecret = selectedAISecret
	}

	resp.Diagnostics.Append(populateAISecretDataSourceModel(ctx, &data, aiSecret)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type aiSecretLookupInputs struct {
	id              string
	name            string
	orgName         string
	aiSecretType    string
	hasID           bool
	hasName         bool
	hasOrgName      bool
	hasAISecretType bool
	nameProvided    bool
}

func buildAISecretLookupInputs(data AISecretDataSourceModel) aiSecretLookupInputs {
	id := strings.TrimSpace(data.ID.ValueString())
	name := strings.TrimSpace(data.Name.ValueString())
	orgName := strings.TrimSpace(data.OrgName.ValueString())
	aiSecretType := strings.TrimSpace(data.AISecretType.ValueString())

	return aiSecretLookupInputs{
		id:              id,
		name:            name,
		orgName:         orgName,
		aiSecretType:    aiSecretType,
		hasID:           id != "",
		hasName:         name != "",
		hasOrgName:      orgName != "",
		hasAISecretType: aiSecretType != "",
		nameProvided:    !data.Name.IsNull(),
	}
}

func validateAISecretLookupInputs(inputs aiSecretLookupInputs) diag.Diagnostics {
	var diags diag.Diagnostics

	if !inputs.hasID && !inputs.hasName {
		if inputs.nameProvided || inputs.hasOrgName || inputs.hasAISecretType {
			diags.AddError(
				"Invalid Lookup Name",
				"When using searchable lookup attributes, 'name' must be provided and non-empty.",
			)
		} else {
			diags.AddError(
				"Missing Required Attribute",
				"Must specify either 'id' or 'name' to look up the AI secret.",
			)
		}
		return diags
	}

	if inputs.hasID && (inputs.hasName || inputs.hasOrgName || inputs.hasAISecretType) {
		diags.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'org_name', 'ai_secret_type').",
		)
		return diags
	}

	return diags
}

func selectSingleAISecretByName(aiSecrets []client.AISecret, aiSecretName string) (*client.AISecret, error) {
	var selected *client.AISecret

	for i := range aiSecrets {
		aiSecret := &aiSecrets[i]
		if aiSecret.Name != aiSecretName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleAISecretsFoundByName, aiSecretName)
		}
		selected = aiSecret
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errAISecretNotFoundByName, aiSecretName)
	}

	return selected, nil
}

func populateAISecretDataSourceModel(ctx context.Context, data *AISecretDataSourceModel, aiSecret *client.AISecret) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(aiSecret.ID)
	data.Name = types.StringValue(aiSecret.Name)
	data.OrgID = stringOrNull(aiSecret.OrgID)
	data.Type = stringOrNull(aiSecret.Type)
	data.PreviewSecret = stringOrNull(aiSecret.PreviewSecret)
	data.Created = stringOrNull(aiSecret.Created)
	data.UpdatedAt = stringOrNull(aiSecret.UpdatedAt)

	metadataValue, metadataDiags := aiSecretMetadataToTerraformMap(ctx, aiSecret.Metadata)
	diags.Append(metadataDiags...)
	if diags.HasError() {
		return diags
	}

	data.Metadata = metadataValue

	return diags
}

func aiSecretMetadataToTerraformMap(ctx context.Context, metadata map[string]interface{}) (types.Map, diag.Diagnostics) {
	if len(metadata) == 0 {
		return types.MapNull(types.StringType), nil
	}

	metadataStrings := make(map[string]string, len(metadata))
	for key, value := range metadata {
		metadataStrings[key] = fmt.Sprintf("%v", value)
	}

	metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
	if diags.HasError() {
		return types.MapNull(types.StringType), diags
	}

	return metadataValue, diags
}
