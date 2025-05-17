package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	client *AwxClient
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an AWX project.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Project ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Project name.",
				Required:    true,
			},
			"organization": schema.Int32Attribute{
				Description: "Organization ID for the project to live in.",
				Required:    true,
			},
			"scm_type": schema.StringAttribute{
				Description: "Type of SCM resource. Options: `manual`, `git`, `svn` `insights`, `archive`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"manual", "git", "svn", "insights", "archive"}...),
				},
			},
			"description": schema.StringAttribute{
				Description: "Project description.",
				Optional:    true,
			},
			"allow_override": schema.BoolAttribute{
				Description: "Allow changing the Source Control branch or revision in a job template that uses this project.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"credential": schema.Int32Attribute{
				Description: "Source Control credential ID.",
				Optional:    true,
			},
			"default_environment": schema.Int32Attribute{
				Description: "The ID of the execution environment that will be used for jobs that use this project.",
				Optional:    true,
			},
			"local_path": schema.StringAttribute{
				Description: "Select from the list of directories found in the Project Base Path. Together the base path and the playbook directory provide the full path used to locate playbooks.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scm_branch": schema.StringAttribute{
				Description: "The branch name in source control.",
				Optional:    true,
			},
			"scm_clean": schema.BoolAttribute{
				Description: "Remove any local modifications prior to performing an update.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"scm_delete_on_update": schema.BoolAttribute{
				Description: "Delete the local repository in its entirety prior to performing an update. Depending on the size of the repository this may significantly increase the amount of time required to complete an update.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"scm_refspec": schema.StringAttribute{
				Description: "The refspec to use for the SCM resource.",
				Optional:    true,
			},
			"scm_track_submodules": schema.BoolAttribute{
				Description: "Track submodules latest commit on specified branch.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"scm_update_on_launch": schema.BoolAttribute{
				Description: "Perform an update to the local repository before launching a job with this project.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"scm_url": schema.StringAttribute{
				Description: "Example URLs for Remote Archive Source Control include: `https://github.com/username/project/archive/v0.0.1.tar.gz` `https://github.com/username/project/archive/v0.0.2.zip`",
				Optional:    true,
				//				Default:     stringdefault.StaticString("https://example.org"),
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeout": schema.Int32Attribute{
				Description: "The amount of time (in seconds) to run before the SCM Update is canceled. A value of 0 means no timeout.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(0),
			},
		},
	}
}

