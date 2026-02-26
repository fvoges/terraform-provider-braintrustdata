package provider

import (
	"context"
	"fmt"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ExperimentResource{}
var _ resource.ResourceWithImportState = &ExperimentResource{}

// NewExperimentResource creates a new experiment resource instance.
func NewExperimentResource() resource.Resource {
	return &ExperimentResource{}
}

// ExperimentResource defines the resource implementation.
type ExperimentResource struct {
	client *client.Client
}

// ExperimentResourceModel describes the resource data model.
type ExperimentResourceModel struct {
	Tags        types.Set    `tfsdk:"tags"`
	Metadata    types.Map    `tfsdk:"metadata"`
	RepoInfo    types.Object `tfsdk:"repo_info"`
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	UserID      types.String `tfsdk:"user_id"`
	OrgID       types.String `tfsdk:"org_id"`
	Public      types.Bool   `tfsdk:"public"`
}

var experimentRepoInfoAttributeTypes = map[string]attr.Type{
	"commit":         types.StringType,
	"branch":         types.StringType,
	"tag":            types.StringType,
	"dirty":          types.BoolType,
	"author_name":    types.StringType,
	"author_email":   types.StringType,
	"commit_message": types.StringType,
	"commit_time":    types.StringType,
	"git_diff":       types.StringType,
}

type experimentRepoInfoModel struct {
	Commit        types.String `tfsdk:"commit"`
	Branch        types.String `tfsdk:"branch"`
	Tag           types.String `tfsdk:"tag"`
	AuthorName    types.String `tfsdk:"author_name"`
	AuthorEmail   types.String `tfsdk:"author_email"`
	CommitMessage types.String `tfsdk:"commit_message"`
	CommitTime    types.String `tfsdk:"commit_time"`
	GitDiff       types.String `tfsdk:"git_diff"`
	Dirty         types.Bool   `tfsdk:"dirty"`
}

type repoInfoValueState int

const (
	repoInfoValueStateKnown repoInfoValueState = iota
	repoInfoValueStateNull
	repoInfoValueStateUnknown
)

// Metadata implements resource.Resource.
func (r *ExperimentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_experiment"
}

// Schema implements resource.Resource.
func (r *ExperimentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Braintrust experiment. Experiments are collections of runs that test different prompts, models, or configurations.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the experiment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project this experiment belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the experiment.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the experiment.",
			},
			"public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the experiment is publicly accessible. Defaults to false.",
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Metadata associated with the experiment as key-value pairs.",
			},
			"repo_info": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Git repository metadata snapshot associated with the experiment.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"commit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "SHA of the most recent commit.",
					},
					"branch": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Branch name of the most recent commit.",
					},
					"tag": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Tag name on the most recent commit.",
					},
					"dirty": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether the repository had uncommitted changes.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"author_name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Author name of the most recent commit.",
					},
					"author_email": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Author email of the most recent commit.",
					},
					"commit_message": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Most recent commit message.",
					},
					"commit_time": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Time of the most recent commit.",
					},
					"git_diff": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Diff against the most recent commit when the repo was dirty.",
					},
				},
			},
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Tags associated with the experiment.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the experiment was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user who created the experiment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the organization this experiment belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure implements resource.Resource.
func (r *ExperimentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create implements resource.Resource by creating a new experiment.
func (r *ExperimentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExperimentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert metadata from Terraform Map to Go map
	var metadata map[string]interface{}
	if !data.Metadata.IsNull() {
		metadata = make(map[string]interface{})
		metadataMap := make(map[string]string)
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadataMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range metadataMap {
			metadata[k] = v
		}
	}

	// Convert tags from Terraform Set to Go slice
	var tags []string
	if !data.Tags.IsNull() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Get public value
	var publicPtr *bool
	if !data.Public.IsNull() {
		publicVal := data.Public.ValueBool()
		publicPtr = &publicVal
	}

	repoInfo, repoInfoState, repoInfoDiags := objectToRepoInfoWithState(ctx, data.RepoInfo)
	resp.Diagnostics.Append(repoInfoDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repoInfoPtr *client.RepoInfo
	if repoInfoState == repoInfoValueStateKnown {
		repoInfoPtr = repoInfo
	}

	// Create experiment via API
	experiment, err := r.client.CreateExperiment(ctx, &client.CreateExperimentRequest{
		ProjectID:   data.ProjectID.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Public:      publicPtr,
		Metadata:    metadata,
		Tags:        tags,
		RepoInfo:    repoInfoPtr,
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create experiment, got error: %s", err))
		return
	}

	// Read the experiment to get the complete state
	experiment, err = r.client.GetExperiment(ctx, experiment.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read experiment after creation, got error: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(experiment.ID)
	data.ProjectID = types.StringValue(experiment.ProjectID)
	data.Name = types.StringValue(experiment.Name)
	if experiment.Description != "" {
		data.Description = types.StringValue(experiment.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.Created = types.StringValue(experiment.Created)
	if experiment.UserID != "" {
		data.UserID = types.StringValue(experiment.UserID)
	} else {
		data.UserID = types.StringNull()
	}
	data.OrgID = types.StringValue(experiment.OrgID)
	data.Public = types.BoolValue(experiment.Public)

	repoInfoValue, repoInfoValueDiags := repoInfoToObject(experiment.RepoInfo)
	resp.Diagnostics.Append(repoInfoValueDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RepoInfo = repoInfoValue

	// Convert metadata from Go map to Terraform Map
	if len(experiment.Metadata) > 0 {
		metadataStrings := make(map[string]string)
		for k, v := range experiment.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// Convert tags from Go slice to Terraform Set
	if len(experiment.Tags) > 0 {
		tagsSet, diags := types.SetValueFrom(ctx, types.StringType, experiment.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tagsSet
	} else {
		data.Tags = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource by reading an experiment.
func (r *ExperimentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ExperimentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get experiment from API
	experiment, err := r.client.GetExperiment(ctx, data.ID.ValueString())

	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read experiment, got error: %s", err))
		return
	}

	// Check for soft delete
	if experiment.DeletedAt != "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model with response data
	data.Name = types.StringValue(experiment.Name)
	if experiment.Description != "" {
		data.Description = types.StringValue(experiment.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.ProjectID = types.StringValue(experiment.ProjectID)
	data.Created = types.StringValue(experiment.Created)
	if experiment.UserID != "" {
		data.UserID = types.StringValue(experiment.UserID)
	} else {
		data.UserID = types.StringNull()
	}
	data.OrgID = types.StringValue(experiment.OrgID)
	data.Public = types.BoolValue(experiment.Public)

	repoInfoValue, repoInfoValueDiags := repoInfoToObject(experiment.RepoInfo)
	resp.Diagnostics.Append(repoInfoValueDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RepoInfo = repoInfoValue

	// Convert metadata from Go map to Terraform Map
	if len(experiment.Metadata) > 0 {
		metadataStrings := make(map[string]string)
		for k, v := range experiment.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// Convert tags from Go slice to Terraform Set
	if len(experiment.Tags) > 0 {
		tagsSet, diags := types.SetValueFrom(ctx, types.StringType, experiment.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tagsSet
	} else {
		data.Tags = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource by updating an existing experiment.
func (r *ExperimentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExperimentResourceModel
	var state ExperimentResourceModel

	// Get current state to preserve fields not returned by update API
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields are not unknown or null
	if data.ID.IsNull() || data.ID.IsUnknown() || data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Plan",
			"Cannot update experiment because id is unknown or empty.",
		)
		return
	}
	if data.Name.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid Plan",
			"Cannot update experiment because name is unknown.",
		)
		return
	}

	updateReq, repoInfoState, reqDiags := buildExperimentUpdateRequestWithState(ctx, data)
	resp.Diagnostics.Append(reqDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update experiment via API
	experiment, err := r.client.UpdateExperiment(ctx, data.ID.ValueString(), updateReq)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update experiment, got error: %s", err))
		return
	}

	// Update model with response data
	data.Name = types.StringValue(experiment.Name)
	if experiment.Description != "" {
		data.Description = types.StringValue(experiment.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.Public = types.BoolValue(experiment.Public)

	repoInfoValue, repoInfoValueDiags := repoInfoToObject(experiment.RepoInfo)
	resp.Diagnostics.Append(repoInfoValueDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// If repo_info was omitted from update payload and API omitted it in response,
	// keep prior state to avoid implicit clears.
	if repoInfoState != repoInfoValueStateKnown && experiment.RepoInfo == nil {
		data.RepoInfo = state.RepoInfo
	} else {
		data.RepoInfo = repoInfoValue
	}

	// Convert metadata from Go map to Terraform Map
	if len(experiment.Metadata) > 0 {
		metadataStrings := make(map[string]string)
		for k, v := range experiment.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, metadataStrings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Metadata = metadataValue
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	// Convert tags from Go slice to Terraform Set
	if len(experiment.Tags) > 0 {
		tagsSet, diags := types.SetValueFrom(ctx, types.StringType, experiment.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tagsSet
	} else {
		data.Tags = types.SetNull(types.StringType)
	}

	// Preserve fields from state that aren't returned by update API
	data.Created = state.Created
	data.ProjectID = state.ProjectID
	data.ID = state.ID
	data.UserID = state.UserID
	data.OrgID = state.OrgID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildExperimentUpdateRequest(ctx context.Context, data ExperimentResourceModel) (*client.UpdateExperimentRequest, diag.Diagnostics) {
	req, _, diags := buildExperimentUpdateRequestWithState(ctx, data)
	return req, diags
}

func buildExperimentUpdateRequestWithState(ctx context.Context, data ExperimentResourceModel) (*client.UpdateExperimentRequest, repoInfoValueState, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.UpdateExperimentRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	// Convert metadata from Terraform Map to Go map.
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metadata := make(map[string]interface{})
		metadataMap := make(map[string]string)
		diags.Append(data.Metadata.ElementsAs(ctx, &metadataMap, false)...)
		if diags.HasError() {
			return req, repoInfoValueStateKnown, diags
		}
		for k, v := range metadataMap {
			metadata[k] = v
		}
		req.Metadata = metadata
	} else if data.Metadata.IsNull() {
		// Explicitly clear metadata by sending empty map.
		req.Metadata = make(map[string]interface{})
	}

	// Convert tags from Terraform Set to Go slice.
	if !data.Tags.IsNull() {
		var tags []string
		diags.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return req, repoInfoValueStateKnown, diags
		}
		req.Tags = tags
	}

	if !data.Public.IsNull() {
		publicVal := data.Public.ValueBool()
		req.Public = &publicVal
	}

	repoInfo, repoInfoState, repoInfoDiags := objectToRepoInfoWithState(ctx, data.RepoInfo)
	diags.Append(repoInfoDiags...)
	if diags.HasError() {
		return req, repoInfoState, diags
	}
	if repoInfoState == repoInfoValueStateKnown {
		req.RepoInfo = repoInfo
	}

	return req, repoInfoState, diags
}

func objectToRepoInfoWithState(ctx context.Context, value types.Object) (*client.RepoInfo, repoInfoValueState, diag.Diagnostics) {
	if value.IsNull() {
		return nil, repoInfoValueStateNull, nil
	}
	if value.IsUnknown() {
		return nil, repoInfoValueStateUnknown, nil
	}

	var model experimentRepoInfoModel
	diags := value.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, repoInfoValueStateKnown, diags
	}

	return &client.RepoInfo{
		Commit:        stringPointerFromValue(model.Commit),
		Branch:        stringPointerFromValue(model.Branch),
		Tag:           stringPointerFromValue(model.Tag),
		Dirty:         boolPointerFromValue(model.Dirty),
		AuthorName:    stringPointerFromValue(model.AuthorName),
		AuthorEmail:   stringPointerFromValue(model.AuthorEmail),
		CommitMessage: stringPointerFromValue(model.CommitMessage),
		CommitTime:    stringPointerFromValue(model.CommitTime),
		GitDiff:       stringPointerFromValue(model.GitDiff),
	}, repoInfoValueStateKnown, nil
}

func repoInfoToObject(repoInfo *client.RepoInfo) (types.Object, diag.Diagnostics) {
	if repoInfo == nil {
		return types.ObjectNull(experimentRepoInfoAttributeTypes), nil
	}

	return types.ObjectValue(
		experimentRepoInfoAttributeTypes,
		map[string]attr.Value{
			"commit":         types.StringPointerValue(repoInfo.Commit),
			"branch":         types.StringPointerValue(repoInfo.Branch),
			"tag":            types.StringPointerValue(repoInfo.Tag),
			"dirty":          types.BoolPointerValue(repoInfo.Dirty),
			"author_name":    types.StringPointerValue(repoInfo.AuthorName),
			"author_email":   types.StringPointerValue(repoInfo.AuthorEmail),
			"commit_message": types.StringPointerValue(repoInfo.CommitMessage),
			"commit_time":    types.StringPointerValue(repoInfo.CommitTime),
			"git_diff":       types.StringPointerValue(repoInfo.GitDiff),
		},
	)
}

func stringPointerFromValue(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func boolPointerFromValue(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

// Delete implements resource.Resource by deleting an experiment.
func (r *ExperimentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ExperimentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete experiment via API
	err := r.client.DeleteExperiment(ctx, data.ID.ValueString())

	if err != nil {
		// Treat 404 as success (already deleted) for idempotency
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete experiment, got error: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState by importing an experiment by ID.
func (r *ExperimentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
