package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ViewsDataSource{}

// NewViewsDataSource creates a new views data source instance.
func NewViewsDataSource() datasource.DataSource {
	return &ViewsDataSource{}
}

// ViewsDataSource defines the data source implementation.
type ViewsDataSource struct {
	client *client.Client
}

// ViewsDataSourceModel describes the data source data model.
type ViewsDataSourceModel struct {
	FilterIDs     types.List            `tfsdk:"filter_ids"`
	ObjectID      types.String          `tfsdk:"object_id"`
	ObjectType    types.String          `tfsdk:"object_type"`
	ViewName      types.String          `tfsdk:"view_name"`
	ViewType      types.String          `tfsdk:"view_type"`
	StartingAfter types.String          `tfsdk:"starting_after"`
	EndingBefore  types.String          `tfsdk:"ending_before"`
	Views         []ViewsDataSourceView `tfsdk:"views"`
	IDs           []string              `tfsdk:"ids"`
	Limit         types.Int64           `tfsdk:"limit"`
}

// ViewsDataSourceView represents a single view in the list.
type ViewsDataSourceView struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ObjectID   types.String `tfsdk:"object_id"`
	ObjectType types.String `tfsdk:"object_type"`
	ViewType   types.String `tfsdk:"view_type"`
	UserID     types.String `tfsdk:"user_id"`
	Created    types.String `tfsdk:"created"`
	DeletedAt  types.String `tfsdk:"deleted_at"`
}

// Metadata implements datasource.DataSource.
func (d *ViewsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_views"
}

// Schema implements datasource.DataSource.
func (d *ViewsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Braintrust views using API-native filters.",
		Attributes: map[string]schema.Attribute{
			"filter_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional list of view IDs to filter by. Maps to repeated `ids` query parameters.",
			},
			"object_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object ID to scope the view listing.",
			},
			"object_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object type to scope the view listing.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"organization",
						"project",
						"experiment",
						"dataset",
						"prompt",
						"prompt_session",
						"group",
						"role",
						"org_member",
						"project_log",
						"org_project",
					),
				},
			},
			"view_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional exact view name filter.",
			},
			"view_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional view type filter.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"projects",
						"experiments",
						"experiment",
						"playgrounds",
						"playground",
						"datasets",
						"dataset",
						"prompts",
						"tools",
						"scorers",
						"logs",
					),
				},
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional max number of views to return.",
			},
			"starting_after": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch views after this ID.",
			},
			"ending_before": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional pagination cursor to fetch views before this ID.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of returned view IDs.",
			},
			"views": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of views.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the view.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the view.",
						},
						"object_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The object ID the view applies to.",
						},
						"object_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The object type the view applies to.",
						},
						"view_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The table type this view corresponds to.",
						},
						"user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the user who created the view.",
						},
						"created": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the view was created.",
						},
						"deleted_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The timestamp when the view was deleted, if deleted.",
						},
					},
				},
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *ViewsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ViewsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ViewsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listOpts, filterDiags := buildListViewsOptions(ctx, data)
	resp.Diagnostics.Append(filterDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.client.ListViews(ctx, listOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Views",
			fmt.Sprintf("Could not list views: %s", err.Error()),
		)
		return
	}

	data.Views = make([]ViewsDataSourceView, 0, len(listResp.Objects))
	data.IDs = make([]string, 0, len(listResp.Objects))

	for i := range listResp.Objects {
		view := &listResp.Objects[i]
		if view.DeletedAt != "" {
			continue
		}

		viewModel := viewsDataSourceViewFromView(view)
		data.Views = append(data.Views, viewModel)
		data.IDs = append(data.IDs, view.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildListViewsOptions(ctx context.Context, data ViewsDataSourceModel) (*client.ListViewsOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	hasStartingAfter := !data.StartingAfter.IsNull() && data.StartingAfter.ValueString() != ""
	hasEndingBefore := !data.EndingBefore.IsNull() && data.EndingBefore.ValueString() != ""

	if hasStartingAfter && hasEndingBefore {
		diags.AddError("Conflicting Attributes", "Cannot specify both 'starting_after' and 'ending_before'.")
		return nil, diags
	}

	objectID := strings.TrimSpace(data.ObjectID.ValueString())
	if objectID == "" {
		diags.AddError("Invalid object_id", "'object_id' must be provided and non-empty.")
		return nil, diags
	}

	objectTypeValue := strings.TrimSpace(data.ObjectType.ValueString())
	if objectTypeValue == "" {
		diags.AddError("Invalid object_type", "'object_type' must be provided and non-empty.")
		return nil, diags
	}

	listOpts := &client.ListViewsOptions{
		ObjectID:   objectID,
		ObjectType: client.ACLObjectType(objectTypeValue),
	}

	if !data.FilterIDs.IsNull() {
		var ids []string
		diags.Append(data.FilterIDs.ElementsAs(ctx, &ids, false)...)
		if diags.HasError() {
			return nil, diags
		}
		listOpts.IDs = ids
	}
	if !data.ViewName.IsNull() && data.ViewName.ValueString() != "" {
		listOpts.ViewName = data.ViewName.ValueString()
	}
	if !data.ViewType.IsNull() && data.ViewType.ValueString() != "" {
		listOpts.ViewType = client.ViewType(data.ViewType.ValueString())
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

func viewsDataSourceViewFromView(view *client.View) ViewsDataSourceView {
	return ViewsDataSourceView{
		ID:         stringOrNull(view.ID),
		Name:       stringOrNull(view.Name),
		ObjectID:   stringOrNull(view.ObjectID),
		ObjectType: stringOrNull(string(view.ObjectType)),
		ViewType:   stringOrNull(string(view.ViewType)),
		UserID:     stringOrNull(view.UserID),
		Created:    stringOrNull(view.Created),
		DeletedAt:  stringOrNull(view.DeletedAt),
	}
}
