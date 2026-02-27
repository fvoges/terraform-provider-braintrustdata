package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &TagsDataSource{}

// NewTagsDataSource creates a new tags data source instance.
func NewTagsDataSource() datasource.DataSource {
	return &TagsDataSource{}
}

// TagsDataSource defines the data source implementation.
type TagsDataSource struct {
	client *client.Client
}

// TagsDataSourceModel describes the data source data model.
type TagsDataSourceModel struct {
	FilterIDs     types.List          `tfsdk:"filter_ids"`
	OrgName       types.String        `tfsdk:"org_name"`
	ProjectID     types.String        `tfsdk:"project_id"`
	ProjectName   types.String        `tfsdk:"project_name"`
	TagName       types.String        `tfsdk:"tag_name"`
	StartingAfter types.String        `tfsdk:"starting_after"`
	EndingBefore  types.String        `tfsdk:"ending_before"`
	Tags          []TagsDataSourceTag `tfsdk:"tags"`
	IDs           []string            `tfsdk:"ids"`
	Limit         types.Int64         `tfsdk:"limit"`
}

// TagsDataSourceTag represents a single tag in the list.
type TagsDataSourceTag struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	UserID      types.String `tfsdk:"user_id"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
}

// Metadata implements datasource.DataSource.
func (d *TagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

// Schema implements datasource.DataSource.
func (d *TagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust project tags using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"filter_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional list of tag IDs to filter by. Maps to repeated `ids` query parameters.",
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
			"tag_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact tag name filter.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of tags to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch tags after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch tags before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned tag IDs.",
			},
			"tags": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of tags.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the tag.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the tag.",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project ID that the tag belongs to.",
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
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *TagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TagsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListTagsOptions(ctx, data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListTags(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Tags",
			fmt.Sprintf("Could not list tags: %s", err.Error()),
		)
		return
	}

	data.Tags = make([]TagsDataSourceTag, 0, len(listResp.Tags))
	data.IDs = make([]string, 0, len(listResp.Tags))

	for i := range listResp.Tags {
		tag := &listResp.Tags[i]
		tagModel := TagsDataSourceTag{
			ID:        types.StringValue(tag.ID),
			Name:      types.StringValue(tag.Name),
			ProjectID: types.StringValue(tag.ProjectID),
		}

		if tag.UserID != "" {
			tagModel.UserID = types.StringValue(tag.UserID)
		} else {
			tagModel.UserID = types.StringNull()
		}
		if tag.Color != "" {
			tagModel.Color = types.StringValue(tag.Color)
		} else {
			tagModel.Color = types.StringNull()
		}
		if tag.Description != "" {
			tagModel.Description = types.StringValue(tag.Description)
		} else {
			tagModel.Description = types.StringNull()
		}
		if tag.Created != "" {
			tagModel.Created = types.StringValue(tag.Created)
		} else {
			tagModel.Created = types.StringNull()
		}

		data.Tags = append(data.Tags, tagModel)
		data.IDs = append(data.IDs, tag.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListTagsOptions(ctx context.Context, data TagsDataSourceModel) (*client.ListTagsOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Invalid Filters", "cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	listOpts := &client.ListTagsOptions{}

	if !data.FilterIDs.IsNull() {
		var ids []string
		diags.Append(data.FilterIDs.ElementsAs(ctx, &ids, false)...)
		if diags.HasError() {
			return nil, diags
		}
		listOpts.IDs = ids
	}
	if !data.OrgName.IsNull() && data.OrgName.ValueString() != "" {
		listOpts.OrgName = data.OrgName.ValueString()
	}
	if !data.ProjectID.IsNull() && data.ProjectID.ValueString() != "" {
		listOpts.ProjectID = data.ProjectID.ValueString()
	}
	if !data.ProjectName.IsNull() && data.ProjectName.ValueString() != "" {
		listOpts.ProjectName = data.ProjectName.ValueString()
	}
	if !data.TagName.IsNull() && data.TagName.ValueString() != "" {
		listOpts.TagName = data.TagName.ValueString()
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
