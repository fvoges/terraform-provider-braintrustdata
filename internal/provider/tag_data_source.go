package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &TagDataSource{}

var (
	errTagNotFoundByName       = errors.New("tag not found by name")
	errMultipleTagsFoundByName = errors.New("multiple tags found by name")
)

// NewTagDataSource creates a new tag data source instance.
func NewTagDataSource() datasource.DataSource {
	return &TagDataSource{}
}

// TagDataSource defines the data source implementation.
type TagDataSource struct {
	client *client.Client
}

// TagDataSourceModel describes the data source data model.
type TagDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
	OrgName     types.String `tfsdk:"org_name"`
	UserID      types.String `tfsdk:"user_id"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *TagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

// Schema implements datasource.DataSource.
func (d *TagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust project tag by `id` or by API-native searchable attributes (`name`, optionally `project_id`, `project_name`, `org_name`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the tag. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The tag name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional project ID filter applied during searchable lookups.",
			},
			"project_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional project name filter applied during searchable lookups.",
			},
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional organization name filter applied during searchable lookups.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the tag.",
			},
			"color": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Color of the tag for the UI.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Textual description of the tag.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the tag was created.",
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *TagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""
	hasProjectID := !data.ProjectID.IsNull() && data.ProjectID.ValueString() != ""
	hasProjectName := !data.ProjectName.IsNull() && data.ProjectName.ValueString() != ""
	hasOrgName := !data.OrgName.IsNull() && data.OrgName.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the tag.",
		)
		return
	}

	if hasID && (hasName || hasProjectID || hasProjectName || hasOrgName) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'project_id', 'project_name', 'org_name').",
		)
		return
	}

	var tag *client.Tag
	if hasID {
		fetchedTag, err := d.client.GetTag(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Tag",
				fmt.Sprintf("Could not read tag ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
		tag = fetchedTag
	} else {
		listOpts := &client.ListTagsOptions{
			TagName: data.Name.ValueString(),
			Limit:   2,
		}
		if hasProjectID {
			listOpts.ProjectID = data.ProjectID.ValueString()
		}
		if hasProjectName {
			listOpts.ProjectName = data.ProjectName.ValueString()
		}
		if hasOrgName {
			listOpts.OrgName = data.OrgName.ValueString()
		}

		listResp, err := d.client.ListTags(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Tags",
				fmt.Sprintf("Could not list tags using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedTag, err := selectSingleTagByName(listResp.Tags, data.Name.ValueString())
		if errors.Is(err, errTagNotFoundByName) {
			resp.Diagnostics.AddError(
				"Tag Not Found",
				fmt.Sprintf("No tag found with name: %s", data.Name.ValueString()),
			)
			return
		}
		if errors.Is(err, errMultipleTagsFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple Tags Found",
				"Searchable attributes matched multiple tags. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Tags",
				fmt.Sprintf("Could not resolve tag using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		tag = selectedTag
	}

	populateTagDataSourceModel(ctx, &data, tag)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func selectSingleTagByName(tags []client.Tag, tagName string) (*client.Tag, error) {
	var selected *client.Tag

	for i := range tags {
		tag := &tags[i]
		if tag.Name != tagName {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleTagsFoundByName, tagName)
		}
		selected = tag
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errTagNotFoundByName, tagName)
	}

	return selected, nil
}

func populateTagDataSourceModel(_ context.Context, data *TagDataSourceModel, tag *client.Tag) {
	data.ID = types.StringValue(tag.ID)
	data.Name = types.StringValue(tag.Name)
	data.ProjectID = types.StringValue(tag.ProjectID)

	if tag.UserID != "" {
		data.UserID = types.StringValue(tag.UserID)
	} else {
		data.UserID = types.StringNull()
	}
	if tag.Color != "" {
		data.Color = types.StringValue(tag.Color)
	} else {
		data.Color = types.StringNull()
	}
	if tag.Description != "" {
		data.Description = types.StringValue(tag.Description)
	} else {
		data.Description = types.StringNull()
	}
	if tag.Created != "" {
		data.Created = types.StringValue(tag.Created)
	} else {
		data.Created = types.StringNull()
	}
}
