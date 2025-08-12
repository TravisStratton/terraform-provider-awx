package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlParser "net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

type ProjectDataSource struct {
	client *providerClient
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get project datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "A valid project ID. You must specify either the `id` or `name` field, but not both. ",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Project name. You must specify either the `id` or `name` field, but not both.",
				Optional:    true,
			},
			"organization": schema.Int32Attribute{
				Description: "Organization ID for the project to live in.",
				Computed:    true,
			},
			"scm_type": schema.StringAttribute{
				Description: "Type of SCM resource. Options: `manual`, `git`, `svn` `insights`, `archive`.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Project description.",
				Computed:    true,
			},
			"allow_override": schema.BoolAttribute{
				Description: "Allow changing the Source Control branch or revision in a job template that uses this project.",
				Computed:    true,
			},
			"credential": schema.Int32Attribute{
				Description: "Source Control credential ID.",
				Computed:    true,
			},
			"default_environment": schema.Int32Attribute{
				Description: "The ID of the execution environment that will be used for jobs that use this project.",
				Computed:    true,
			},
			"local_path": schema.StringAttribute{
				Description: "Select from the list of directories found in the Project Base Path. Together the base path and the playbook directory provide the full path used to locate playbooks.",
				Computed:    true,
			},
			"scm_branch": schema.StringAttribute{
				Description: "The branch name in source control.",
				Computed:    true,
			},
			"scm_clean": schema.BoolAttribute{
				Description: "Remove any local modifications prior to performing an update.",
				Computed:    true,
			},
			"scm_delete_on_update": schema.BoolAttribute{
				Description: "Delete the local repository in its entirety prior to performing an update. Depending on the size of the repository this may significantly increase the amount of time required to complete an update.",
				Computed:    true,
			},
			"scm_refspec": schema.StringAttribute{
				Description: "The refspec to use for the SCM resource.",
				Computed:    true,
			},
			"scm_track_submodules": schema.BoolAttribute{
				Description: "Track submodules latest commit on specified branch.",
				Computed:    true,
			},
			"scm_update_on_launch": schema.BoolAttribute{
				Description: "Perform an update to the local repository before launching a job with this project.",
				Computed:    true,
			},
			"scm_url": schema.StringAttribute{
				Description: "Example URLs for Remote Archive Source Control include: `https://github.com/username/project/archive/v0.0.1.tar.gz` `https://github.com/username/project/archive/v0.0.2.zip`",
				Computed:    true,
			},
			"timeout": schema.Int32Attribute{
				Description: "The amount of time (in seconds) to run before the SCM Update is canceled. A value of 0 means no timeout.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*providerClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = configureData
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	if !data.Id.IsNull() {
		id, err := strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable convert id from string to int.",
				fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
			return
		}
		url = fmt.Sprintf("projects/?id=%d", id)
	}
	if !data.Name.IsNull() {
		url = fmt.Sprintf("projects/?name=%s", urlParser.QueryEscape(data.Name.ValueString()))
	}

	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	if statusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	var responseData ProjectAPIModel

	countResult := struct {
		Count   int               `json:"count"`
		Results []ProjectAPIModel `json:"results"`
	}{}

	err = json.Unmarshal(body, &countResult)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into object",
			fmt.Sprintf("Error:  %v.", err.Error()))
		return
	}
	if countResult.Count == 1 {
		responseData = countResult.Results[0]
	} else {
		resp.Diagnostics.AddError(
			"Incorrect number of projects returned",
			fmt.Sprintf("Unable to read project as API returned %v projects.", countResult.Count))
		return
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Name = types.StringValue(responseData.Name)
	data.Organization = types.Int32Value(int32(responseData.Organization))
	data.ScmType = types.StringValue(responseData.ScmType)
	data.Timeout = types.Int32Value(int32(responseData.Timeout))

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.AllowOverride {
		data.AllowOverride = types.BoolValue(responseData.AllowOverride)
	}

	if responseData.Credential != 0 {
		data.Credential = types.Int32Value(int32(responseData.Credential))
	}

	if responseData.DefaultEnv != 0 {
		data.DefaultEnv = types.Int32Value(int32(responseData.DefaultEnv))
	}

	if responseData.LocalPath != "" {
		data.LocalPath = types.StringValue(responseData.LocalPath)
	}

	if responseData.ScmBranch != "" {
		data.ScmBranch = types.StringValue(responseData.ScmBranch)
	}

	if responseData.ScmClean {
		data.ScmClean = types.BoolValue(responseData.ScmClean)
	}

	if responseData.ScmDelOnUpdate {
		data.ScmDelOnUpdate = types.BoolValue(responseData.ScmDelOnUpdate)
	}

	if responseData.ScmRefSpec != "" {
		data.ScmRefSpec = types.StringValue(responseData.ScmRefSpec)
	}

	if responseData.ScmTrackSubmodules {
		data.ScmTrackSubmodules = types.BoolValue(responseData.ScmTrackSubmodules)
	}

	if responseData.ScmUpdOnLaunch {
		data.ScmUpdOnLaunch = types.BoolValue(responseData.ScmUpdOnLaunch)
	}

	if responseData.ScmUrl != "" {
		data.ScmUrl = types.StringValue(responseData.ScmUrl)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
