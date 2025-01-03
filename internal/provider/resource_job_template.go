package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &JobTemplateResource{}
var _ resource.ResourceWithImportState = &JobTemplateResource{}

func NewJobTemplateResource() resource.Resource {
	return &JobTemplateResource{}
}

// JobTemplateResource defines the resource implementation.
type JobTemplateResource struct {
	client *AwxClient
}

// JobTemplateResourceModel describes the resource data model.
type JobTemplateResourceModel struct {
	Id                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	Description                    types.String `tfsdk:"description"`
	JobType                        types.String `tfsdk:"job_type"`
	Inventory                      types.Int32  `tfsdk:"inventory"`
	Project                        types.Int32  `tfsdk:"project"`
	Playbook                       types.String `tfsdk:"playbook"`
	ScmBranch                      types.String `tfsdk:"scm_branch"`
	Forks                          types.Int32  `tfsdk:"forks"`
	Limit                          types.String `tfsdk:"limit"`
	Verbosity                      types.Int32  `tfsdk:"verbosity"`
	ExtraVars                      types.String `tfsdk:"extra_vars"`
	JobTags                        types.String `tfsdk:"job_tags"`
	ForceHandlers                  types.Bool   `tfsdk:"force_handlers"`
	SkipTags                       types.String `tfsdk:"skip_tags"`
	StartAtTask                    types.String `tfsdk:"start_at_tags"`
	Timeout                        types.Int32  `tfsdk:"timeout"`
	UseFactCache                   types.Bool   `tfsdk:"use_fact_cache"`
	ExecutionEnvironment           types.Int32  `tfsdk:"execution_environment"`
	HostConfigKey                  types.String `tfsdk:"host_config_key"`
	AskScmBranchOnLaunch           types.Bool   `tfsdk:"ask_scm_branch_on_launch"`
	AskDiffModeOnLaunch            types.Bool   `tfsdk:"ask_diff_mode_on_launch"`
	AskVariablesOnLaunch           types.Bool   `tfsdk:"ask_variables_on_launch"`
	AskLimitOnLaunch               types.Bool   `tfsdk:"ask_limit_on_launch"`
	AskTagsOnLaunch                types.Bool   `tfsdk:"ask_tags_on_launch"`
	AskSkipTagsOnLaunch            types.Bool   `tfsdk:"ask_skip_tags_on_launch"`
	AskJobTypeOnLaunch             types.Bool   `tfsdk:"ask_job_type_on_launch"`
	AskVerbosityOnLaunch           types.Bool   `tfsdk:"ask_verbosity_on_launch"`
	AskInventoryOnLaunch           types.Bool   `tfsdk:"ask_inventory_on_launch"`
	AskCredentialOnLaunch          types.Bool   `tfsdk:"ask_credential_on_launch"`
	AskExecutionEnvironmenOnLaunch types.Bool   `tfsdk:"ask_execution_environment_on_launch"`
	AskLablesOnLaunch              types.Bool   `tfsdk:"ask_labels_on_launch"`
	AskForksOnLaunch               types.Bool   `tfsdk:"ask_forks_on_launch"`
	AskJobSliceCountOnLaunch       types.Bool   `tfsdk:"ask_job_slice_count_on_launch"`
	AskTimeoutOnLaunch             types.Bool   `tfsdk:"ask_timeout_on_launch"`
	AskInstanceGroupsOnLaunch      types.Bool   `tfsdk:"ask_instance_groups_on_launch"`
	SurveyEnabled                  types.Bool   `tfsdk:"survey_enabled"`
	BecomeEnabled                  types.Bool   `tfsdk:"become_enabled"`
	DiffMode                       types.Bool   `tfsdk:"diff_mode"`
	AllowSimultaneous              types.Bool   `tfsdk:"allow_simultaneous"`
	CustomVirtualEnv               types.String `tfsdk:"custom_virtualenv"`
	JobSliceCount                  types.Int32  `tfsdk:"job_slice_count"`
	WebhookService                 types.String `tfsdk:"webhook_service"`
	WebhookCredential              types.String `tfsdk:"webhook_credential"`
	PreventInstanceGroupFallback   types.Bool   `tfsdk:"prevent_instance_group_fallback"`
}

