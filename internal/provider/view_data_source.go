package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ViewDataSource{}

var (
	errViewNotFoundByName       = errors.New("view not found by name")
	errMultipleViewsFoundByName = errors.New("multiple views found by name")
)

// NewViewDataSource creates a new view data source instance.
func NewViewDataSource() datasource.DataSource {
	return &ViewDataSource{}
}

// ViewDataSource defines the data source implementation.
type ViewDataSource struct {
	client *client.Client
}

// ViewDataSourceModel describes the data source data model.
type ViewDataSourceModel struct {
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
func (d *ViewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_view"
}

// Schema implements datasource.DataSource.
func (d *ViewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Braintrust view by `id` or API-native searchable attributes (`name` with `object_id` + `object_type`, optionally `view_type`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the view. Specify either `id` or `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The view name. Can be used as a searchable attribute when `id` is not provided.",
			},
			"object_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object ID that scopes the view lookup.",
			},
			"object_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The object type that scopes the view lookup.",
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
			"view_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional view type filter for name-based lookup.",
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
	}
}

// Configure implements datasource.DataSource.
func (d *ViewDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ViewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ViewDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectID := strings.TrimSpace(data.ObjectID.ValueString())
	objectTypeValue := strings.TrimSpace(data.ObjectType.ValueString())
	if objectID == "" {
		resp.Diagnostics.AddError(
			"Invalid object_id",
			"'object_id' must be provided and non-empty.",
		)
		return
	}
	if objectTypeValue == "" {
		resp.Diagnostics.AddError(
			"Invalid object_type",
			"'object_type' must be provided and non-empty.",
		)
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""
	hasViewType := !data.ViewType.IsNull() && data.ViewType.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Must specify either 'id' or 'name' to look up the view.",
		)
		return
	}

	if hasID && (hasName || hasViewType) {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Cannot combine 'id' with searchable attributes ('name', 'view_type').",
		)
		return
	}

	var view *client.View
	if hasID {
		fetchedView, err := d.client.GetView(ctx, data.ID.ValueString(), &client.GetViewOptions{
			ObjectID:   objectID,
			ObjectType: client.ACLObjectType(objectTypeValue),
		})
		if err != nil {
			if client.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"View Not Found",
					fmt.Sprintf("No view found with ID: %s", data.ID.ValueString()),
				)
				return
			}

			resp.Diagnostics.AddError(
				"Error Reading View",
				fmt.Sprintf("Could not read view ID %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
		view = fetchedView
	} else {
		listOpts := &client.ListViewsOptions{
			ObjectID:   objectID,
			ObjectType: client.ACLObjectType(objectTypeValue),
			ViewName:   data.Name.ValueString(),
			Limit:      2,
		}
		if hasViewType {
			listOpts.ViewType = client.ViewType(data.ViewType.ValueString())
		}

		listResp, err := d.client.ListViews(ctx, listOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Views",
				fmt.Sprintf("Could not list views using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		selectedView, err := selectSingleViewByName(listResp.Objects, data.Name.ValueString())
		if errors.Is(err, errViewNotFoundByName) {
			resp.Diagnostics.AddError(
				"View Not Found",
				fmt.Sprintf("No view found with name: %s", data.Name.ValueString()),
			)
			return
		}
		if errors.Is(err, errMultipleViewsFoundByName) {
			resp.Diagnostics.AddError(
				"Multiple Views Found",
				"Searchable attributes matched multiple views. Refine the query or use 'id' for deterministic lookup.",
			)
			return
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Listing Views",
				fmt.Sprintf("Could not resolve view using the provided searchable attributes: %s", err.Error()),
			)
			return
		}

		view = selectedView
	}

	if view.DeletedAt != "" {
		resp.Diagnostics.AddError(
			"View Deleted",
			fmt.Sprintf("View %s has been deleted", view.ID),
		)
		return
	}

	populateViewDataSourceModel(ctx, &data, view)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func selectSingleViewByName(views []client.View, viewName string) (*client.View, error) {
	var selected *client.View

	for i := range views {
		view := &views[i]
		if view.Name != viewName {
			continue
		}
		if view.DeletedAt != "" {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("%w: %s", errMultipleViewsFoundByName, viewName)
		}
		selected = view
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %s", errViewNotFoundByName, viewName)
	}

	return selected, nil
}

func populateViewDataSourceModel(_ context.Context, data *ViewDataSourceModel, view *client.View) {
	data.ID = stringOrNull(view.ID)
	data.Name = stringOrNull(view.Name)
	data.ObjectID = stringOrNull(view.ObjectID)
	data.ObjectType = stringOrNull(string(view.ObjectType))
	data.ViewType = stringOrNull(string(view.ViewType))
	data.UserID = stringOrNull(view.UserID)
	data.Created = stringOrNull(view.Created)
	data.DeletedAt = stringOrNull(view.DeletedAt)
}
