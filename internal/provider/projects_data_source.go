package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectsDataSource{}

// NewProjectsDataSource creates a new projects data source instance.
func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// ProjectsDataSource defines the data source implementation.
type ProjectsDataSource struct {
	client *client.Client
}

// ProjectsDataSourceModel describes the data source data model.
type ProjectsDataSourceModel struct {
	OrgName       types.String                `tfsdk:"org_name"`
	ProjectName   types.String                `tfsdk:"project_name"`
	StartingAfter types.String                `tfsdk:"starting_after"`
	EndingBefore  types.String                `tfsdk:"ending_before"`
	Projects      []ProjectsDataSourceProject `tfsdk:"projects"`
	IDs           []string                    `tfsdk:"ids"`
	Limit         types.Int64                 `tfsdk:"limit"`
}

// ProjectsDataSourceProject represents a single project in the list.
type ProjectsDataSourceProject struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgID       types.String `tfsdk:"org_id"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	UserID      types.String `tfsdk:"user_id"`
}

// Metadata implements datasource.DataSource.
func (d *ProjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

// Schema implements datasource.DataSource.
func (d *ProjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust projects using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter.",
			},
			"project_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact project name filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of projects to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch projects after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch projects before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned project IDs.",
			},
			"projects": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of projects.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the project.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the project.",
						},
						"org_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The organization ID the project belongs to.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of the project.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the project was created.",
						},
						"user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the user who created the project.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *ProjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""
	if hasStartingAfter && hasEndingBefore {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot specify both 'starting_after' and 'ending_before'.",
		)
		return
	}

	listOpts := &client.ListProjectsOptions{}
	if !data.OrgName.IsNull() && data.OrgName.ValueString() != "" {
		listOpts.OrgName = data.OrgName.ValueString()
	}
	if !data.ProjectName.IsNull() && data.ProjectName.ValueString() != "" {
		listOpts.ProjectName = data.ProjectName.ValueString()
	}
	if !data.Limit.IsNull() {
		limit := data.Limit.ValueInt64()
		if limit < 1 {
			resp.Diagnostics.AddError(
				"Invalid Limit",
				"'limit' must be greater than or equal to 1.",
			)
			return
		}

		maxInt := int64(^uint(0) >> 1)
		if limit > maxInt {
			resp.Diagnostics.AddError(
				"Invalid Limit",
				"'limit' exceeds supported platform integer size.",
			)
			return
		}

		listOpts.Limit = int(limit)
	}
	if hasStartingAfter {
		listOpts.StartingAfter = data.StartingAfter.ValueString()
	}
	if hasEndingBefore {
		listOpts.EndingBefore = data.EndingBefore.ValueString()
	}

	listResp, err := d.client.ListProjects(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Projects",
			fmt.Sprintf("Could not list projects: %s", err.Error()),
		)
		return
	}

	data.Projects = make([]ProjectsDataSourceProject, 0, len(listResp.Projects))
	data.IDs = make([]string, 0, len(listResp.Projects))

	for _, project := range listResp.Projects {
		if project.DeletedAt != "" {
			continue
		}

		projectModel := ProjectsDataSourceProject{
			ID:   types.StringValue(project.ID),
			Name: types.StringValue(project.Name),
		}
		if project.OrgID != "" {
			projectModel.OrgID = types.StringValue(project.OrgID)
		} else {
			projectModel.OrgID = types.StringNull()
		}
		if project.Description != "" {
			projectModel.Description = types.StringValue(project.Description)
		} else {
			projectModel.Description = types.StringNull()
		}
		if project.Created != "" {
			projectModel.Created = types.StringValue(project.Created)
		} else {
			projectModel.Created = types.StringNull()
		}
		if project.UserID != "" {
			projectModel.UserID = types.StringValue(project.UserID)
		} else {
			projectModel.UserID = types.StringNull()
		}

		data.Projects = append(data.Projects, projectModel)
		data.IDs = append(data.IDs, project.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
