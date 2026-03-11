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

var _ datasource.DataSource = &ScoreDataSource{}

var (
	errScoreNotFoundByName       = errors.New("score not found by name")
	errMultipleScoresFoundByName = errors.New("multiple scores found by name")
)

// NewScoreDataSource creates a new score data source instance.
func NewScoreDataSource() datasource.DataSource {
	return &ScoreDataSource{}
}

// ScoreDataSource defines the data source implementation.
type ScoreDataSource struct {
	client *client.Client
}

// ScoreDataSourceModel describes the data source data model.
type ScoreDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
	OrgName     types.String `tfsdk:"org_name"`
	UserID      types.String `tfsdk:"user_id"`
	Created     types.String `tfsdk:"created"`
	Description types.String `tfsdk:"description"`
	ScoreType   types.String `tfsdk:"score_type"`
	Categories  types.String `tfsdk:"categories"`
	Config      types.String `tfsdk:"config"`
	Position    types.String `tfsdk:"position"`
}

// Metadata implements datasource.DataSource.
func (d *ScoreDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_score"
}

// Schema implements datasource.DataSource.
func (d *ScoreDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust project score by `id` or by API-native searchable attributes (`name`, optionally `project_id`, `project_name`, `score_type`, `org_name`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the score. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The score name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID associated with the score. Can also be used as a searchable attribute.",
			},
			"project_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional project name filter applied during searchable lookups.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter applied during searchable lookups.",
			},
			"score_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The score type. Can also be used as a searchable attribute.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the score.",
			},
			"categories": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The score categories as a JSON-encoded string.",
			},
			"config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The score configuration as a JSON-encoded string.",
			},
			"position": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "LexoRank position of the score within the project.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the score.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the score was created.",
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *ScoreDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *ScoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ScoreDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := strings.TrimSpace(data.ID.ValueString())
	name := strings.TrimSpace(data.Name.ValueString())
	projectID := strings.TrimSpace(data.ProjectID.ValueString())
	projectName := strings.TrimSpace(data.ProjectName.ValueString())
	scoreType := strings.TrimSpace(data.ScoreType.ValueString())
	orgName := strings.TrimSpace(data.OrgName.ValueString())

	hasID := !data.ID.IsNull() && id != ""
	hasName := !data.Name.IsNull() && name != ""
	hasProjectID := !data.ProjectID.IsNull() && projectID != ""
	hasProjectName := !data.ProjectName.IsNull() && projectName != ""
	hasScoreType := !data.ScoreType.IsNull() && scoreType != ""
	hasOrgName := !data.OrgName.IsNull() && orgName != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the score.",
		)
		return
	}

	if hasID && (hasName || hasProjectID || hasProjectName || hasScoreType || hasOrgName) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'project_id', 'project_name', 'score_type', 'org_name').",
		)
		return
	}

	var score *client.ProjectScore
	if hasID {
		fetchedScore, err := d.client.GetScore(ctx, id)
		if err != nil {
			if client.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"Score Not Found",
					fmt.Sprintf("No score found with ID: %s", id),
				)
				return
			}

			resp.Diagnostics.AddError(
				"Error Reading Score",
				fmt.Sprintf("Could not read score ID %s: %s", id, err.Error()),
			)
			return
		}
		score = fetchedScore
	} else {
		listOpts := &client.ListScoresOptions{
			ScoreName: name,
			Limit:     2,
		}
		if hasProjectID {
			listOpts.ProjectID = projectID
		}
		if hasProjectName {
			listOpts.ProjectName = projectName
		}
		if hasScoreType {
			listOpts.ScoreType = scoreType
		}
		if hasOrgName {
			listOpts.OrgName = orgName
		}

		listResp, err := d.client.ListScores(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Scores",
				fmt.Sprintf("Could not list scores using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedScore, err := selectSingleScoreByName(listResp.Objects, name)
		if errors.Is(err, errScoreNotFoundByName) {
			resp.Diagnostics.AddError(
				"Score Not Found",
				fmt.Sprintf("No score found with name: %s", name),
			)
			return
		}
		if errors.Is(err, errMultipleScoresFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple Scores Found",
				"Searchable attributes matched multiple scores. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Scores",
				fmt.Sprintf("Could not resolve score using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		score = selectedScore
	}

	resp.Diagnostics.Append(populateScoreDataSourceModel(ctx, &data, score)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func selectSingleScoreByName(scores []client.ProjectScore, scoreName string) (*client.ProjectScore, error) {
	var selected *client.ProjectScore

	for i := range scores {
		score := &scores[i]
		if score.Name != scoreName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleScoresFoundByName, scoreName)
		}
		selected = score
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errScoreNotFoundByName, scoreName)
	}

	return selected, nil
}

func populateScoreDataSourceModel(_ context.Context, data *ScoreDataSourceModel, score *client.ProjectScore) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = stringOrNull(score.ID)
	data.Name = stringOrNull(score.Name)
	data.ProjectID = stringOrNull(score.ProjectID)
	data.UserID = stringOrNull(score.UserID)
	data.Created = stringOrNull(score.Created)
	data.Description = stringOrNull(score.Description)
	data.ScoreType = stringOrNull(score.ScoreType)

	categoriesValue, err := projectScoreJSONOrNull(score.Categories)
	if err != nil {
		diags.AddError(
			"Error Encoding categories",
			fmt.Sprintf("Unable to encode categories as JSON: %s", err),
		)
		return diags
	}
	data.Categories = categoriesValue

	configValue, err := projectScoreJSONOrNull(score.Config)
	if err != nil {
		diags.AddError(
			"Error Encoding config",
			fmt.Sprintf("Unable to encode config as JSON: %s", err),
		)
		return diags
	}
	data.Config = configValue

	if score.Position != nil {
		data.Position = stringOrNull(*score.Position)
	} else {
		data.Position = types.StringNull()
	}

	return diags
}

func projectScoreJSONOrNull(v interface{}) (types.String, error) {
	if v == nil {
		return types.StringNull(), nil
	}

	encoded, err := json.Marshal(v)
	if err != nil {
		return types.StringNull(), err
	}

	return types.StringValue(string(encoded)), nil
}
