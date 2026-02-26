package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

// NewProjectDataSource creates a new project data source instance.
func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource defines the data source implementation.
type ProjectDataSource struct {
	client *client.Client
}

// ProjectDataSourceModel describes the data source data model.
type ProjectDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgName     types.String `tfsdk:"org_name"`
	OrgID       types.String `tfsdk:"org_id"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	UserID      types.String `tfsdk:"user_id"`
}

// Metadata implements datasource.DataSource.
func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema implements datasource.DataSource.
func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust project by `id` or by API-native searchable attributes (`name`, optionally `org_name`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the project. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter applied during searchable lookups.",
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID that the project belongs to.",
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
	}
}

// Configure implements datasource.DataSource.
func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""
	hasOrgName := !data.OrgName.IsNull() && data.OrgName.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the project.",
		)
		return
	}

	if hasID && (hasName || hasOrgName) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'org_name').",
		)
		return
	}

	var project *client.Project
	if hasID {
		fetchedProject, err := d.client.GetProject(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Project",
				fmt.Sprintf("Could not read project ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
		project = fetchedProject
	} else {
		listOpts := &client.ListProjectsOptions{
			ProjectName: data.Name.ValueString(),
			Limit:       2,
		}
		if hasOrgName {
			listOpts.OrgName = data.OrgName.ValueString()
		}

		listResp, err := d.client.ListProjects(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Projects",
				fmt.Sprintf("Could not list projects using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		if len(listResp.Projects) == 0 {
			resp.Diagnostics.AddError(
				"Project Not Found",
				fmt.Sprintf("No project found with name: %s", data.Name.ValueString()),
			)
			return
		}

		if len(listResp.Projects) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Projects Found",
				"Searchable attributes matched multiple projects. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}

		project = &listResp.Projects[0]
	}

	if project.DeletedAt != "" {
		resp.Diagnostics.AddError(
			"Project Not Found",
			"The requested project has been deleted.",
		)
		return
	}

	data.ID = types.StringValue(project.ID)
	data.Name = types.StringValue(project.Name)

	if project.OrgID != "" {
		data.OrgID = types.StringValue(project.OrgID)
	} else {
		data.OrgID = types.StringNull()
	}
	if project.Description != "" {
		data.Description = types.StringValue(project.Description)
	} else {
		data.Description = types.StringNull()
	}
	if project.Created != "" {
		data.Created = types.StringValue(project.Created)
	} else {
		data.Created = types.StringNull()
	}
	if project.UserID != "" {
		data.UserID = types.StringValue(project.UserID)
	} else {
		data.UserID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