type JobTemplate struct {
	Name                           string `json:"name"`
	Description                    string `json:"description"`
	JobType                        string `json:"job_type"`
	Inventory                      int    `json:"inventory"`
	Project                        int    `json:"project"`
	Playbook                       string `json:"playbook"`
	ScmBranch                      string `json:"scm_branch"`
	Forks                          int    `json:"forks"`
	Limit                          string `json:"limit"`
	Verbosity                      int    `json:"verbosity"`
	ExtraVars                      string `json:"extra_vars"`
	JobTags                        string `json:"job_tags"`
	ForceHandlers                  bool   `json:"force_handlers"`
	SkipTags                       string `json:"skip_tags"`
	StartAtTask                    string `json:"start_at_tags"`
	Timeout                        int    `json:"timeout"`
	UseFactCache                   bool   `json:"use_fact_cache"`
	ExecutionEnvironment           int    `json:"execution_environment"`
	HostConfigKey                  string `json:"host_config_key"`
	AskScmBranchOnLaunch           bool   `json:"ask_scm_branch_on_launch"`
	AskDiffModeOnLaunch            bool   `json:"ask_diff_mode_on_launch"`
	AskVariablesOnLaunch           bool   `json:"ask_variables_on_launch"`
	AskLimitOnLaunch               bool   `json:"ask_limit_on_launch"`
	AskTagsOnLaunch                bool   `json:"ask_tags_on_launch"`
	AskSkipTagsOnLaunch            bool   `json:"ask_skip_tags_on_launch"`
	AskJobTypeOnLaunch             bool   `json:"ask_job_type_on_launch"`
	AskVerbosityOnLaunch           bool   `json:"ask_verbosity_on_launch"`
	AskInventoryOnLaunch           bool   `json:"ask_inventory_on_launch"`
	AskCredentialOnLaunch          bool   `json:"ask_credential_on_launch"`
	AskExecutionEnvironmenOnLaunch bool   `json:"ask_execution_environment_on_launch"`
	AskLablesOnLaunch              bool   `json:"ask_labels_on_launch"`
	AskForksOnLaunch               bool   `json:"ask_forks_on_launch"`
	AskJobSliceCountOnLaunch       bool   `json:"ask_job_slice_count_on_launch"`
	AskTimeoutOnLaunch             bool   `json:"ask_timeout_on_launch"`
	AskInstanceGroupsOnLaunch      bool   `json:"ask_instance_groups_on_launch"`
	SurveyEnabled                  bool   `json:"survey_enabled"`
	BecomeEnabled                  bool   `json:"become_enabled"`
	DiffMode                       bool   `json:"diff_mode"`
	AllowSimultaneous              bool   `json:"allow_simultaneous"`
	CustomVirtualEnv               any    `json:"custom_virtualenv"` //blank is returned by api as "custom_virtual": null (not "")
	JobSliceCount                  int    `json:"job_slice_count"`
	WebhookService                 string `json:"webhook_service"`
	WebhookCredential              any    `json:"webhook_credential"` //blank is returned by api as "webhook_credentials": null (not "")
	PreventInstanceGroupFallback   bool   `json:"prevent_instance_group_fallback"`
}

func (r *JobTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_template"
}