func (r ProjectResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Required attribute validation.

	// credential
	if data.ScmType.ValueString() == "insights" {
		if data.Credential.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("credential"),
				"Missing Attribute Configuration",
				"insights Source Control Type requires credential to be set",
			)
		}
	}

	// local_path
	if data.ScmType.ValueString() == "manual" {
		if data.LocalPath.ValueString() == "" || data.LocalPath.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("local_path"),
				"Missing Attribute Configuration",
				"manual Source Control Type requires local_path to be set",
			)
		}
	}

	// scm_url
	if data.ScmType.ValueString() == "git" || data.ScmType.ValueString() == "svn" || data.ScmType.ValueString() == "archive" {
		if data.ScmUrl.ValueString() == "" || data.ScmUrl.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_url"),
				"Missing Attribute Configuration",
				"[git, svn, archive] Source Control Types requires scm_url to be set",
			)
		}
	}

	// Not allowed validation.

	// allow_override
	if data.ScmType.ValueString() == "manual" || data.ScmType.ValueString() == "insights" {
		if data.AllowOverride.ValueBool() || !data.AllowOverride.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("allow_override"),
				"Attribute Configuration Error",
				"allow_override should not be set for [manual, insights] Source Control Types",
			)
		}
	}

	// credential
	if data.ScmType.ValueString() == "manual" {
		if !data.Credential.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("credential"),
				"Attribute Configuration Error",
				"credential should not be set for Manual Source Control Type",
			)
		}
	}

	// local_path
	if data.ScmType.ValueString() == "git" || data.ScmType.ValueString() == "svn" || data.ScmType.ValueString() == "insights" || data.ScmType.ValueString() == "archive" {
		if data.LocalPath.ValueString() != "" || !data.LocalPath.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("local_path"),
				"Attribute Configuration Error",
				"local_path should not be set for [git, svn, insights, archive] Source Control Types",
			)
		}
	}

	// scm_branch
	if data.ScmType.ValueString() == "manual" || data.ScmType.ValueString() == "insights" || data.ScmType.ValueString() == "archive" {
		if data.ScmBranch.ValueString() != "" || !data.ScmBranch.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_branch"),
				"Attribute Configuration Error",
				"scm_branch should not be set for [manual, insights, archive] Source Control Types",
			)
		}
	}

	// scm_clean
	if data.ScmType.ValueString() == "manual" {
		if data.ScmClean.ValueBool() || !data.ScmClean.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_clean"),
				"Attribute Configuration Error",
				"scm_clean should not be set for manual Source Control Type",
			)
		}
	}

	// scm_del_on_update
	if data.ScmType.ValueString() == "manual" {
		if data.ScmDelOnUpdate.ValueBool() || !data.ScmDelOnUpdate.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_del_on_update"),
				"Attribute Configuration Error",
				"scm_del_on_update should not be set for manual Source Control Type",
			)
		}
	}

	// scm_refspec
	if data.ScmType.ValueString() == "manual" || data.ScmType.ValueString() == "svn" || data.ScmType.ValueString() == "insights" || data.ScmType.ValueString() == "archive" {
		if data.ScmRefSpec.ValueString() != "" || !data.ScmRefSpec.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_refspec"),
				"Attribute Configuration Error",
				"scm_refspec should not be set for [manual, svn, insights, archive] Source Control Types",
			)
		}
	}

	// scm_track_submodules
	if data.ScmType.ValueString() == "manual" || data.ScmType.ValueString() == "svn" || data.ScmType.ValueString() == "insights" || data.ScmType.ValueString() == "archive" {
		if data.ScmTrackSubmodules.ValueBool() || !data.ScmTrackSubmodules.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_track_submodules"),
				"Attribute Configuration Error",
				"scm_track_submodules should not be set for [manual, svn, insights, archive] Source Control Types",
			)
		}
	}

	// scm_update_on_launch
	if data.ScmType.ValueString() == "manual" {
		if data.ScmUpdOnLaunch.ValueBool() || !data.ScmUpdOnLaunch.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_update_on_launch"),
				"Attribute Configuration Error",
				"scm_update_on_launch should not be set for manual Source Control Type",
			)
		}
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*AwxClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = configureData
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData ProjectAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.Organization = int(data.Organization.ValueInt32())

	// Convert manual to blank for API
	if data.ScmType.ValueString() == "manual" {
		bodyData.ScmType = ""
	} else {
		bodyData.ScmType = data.ScmType.ValueString()
	}

	bodyData.Timeout = int(data.Timeout.ValueInt32())

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.AllowOverride.IsNull()) {
		bodyData.AllowOverride = data.AllowOverride.ValueBool()
	}
	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}
	if !(data.DefaultEnv.IsNull()) {
		bodyData.DefaultEnv = int(data.DefaultEnv.ValueInt32())
	}
	if !(data.LocalPath.IsNull()) {
		bodyData.LocalPath = data.LocalPath.ValueString()
	}
	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !(data.ScmClean.IsNull()) {
		bodyData.ScmClean = data.ScmClean.ValueBool()
	}
	if !(data.ScmDelOnUpdate.IsNull()) {
		bodyData.ScmDelOnUpdate = data.ScmDelOnUpdate.ValueBool()
	}
	if !(data.ScmRefSpec.IsNull()) {
		bodyData.ScmRefSpec = data.ScmRefSpec.ValueString()
	}
	if !(data.ScmTrackSubmodules.IsNull()) {
		bodyData.ScmTrackSubmodules = data.ScmTrackSubmodules.ValueBool()
	}
	if !(data.ScmUpdOnLaunch.IsNull()) {
		bodyData.ScmUpdOnLaunch = data.ScmUpdOnLaunch.ValueBool()
	}
	if !(data.ScmUrl.IsNull()) {
		bodyData.ScmUrl = data.ScmUrl.ValueString()
	}

	url := "projects/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues := []string{"id", "local_path", "scm_url"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))
	data.LocalPath = types.StringValue(fmt.Sprintf("%v", returnedData["local_path"]))
	data.ScmUrl = types.StringValue(fmt.Sprintf("%v", returnedData["scm_url"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id))
		return
	}

	url := fmt.Sprintf("projects/%d/", id)
	body, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404})
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

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), responseData.Organization)...)

	if responseData.ScmType == "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_type"), "manual")...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_type"), responseData.ScmType)...)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allow_override"), responseData.AllowOverride)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_clean"), responseData.ScmClean)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_delete_on_update"), responseData.ScmDelOnUpdate)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_track_submodules"), responseData.ScmTrackSubmodules)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_update_on_launch"), responseData.ScmUpdOnLaunch)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeout"), responseData.Timeout)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Credential.IsNull() || responseData.Credential != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("credential"), responseData.Credential)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.DefaultEnv.IsNull() || responseData.DefaultEnv != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("default_environment"), responseData.DefaultEnv)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.LocalPath.IsNull() || responseData.LocalPath != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("local_path"), responseData.LocalPath)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.ScmBranch.IsNull() || responseData.ScmBranch != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_branch"), responseData.ScmBranch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.ScmRefSpec.IsNull() || responseData.ScmRefSpec != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_refspec"), responseData.ScmRefSpec)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.ScmUrl.IsNull() || responseData.ScmUrl != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_url"), responseData.ScmUrl)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data))
		return
	}

	var bodyData ProjectAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.Organization = int(data.Organization.ValueInt32())

	// Convert manual to blank for API
	if data.ScmType.ValueString() == "manual" {
		bodyData.ScmType = ""
	} else {
		bodyData.ScmType = data.ScmType.ValueString()
	}

	bodyData.Timeout = int(data.Timeout.ValueInt32())

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.AllowOverride.IsNull()) {
		bodyData.AllowOverride = data.AllowOverride.ValueBool()
	}
	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}
	if !(data.DefaultEnv.IsNull()) {
		bodyData.DefaultEnv = int(data.DefaultEnv.ValueInt32())
	}
	if !(data.LocalPath.IsNull()) {
		bodyData.LocalPath = data.LocalPath.ValueString()
	}
	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !(data.ScmClean.IsNull()) {
		bodyData.ScmClean = data.ScmClean.ValueBool()
	}
	if !(data.ScmDelOnUpdate.IsNull()) {
		bodyData.ScmDelOnUpdate = data.ScmDelOnUpdate.ValueBool()
	}
	if !(data.ScmRefSpec.IsNull()) {
		bodyData.ScmRefSpec = data.ScmRefSpec.ValueString()
	}
	if !(data.ScmTrackSubmodules.IsNull()) {
		bodyData.ScmTrackSubmodules = data.ScmTrackSubmodules.ValueBool()
	}
	if !(data.ScmUpdOnLaunch.IsNull()) {
		bodyData.ScmUpdOnLaunch = data.ScmUpdOnLaunch.ValueBool()
	}
	if !(data.ScmUrl.IsNull()) {
		bodyData.ScmUrl = data.ScmUrl.ValueString()
	}

	url := fmt.Sprintf("projects/%d/", id)
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues := []string{"local_path", "scm_url"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	data.LocalPath = types.StringValue(fmt.Sprintf("%v", returnedData["local_path"]))
	data.ScmUrl = types.StringValue(fmt.Sprintf("%v", returnedData["scm_url"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("projects/%d/", id)
	// 403 & 409 return code indicates project is being used by a job or had sync failures.
	// Newly create projects will be syncing and not able to be deleted immediately.
	attempts := 30
	var del_err error
	var statusCode int
	for attempt := 0; attempt < attempts; attempt++ {
		_, statusCode, del_err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204})
		if del_err == nil {
			if attempt > 0 {
				resp.Diagnostics.AddWarning(
					"Retry required to delete project",
					fmt.Sprintf("Project was successfully delete after %d attempt(s).", attempt))
			}
			return
		}
		if statusCode != 403 && statusCode != 409 {
			resp.Diagnostics.AddError(
				"Error making API delete request",
				fmt.Sprintf("Error was: %s.", del_err.Error()))
			return
		}
		time.Sleep(4 * time.Second)
	}

	if del_err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error after %v attempt(s) was: %s.", attempts, del_err.Error()))
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
