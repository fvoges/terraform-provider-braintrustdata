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

var _ datasource.DataSource = &ScoresDataSource{}

// NewScoresDataSource creates a new scores data source instance.
func NewScoresDataSource() datasource.DataSource {
	return &ScoresDataSource{}
}

// ScoresDataSource defines the data source implementation.
type ScoresDataSource struct {
	client *client.Client
}

// ScoresDataSourceModel describes the data source data model.
type ScoresDataSourceModel struct {
	FilterIDs     types.List              `tfsdk:"filter_ids"`
	OrgName       types.String            `tfsdk:"org_name"`
	ProjectID     types.String            `tfsdk:"project_id"`
	ProjectName   types.String            `tfsdk:"project_name"`
	ScoreName     types.String            `tfsdk:"score_name"`
	ScoreType     types.String            `tfsdk:"score_type"`
	StartingAfter types.String            `tfsdk:"starting_after"`
	EndingBefore  types.String            `tfsdk:"ending_before"`
	Scores        []ScoresDataSourceScore `tfsdk:"scores"`
	IDs           []string                `tfsdk:"ids"`
	Limit         types.Int64             `tfsdk:"limit"`
}

// ScoresDataSourceScore represents a single score in the list.
type ScoresDataSourceScore struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	UserID      types.String `tfsdk:"user_id"`
	Created     types.String `tfsdk:"created"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ScoreType   types.String `tfsdk:"score_type"`
	Categories  types.String `tfsdk:"categories"`
	Config      types.String `tfsdk:"config"`
	Position    types.String `tfsdk:"position"`
}

// Metadata implements datasource.DataSource.
func (d *ScoresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scores"
}

// Schema implements datasource.DataSource.
func (d *ScoresDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust project scores using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"filter_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional list of score IDs to filter by. Maps to repeated `ids` query parameters.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional project ID filter.",
			},
			"project_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional project name filter.",
			},
			"score_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact score name filter.",
			},
			"score_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional score type filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of scores to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch scores after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch scores before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned score IDs.",
			},
			"scores": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of scores.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the score.",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project ID that owns the score.",
						},
						"user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the user who created the score.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the score was created.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the score.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description of the score.",
						},
						"score_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Type of the score.",
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
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *ScoresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScoresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ScoresDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListScoresOptions(ctx, data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListScores(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Scores",
			fmt.Sprintf("Could not list scores: %s", err.Error()),
		)
		return
	}

	data.Scores = make([]ScoresDataSourceScore, 0, len(listResp.Objects))
	data.IDs = make([]string, 0, len(listResp.Objects))

	for i := range listResp.Objects {
		scoreModel, scoreDiags := scoresDataSourceScoreFromScore(&listResp.Objects[i])
		resp.Diagnostics.Append(scoreDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Scores = append(data.Scores, scoreModel)
		data.IDs = append(data.IDs, listResp.Objects[i].ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListScoresOptions(ctx context.Context, data ScoresDataSourceModel) (*client.ListScoresOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	startingAfter := strings.TrimSpace(data.StartingAfter.ValueString())
	endingBefore := strings.TrimSpace(data.EndingBefore.ValueString())
	hasStartingAfter := !data.StartingAfter.IsNull() && startingAfter != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && endingBefore != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Conflicting Attributes", "Cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	listOpts := &client.ListScoresOptions{}

	if !data.FilterIDs.IsNull() {
		var ids []string
		diags.Append(data.FilterIDs.ElementsAs(ctx, &ids, false)...)
		if diags.HasError() {
			return nil, diags
		}
		listOpts.IDs = ids
	}
	if !data.OrgName.IsNull() && strings.TrimSpace(data.OrgName.ValueString()) != "" {
		listOpts.OrgName = strings.TrimSpace(data.OrgName.ValueString())
	}
	if !data.ProjectID.IsNull() && strings.TrimSpace(data.ProjectID.ValueString()) != "" {
		listOpts.ProjectID = strings.TrimSpace(data.ProjectID.ValueString())
	}
	if !data.ProjectName.IsNull() && strings.TrimSpace(data.ProjectName.ValueString()) != "" {
		listOpts.ProjectName = strings.TrimSpace(data.ProjectName.ValueString())
	}
	if !data.ScoreName.IsNull() && strings.TrimSpace(data.ScoreName.ValueString()) != "" {
		listOpts.ScoreName = strings.TrimSpace(data.ScoreName.ValueString())
	}
	if !data.ScoreType.IsNull() && strings.TrimSpace(data.ScoreType.ValueString()) != "" {
		listOpts.ScoreType = strings.TrimSpace(data.ScoreType.ValueString())
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
		listOpts.StartingAfter = startingAfter
	}
	if hasEndingBefore {
		listOpts.EndingBefore = endingBefore
	}

	return listOpts, diags
}

func scoresDataSourceScoreFromScore(score *client.ProjectScore) (ScoresDataSourceScore, diag.Diagnostics) {
	var diags diag.Diagnostics

	scoreModel := ScoresDataSourceScore{
		ID:          stringOrNull(score.ID),
		ProjectID:   stringOrNull(score.ProjectID),
		UserID:      stringOrNull(score.UserID),
		Created:     stringOrNull(score.Created),
		Name:        stringOrNull(score.Name),
		Description: stringOrNull(score.Description),
		ScoreType:   stringOrNull(score.ScoreType),
	}

	categoriesValue, err := projectScoreJSONOrNull(score.Categories)
	if err != nil {
		diags.AddError(
			"Error Encoding categories",
			fmt.Sprintf("Unable to encode categories as JSON: %s", err),
		)
		return scoreModel, diags
	}
	scoreModel.Categories = categoriesValue

	configValue, err := projectScoreJSONOrNull(score.Config)
	if err != nil {
		diags.AddError(
			"Error Encoding config",
			fmt.Sprintf("Unable to encode config as JSON: %s", err),
		)
		return scoreModel, diags
	}
	scoreModel.Config = configValue

	if score.Position != nil {
		scoreModel.Position = stringOrNull(*score.Position)
	} else {
		scoreModel.Position = types.StringNull()
	}

	return scoreModel, diags
}