func (r *JobTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Managed Job Templates",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"job_type": schema.StringAttribute{
				Required:    true,
				Description: "Acceptable values are a choice of: run, or check.",
			},
			"inventory": schema.Int32Attribute{
				Optional:    true,
				Description: "ID number of the inventory to associate with the job template. Supply this or set ask_inventory_on_launch = true.",
			},
			"project": schema.Int32Attribute{
				Required:    true,
				Description: "ID number of the project to associate with the job template",
			},
			"playbook": schema.StringAttribute{
				Required:    true,
				Description: "Playbook name to be executed by this job",
			},
			"scm_branch": schema.StringAttribute{
				Optional: true,
			},
			"forks": schema.Int32Attribute{
				Optional: true,
			},
			"limit": schema.StringAttribute{
				Optional: true,
			},
			"verbosity": schema.Int32Attribute{
				Optional: true,
			},
			"extra_vars": schema.StringAttribute{
				Optional: true,
			},
			"job_tags": schema.StringAttribute{
				Optional: true,
			},
			"force_handlers": schema.BoolAttribute{
				Optional: true,
			},
			"skip_tags": schema.StringAttribute{
				Optional: true,
			},
			"start_at_tags": schema.StringAttribute{
				Optional: true,
			},
			"timeout": schema.Int32Attribute{
				Optional: true,
			},
			"use_fact_cache": schema.BoolAttribute{
				Optional: true,
			},
			"execution_environment": schema.Int32Attribute{
				Optional: true,
			},
			"host_config_key": schema.StringAttribute{
				Optional: true,
			},
			"ask_scm_branch_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_diff_mode_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_variables_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_limit_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_tags_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_skip_tags_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_job_type_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_verbosity_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_inventory_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_credential_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_execution_environment_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_labels_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_forks_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_job_slice_count_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_timeout_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"ask_instance_groups_on_launch": schema.BoolAttribute{
				Optional: true,
			},
			"survey_enabled": schema.BoolAttribute{
				Optional: true,
			},
			"become_enabled": schema.BoolAttribute{
				Optional: true,
			},
			"diff_mode": schema.BoolAttribute{
				Optional: true,
			},
			"allow_simultaneous": schema.BoolAttribute{
				Optional: true,
			},
			"custom_virtualenv": schema.StringAttribute{
				Optional: true,
			},
			"job_slice_count": schema.Int32Attribute{
				Optional: true,
			},
			"webhook_service": schema.StringAttribute{
				Optional: true,
			},
			"webhook_credential": schema.StringAttribute{
				Optional: true,
			},
			"prevent_instance_group_fallback": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (r *JobTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobTemplateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := r.client.endpoint + "/api/v2/job_templates/"

	// get body data for HTTP request
	var bodyData JobTemplate
	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	bodyData.Description = data.Description.ValueString()

	if !(data.JobType.IsNull()) {
		bodyData.JobType = data.JobType.ValueString()
	}
	if !(data.Inventory.IsNull()) {
		bodyData.Inventory = int(data.Inventory.ValueInt32())
	}
	if !(data.Project.IsNull()) {
		bodyData.Project = int(data.Project.ValueInt32())
	}
	if !(data.Playbook.IsNull()) {
		bodyData.Playbook = data.Playbook.ValueString()
	}
	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !(data.Forks.IsNull()) {
		bodyData.Forks = int(data.Forks.ValueInt32())
	}
	if !(data.Limit.IsNull()) {
		bodyData.Limit = data.Limit.ValueString()
	}
	if !(data.Verbosity.IsNull()) {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}
	if !(data.ExtraVars.IsNull()) {
		bodyData.ExtraVars = data.ExtraVars.ValueString()
	}
	if !(data.JobTags.IsNull()) {
		bodyData.JobTags = data.JobTags.ValueString()
	}
	if !(data.ForceHandlers.IsNull()) {
		bodyData.ForceHandlers = data.ForceHandlers.ValueBool()
	}
	if !(data.SkipTags.IsNull()) {
		bodyData.SkipTags = data.SkipTags.ValueString()
	}
	if !(data.StartAtTask.IsNull()) {
		bodyData.StartAtTask = data.StartAtTask.ValueString()
	}
	if !(data.Timeout.IsNull()) {
		bodyData.Timeout = int(data.Timeout.ValueInt32())
	}
	if !(data.UseFactCache.IsNull()) {
		bodyData.UseFactCache = data.UseFactCache.ValueBool()
	}
	if !(data.UseFactCache.IsNull()) {
		bodyData.ExecutionEnvironment = int(data.ExecutionEnvironment.ValueInt32())
	}
	if !(data.HostConfigKey.IsNull()) {
		bodyData.HostConfigKey = data.HostConfigKey.ValueString()
	}
	if !(data.AskScmBranchOnLaunch.IsNull()) {
		bodyData.AskScmBranchOnLaunch = data.AskScmBranchOnLaunch.ValueBool()
	}
	if !(data.AskDiffModeOnLaunch.IsNull()) {
		bodyData.AskDiffModeOnLaunch = data.AskDiffModeOnLaunch.ValueBool()
	}
	if !(data.AskVariablesOnLaunch.IsNull()) {
		bodyData.AskVariablesOnLaunch = data.AskVariablesOnLaunch.ValueBool()
	}
	if !(data.AskLimitOnLaunch.IsNull()) {
		bodyData.AskLimitOnLaunch = data.AskLimitOnLaunch.ValueBool()
	}
	if !(data.AskTagsOnLaunch.IsNull()) {
		bodyData.AskTagsOnLaunch = data.AskTagsOnLaunch.ValueBool()
	}
	if !(data.AskSkipTagsOnLaunch.IsNull()) {
		bodyData.AskSkipTagsOnLaunch = data.AskSkipTagsOnLaunch.ValueBool()
	}
	if !(data.AskJobTypeOnLaunch.IsNull()) {
		bodyData.AskJobTypeOnLaunch = data.AskJobTypeOnLaunch.ValueBool()
	}
	if !(data.AskVerbosityOnLaunch.IsNull()) {
		bodyData.AskVerbosityOnLaunch = data.AskVerbosityOnLaunch.ValueBool()
	}
	if !(data.AskInventoryOnLaunch.IsNull()) {
		bodyData.AskInventoryOnLaunch = data.AskInventoryOnLaunch.ValueBool()
	}
	if !(data.AskCredentialOnLaunch.IsNull()) {
		bodyData.AskCredentialOnLaunch = data.AskCredentialOnLaunch.ValueBool()
	}
	if !(data.AskExecutionEnvironmenOnLaunch.IsNull()) {
		bodyData.AskExecutionEnvironmenOnLaunch = data.AskExecutionEnvironmenOnLaunch.ValueBool()
	}
	if !(data.AskLablesOnLaunch.IsNull()) {
		bodyData.AskLablesOnLaunch = data.AskLablesOnLaunch.ValueBool()
	}
	if !(data.AskForksOnLaunch.IsNull()) {
		bodyData.AskForksOnLaunch = data.AskForksOnLaunch.ValueBool()
	}
	if !(data.AskJobSliceCountOnLaunch.IsNull()) {
		bodyData.AskJobSliceCountOnLaunch = data.AskJobSliceCountOnLaunch.ValueBool()
	}
	if !(data.AskTimeoutOnLaunch.IsNull()) {
		bodyData.AskTimeoutOnLaunch = data.AskTimeoutOnLaunch.ValueBool()
	}
	if !(data.AskInstanceGroupsOnLaunch.IsNull()) {
		bodyData.AskInstanceGroupsOnLaunch = data.AskInstanceGroupsOnLaunch.ValueBool()
	}
	if !(data.SurveyEnabled.IsNull()) {
		bodyData.SurveyEnabled = data.SurveyEnabled.ValueBool()
	}
	if !(data.BecomeEnabled.IsNull()) {
		bodyData.BecomeEnabled = data.BecomeEnabled.ValueBool()
	}
	if !(data.DiffMode.IsNull()) {
		bodyData.DiffMode = data.DiffMode.ValueBool()
	}
	if !(data.AllowSimultaneous.IsNull()) {
		bodyData.AllowSimultaneous = data.AllowSimultaneous.ValueBool()
	}
	if !(data.CustomVirtualEnv.IsNull()) {
		bodyData.CustomVirtualEnv = data.CustomVirtualEnv.ValueString()
	}
	if !(data.JobSliceCount.IsNull()) {
		bodyData.JobSliceCount = int(data.JobSliceCount.ValueInt32())
	}
	if !(data.WebhookService.IsNull()) {
		bodyData.WebhookService = data.WebhookService.ValueString()
	}
	if !(data.WebhookCredential.IsNull()) {
		bodyData.WebhookCredential = data.WebhookCredential.ValueString()
	}
	if !(data.PreventInstanceGroupFallback.IsNull()) {
		bodyData.PreventInstanceGroupFallback = data.PreventInstanceGroupFallback.ValueBool()
	}

	jsonData, err := json.Marshal(bodyData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable marshal json",
			fmt.Sprintf("Unable to convert id: %+v. ", bodyData))
		return
	}

	// create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(jsonData)))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to generate request",
			fmt.Sprintf("Unable to gen url: %v. ", url))
		return
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", "Bearer"+" "+r.client.token)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}
	if httpResp.StatusCode != 201 {
		defer httpResp.Body.Close()
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable read http request response body.",
				err.Error())
			return
		}

		resp.Diagnostics.AddError(
			"Bad request status code.",
			fmt.Sprintf("Expected 201, got %v with message %s. ", httpResp.StatusCode, body))
		return
	}

	tmp := struct {
		Id int `json:"id"`
	}{}

	defer httpResp.Body.Close()
	httpRepsBodyData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get http response body",
			fmt.Sprintf("Error was %v", err))
		return
	}
	err = json.Unmarshal(httpRepsBodyData, &tmp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get unmarshall http response to grab ID",
			fmt.Sprintf("error was %v", err))
		return
	}

	idAsString := strconv.Itoa(tmp.Id)

	data.Id = types.StringValue(idAsString)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobTemplateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// set url for create HTTP request
	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}
	url := r.client.endpoint + fmt.Sprintf("/api/v2/job_templates/%d/", id)

	// create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to generate request",
			fmt.Sprintf("Unable to gen url: %v. ", url))
		return
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", "Bearer"+" "+r.client.token)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}
	if httpResp.StatusCode != 200 && httpResp.StatusCode != 404 {
		resp.Diagnostics.AddError(
			"Bad request status code.",
			fmt.Sprintf("Expected 200, got %v. ", httpResp.StatusCode))
		return

	}

	if httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	var responseData JobTemplate

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Uanble to get all data out of the http response data body",
			fmt.Sprintf("Body got %v. ", body))
		return
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Uanble unmarshall response body into object",
			fmt.Sprintf("Error =  %v. ", err.Error()))
		return
	}

	//responseID := fmt.Sprint(responseData.Id)
	//data.Id = types.StringValue(responseID)
	//data.Name = types.StringValue(responseData.Name)
	if !(data.Name.IsNull() && responseData.Name == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	//data.Description = types.StringValue(responseData.Description)
	if !(data.Description.IsNull() && responseData.Description == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.JobType = types.StringValue(responseData.JobType)
	if !(data.JobType.IsNull() && responseData.JobType == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_type"), responseData.JobType)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	//data.Inventory = types.Int32Value(int32(responseData.Inventory))
	if !(data.Inventory.IsNull() && responseData.Inventory == 0) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inventory"), responseData.Inventory)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.Project = types.Int32Value(int32(responseData.Project))
	if !(data.Project.IsNull() && responseData.Project == 0) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project"), responseData.Project)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.Playbook = types.StringValue(responseData.Playbook)
	if !(data.Playbook.IsNull() && responseData.Playbook == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("playbook"), responseData.Playbook)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.ScmBranch = types.StringValue(responseData.ScmBranch)
	if !(data.ScmBranch.IsNull() && responseData.ScmBranch == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_branch"), responseData.ScmBranch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.Forks = types.Int32Value(int32(responseData.Forks))
	if !(data.Forks.IsNull() && responseData.Forks == 0) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("forks"), responseData.Forks)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.Limit = types.StringValue(responseData.Limit)
	if !(data.Limit.IsNull() && responseData.Limit == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("limit"), responseData.Limit)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.Verbosity = types.Int32Value(int32(responseData.Verbosity))
	if !(data.Verbosity.IsNull() && responseData.Verbosity == 0) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("verbosity"), responseData.Verbosity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.ExtraVars = types.StringValue(responseData.ExtraVars)
	if !(data.ExtraVars.IsNull() && responseData.ExtraVars == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("extra_vars"), responseData.ExtraVars)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.JobTags = types.StringValue(responseData.JobTags)
	if !(data.JobTags.IsNull() && responseData.JobTags == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_tags"), responseData.JobTags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.ForceHandlers = types.BoolValue(responseData.ForceHandlers)
	if !(data.ForceHandlers.IsNull() && data.ForceHandlers.ValueBool() == responseData.ForceHandlers) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("force_handlers"), responseData.ForceHandlers)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.SkipTags = types.StringValue(responseData.SkipTags)
	if !(data.SkipTags.IsNull() && responseData.SkipTags == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("skip_tags"), responseData.SkipTags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.StartAtTask = types.StringValue(responseData.StartAtTask)
	if !(data.StartAtTask.IsNull() && responseData.StartAtTask == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("start_at_tags"), responseData.StartAtTask)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.Timeout = types.Int32Value(int32(responseData.Timeout))
	if !(data.Timeout.IsNull() && responseData.Timeout == 0) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeout"), responseData.Timeout)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.UseFactCache = types.BoolValue(responseData.UseFactCache)
	if !(data.UseFactCache.IsNull() && data.UseFactCache.ValueBool() == responseData.UseFactCache) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("use_fact_cache"), responseData.UseFactCache)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.ExecutionEnvironment = types.Int32Value(int32(responseData.ExecutionEnvironment))
	if !(data.ExecutionEnvironment.IsNull() && responseData.ExecutionEnvironment == 0) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("execution_environment"), responseData.ExecutionEnvironment)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.HostConfigKey = types.StringValue(responseData.HostConfigKey)
	if !(data.HostConfigKey.IsNull() && responseData.HostConfigKey == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host_config_key"), responseData.HostConfigKey)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskScmBranchOnLaunch = types.BoolValue(responseData.AskScmBranchOnLaunch)
	if !(data.AskScmBranchOnLaunch.IsNull() && data.AskScmBranchOnLaunch.ValueBool() == responseData.AskScmBranchOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_scm_branch_on_launch"), responseData.AskScmBranchOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskDiffModeOnLaunch = types.BoolValue(responseData.AskDiffModeOnLaunch)
	if !(data.AskDiffModeOnLaunch.IsNull() && data.AskDiffModeOnLaunch.ValueBool() == responseData.AskDiffModeOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_diff_mode_on_launch"), responseData.AskDiffModeOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskVariablesOnLaunch = types.BoolValue(responseData.AskVariablesOnLaunch)
	if !(data.AskVariablesOnLaunch.IsNull() && data.AskVariablesOnLaunch.ValueBool() == responseData.AskVariablesOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_variables_on_launch"), responseData.AskVariablesOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.AskLimitOnLaunch = types.BoolValue(responseData.AskLimitOnLaunch)
	if !(data.AskLimitOnLaunch.IsNull() && data.AskLimitOnLaunch.ValueBool() == responseData.AskLimitOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_limit_on_launch"), responseData.AskLimitOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskTagsOnLaunch = types.BoolValue(responseData.AskTagsOnLaunch)
	if !(data.AskTagsOnLaunch.IsNull() && data.AskTagsOnLaunch.ValueBool() == responseData.AskTagsOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_tags_on_launch"), responseData.AskTagsOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskSkipTagsOnLaunch = types.BoolValue(responseData.AskSkipTagsOnLaunch)
	if !(data.AskSkipTagsOnLaunch.IsNull() && data.AskSkipTagsOnLaunch.ValueBool() == responseData.AskSkipTagsOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_skip_tags_on_launch"), responseData.AskSkipTagsOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskJobTypeOnLaunch = types.BoolValue(responseData.AskJobTypeOnLaunch)
	if !(data.AskJobTypeOnLaunch.IsNull() && data.AskJobTypeOnLaunch.ValueBool() == responseData.AskJobTypeOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_job_type_on_launch"), responseData.AskJobTypeOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskVerbosityOnLaunch = types.BoolValue(responseData.AskVerbosityOnLaunch)
	if !(data.AskVerbosityOnLaunch.IsNull() && data.AskVerbosityOnLaunch.ValueBool() == responseData.AskVerbosityOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_verbosity_on_launch"), responseData.AskVerbosityOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskInventoryOnLaunch = types.BoolValue(responseData.AskInventoryOnLaunch)
	if !(data.AskInventoryOnLaunch.IsNull() && data.AskInventoryOnLaunch.ValueBool() == responseData.AskInventoryOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_inventory_on_launch"), responseData.AskInventoryOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskCredentialOnLaunch = types.BoolValue(responseData.AskCredentialOnLaunch)
	if !(data.AskCredentialOnLaunch.IsNull() && data.AskCredentialOnLaunch.ValueBool() == responseData.AskCredentialOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_credential_on_launch"), responseData.AskCredentialOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskExecutionEnvironmenOnLaunch = types.BoolValue(responseData.AskExecutionEnvironmenOnLaunch)
	if !(data.AskExecutionEnvironmenOnLaunch.IsNull() && data.AskExecutionEnvironmenOnLaunch.ValueBool() == responseData.AskExecutionEnvironmenOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_execution_environment_on_launch"), responseData.AskExecutionEnvironmenOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskLablesOnLaunch = types.BoolValue(responseData.AskLablesOnLaunch)
	if !(data.AskLablesOnLaunch.IsNull() && data.AskLablesOnLaunch.ValueBool() == responseData.AskLablesOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_labels_on_launch"), responseData.AskLablesOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskForksOnLaunch = types.BoolValue(responseData.AskForksOnLaunch)
	if !(data.AskForksOnLaunch.IsNull() && data.AskForksOnLaunch.ValueBool() == responseData.AskForksOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_forks_on_launch"), responseData.AskForksOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskJobSliceCountOnLaunch = types.BoolValue(responseData.AskJobSliceCountOnLaunch)
	if !(data.AskJobSliceCountOnLaunch.IsNull() && data.AskJobSliceCountOnLaunch.ValueBool() == responseData.AskJobSliceCountOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_job_slice_count_on_launch"), responseData.AskJobSliceCountOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AskTimeoutOnLaunch = types.BoolValue(responseData.AskTimeoutOnLaunch)
	if !(data.AskTimeoutOnLaunch.IsNull() && data.AskTimeoutOnLaunch.ValueBool() == responseData.AskTimeoutOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_timeout_on_launch"), responseData.AskTimeoutOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.AskInstanceGroupsOnLaunch = types.BoolValue(responseData.AskInstanceGroupsOnLaunch)
	if !(data.AskInstanceGroupsOnLaunch.IsNull() && data.AskInstanceGroupsOnLaunch.ValueBool() == responseData.AskInstanceGroupsOnLaunch) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_instance_groups_on_launch"), responseData.AskInstanceGroupsOnLaunch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.SurveyEnabled = types.BoolValue(responseData.SurveyEnabled)
	if !(data.SurveyEnabled.IsNull() && data.SurveyEnabled.ValueBool() == responseData.SurveyEnabled) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("survey_enabled"), responseData.SurveyEnabled)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.BecomeEnabled = types.BoolValue(responseData.BecomeEnabled)
	if !(data.BecomeEnabled.IsNull() && data.BecomeEnabled.ValueBool() == responseData.BecomeEnabled) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("become_enabled"), responseData.BecomeEnabled)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.DiffMode = types.BoolValue(responseData.DiffMode)
	if !(data.DiffMode.IsNull() && data.DiffMode.ValueBool() == responseData.DiffMode) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("diff_mode"), responseData.DiffMode)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.AllowSimultaneous = types.BoolValue(responseData.AllowSimultaneous)
	if !(data.AllowSimultaneous.IsNull() && data.AllowSimultaneous.ValueBool() == responseData.AllowSimultaneous) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allow_simultaneous"), responseData.AllowSimultaneous)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// data.CustomVirtualEnv = types.StringValue(responseData.CustomVirtualEnv)
	if !(data.CustomVirtualEnv.IsNull() && responseData.CustomVirtualEnv == nil) {
		if responseData.CustomVirtualEnv == nil {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("custom_virtualenv"), "")...)
		} else {
			if customVirtualEnv, ok := responseData.CustomVirtualEnv.(string); ok {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("custom_virtualenv"), customVirtualEnv)...)
			} else {
				resp.Diagnostics.AddError(
					"Invalid Type",
					"Expected responseData.CustomVirtualEnv to be a string",
				)
				return
			}
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}
	// data.JobSliceCount = types.Int32Value(int32(responseData.JobSliceCount))
	if !(data.JobSliceCount.IsNull() && responseData.JobSliceCount == 1) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_slice_count"), responseData.JobSliceCount)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	//data.WebhookService = types.StringValue(responseData.WebhookService)
	if !(data.WebhookService.IsNull() && responseData.WebhookService == "") {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_service"), responseData.WebhookService)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//data.WebhookCredential = types.StringValue(responseData.WebhookCredential.(string))
	if !(data.WebhookCredential.IsNull() && responseData.WebhookCredential == nil) {
		if responseData.WebhookCredential == nil {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_credential"), "")...)
		} else {
			if webhookCredential, ok := responseData.WebhookCredential.(string); ok {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_credential"), webhookCredential)...)
			} else {
				resp.Diagnostics.AddError(
					"Invalid Type",
					"Expected responseData.WebhookCredential to be a string",
				)
				return
			}
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}
	//data.PreventInstanceGroupFallback = types.BoolValue(responseData.PreventInstanceGroupFallback)
	if !(data.PreventInstanceGroupFallback.IsNull() && data.PreventInstanceGroupFallback.ValueBool() == responseData.PreventInstanceGroupFallback) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prevent_instance_group_fallback"), responseData.PreventInstanceGroupFallback)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Save updated data into Terraform state
	//resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JobTemplateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData JobTemplate
	bodyData.Name = data.Name.ValueString()
	bodyData.Description = data.Description.ValueString()
	bodyData.JobType = data.JobType.ValueString()
	bodyData.Inventory = int(data.Inventory.ValueInt32())
	bodyData.Project = int(data.Project.ValueInt32())
	bodyData.Playbook = data.Playbook.ValueString()
	bodyData.ScmBranch = data.ScmBranch.ValueString()
	bodyData.Forks = int(data.Forks.ValueInt32())
	bodyData.Limit = data.Limit.ValueString()
	bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	bodyData.ExtraVars = data.ExtraVars.ValueString()
	bodyData.JobTags = data.JobTags.ValueString()
	bodyData.ForceHandlers = data.ForceHandlers.ValueBool()
	bodyData.SkipTags = data.SkipTags.ValueString()
	bodyData.StartAtTask = data.StartAtTask.ValueString()
	bodyData.Timeout = int(data.Timeout.ValueInt32())
	bodyData.UseFactCache = data.UseFactCache.ValueBool()
	bodyData.ExecutionEnvironment = int(data.ExecutionEnvironment.ValueInt32())
	bodyData.HostConfigKey = data.HostConfigKey.ValueString()
	bodyData.AskScmBranchOnLaunch = data.AskScmBranchOnLaunch.ValueBool()
	bodyData.AskDiffModeOnLaunch = data.AskDiffModeOnLaunch.ValueBool()
	bodyData.AskVariablesOnLaunch = data.AskVariablesOnLaunch.ValueBool()
	bodyData.AskLimitOnLaunch = data.AskLimitOnLaunch.ValueBool()
	bodyData.AskTagsOnLaunch = data.AskTagsOnLaunch.ValueBool()
	bodyData.AskSkipTagsOnLaunch = data.AskSkipTagsOnLaunch.ValueBool()
	bodyData.AskJobTypeOnLaunch = data.AskJobTypeOnLaunch.ValueBool()
	bodyData.AskVerbosityOnLaunch = data.AskVerbosityOnLaunch.ValueBool()
	bodyData.AskInventoryOnLaunch = data.AskInventoryOnLaunch.ValueBool()
	bodyData.AskCredentialOnLaunch = data.AskCredentialOnLaunch.ValueBool()
	bodyData.AskExecutionEnvironmenOnLaunch = data.AskExecutionEnvironmenOnLaunch.ValueBool()
	bodyData.AskLablesOnLaunch = data.AskLablesOnLaunch.ValueBool()
	bodyData.AskForksOnLaunch = data.AskForksOnLaunch.ValueBool()
	bodyData.AskJobSliceCountOnLaunch = data.AskJobSliceCountOnLaunch.ValueBool()
	bodyData.AskTimeoutOnLaunch = data.AskTimeoutOnLaunch.ValueBool()
	bodyData.AskInstanceGroupsOnLaunch = data.AskInstanceGroupsOnLaunch.ValueBool()
	bodyData.SurveyEnabled = data.SurveyEnabled.ValueBool()
	bodyData.BecomeEnabled = data.BecomeEnabled.ValueBool()
	bodyData.DiffMode = data.DiffMode.ValueBool()
	bodyData.AllowSimultaneous = data.AllowSimultaneous.ValueBool()
	bodyData.CustomVirtualEnv = data.CustomVirtualEnv.ValueString()
	bodyData.JobSliceCount = int(data.JobSliceCount.ValueInt32())
	bodyData.WebhookService = data.WebhookService.ValueString()
	bodyData.WebhookCredential = data.WebhookCredential.ValueString()
	bodyData.PreventInstanceGroupFallback = data.PreventInstanceGroupFallback.ValueBool()

	jsonData, err := json.Marshal(bodyData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable marshal json",
			fmt.Sprintf("Unable to convert id: %+v. ", bodyData))
		return
	}

	// set url for create HTTP request
	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}
	url := r.client.endpoint + fmt.Sprintf("/api/v2/job_templates/%d/", id)

	// create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(jsonData)))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to generate request",
			fmt.Sprintf("Unable to gen url: %v. ", url))
		return
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", "Bearer"+" "+r.client.token)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}
	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError(
			"Bad request status code.",
			fmt.Sprintf("Expected 200, got %v. ", httpResp.StatusCode))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data JobTemplateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// set url for create HTTP request
	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}
	url := r.client.endpoint + fmt.Sprintf("/api/v2/job_templates/%d/", id)

	// create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to generate delete request",
			fmt.Sprintf("Unable to gen url: %v. ", url))
		return
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", "Bearer"+" "+r.client.token)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete got error: %s", err))
		return
	}
	if httpResp.StatusCode != 204 {
		resp.Diagnostics.AddError(
			"Bad request status code.",
			fmt.Sprintf("Expected 204, got %v. ", httpResp.StatusCode))
		return
	}
}

func (r *JobTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
