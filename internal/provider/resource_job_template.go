package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &JobTemplateResource{}
var _ resource.ResourceWithImportState = &JobTemplateResource{}

func NewJobTemplateResource() resource.Resource {
	return &JobTemplateResource{}
}

type JobTemplateResource struct {
	client *providerClient
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
				Description: "Job template id.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Job template name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Job template description.",
			},
			"job_type": schema.StringAttribute{
				Optional:    true,
				Description: "Acceptable values are a choice of: `run`, `check`. For job templates, select run to execute the playbook. Select check to only check playbook syntax, test environment setup, and report problems without executing the playbook.",
				Default:     stringdefault.StaticString("run"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"run", "check"}...),
				},
				Computed: true,
			},
			"inventory": schema.Int32Attribute{
				Optional:    true,
				Description: "ID number of the inventory to associate with the job template. Supply this or set `ask_inventory_on_launch = true`.",
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
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Branch to use in job run. Project default used if blank. Only allowed if project allow_override field is set to true.",
			},
			"forks": schema.Int32Attribute{
				Optional:    true,
				Default:     int32default.StaticInt32(0),
				Computed:    true,
				Description: "The number of parallel or simultaneous processes to use while executing the playbook. An empty value, or a value less than 1 will use the Ansible default which is usually 5. The default number of forks can be overwritten with a change to ansible.cfg.",
			},
			"limit": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Provide a host pattern to further constrain the list of hosts that will be managed or affected by the playbook. Multiple patterns are allowed.",
			},
			"verbosity": schema.Int32Attribute{
				Optional:    true,
				Default:     int32default.StaticInt32(0),
				Computed:    true,
				Description: "Control the level of output ansible will produce as the playbook executes. `0 - Normal`, `1 - Verbose`, `2 - More Verbose`, `3 - Debug`, `4 - r.client.auth Debug`, `5 - WinRM Debug`",
				Validators: []validator.Int32{
					int32validator.Between(0, 5),
				},
			},
			"extra_vars": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString("---"),
				Computed:    true,
				Description: "Specify `extra_vars` for the template. Default value is `\"---\"`",
			},
			"job_tags": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Tags are useful when you have a large playbook, and you want to run a specific part of a play or task. Use commas to separate multiple tags.",
			},
			"force_handlers": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Enable forcing playbook handlers to run even if a task fails.",
			},
			"skip_tags": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Skip tags are useful when you have a large playbook, and you want to skip specific parts of a play or task. Use commas to separate multiple tags.",
			},
			"start_at_task": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Start the playbook at the task matching this name.",
			},
			"timeout": schema.Int32Attribute{
				Optional:    true,
				Default:     int32default.StaticInt32(0),
				Computed:    true,
				Description: "The amount of time (in seconds) to run before the job is canceled. Defaults to 0 for no job timeout.",
			},
			"use_fact_cache": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Enable use of fact caching for the job template.",
			},
			"execution_environment": schema.Int32Attribute{
				Optional:    true,
				Description: "Execution Environment ID to use for the job template.",
			},
			"host_config_key": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Allow provisioning callbacks using this host config key.",
			},
			"ask_scm_branch_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for scm branch on launch.",
			},
			"ask_diff_mode_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user to enable diff mode (show changes) to files when supported by modules.",
			},
			"ask_variables_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for variables on launch.",
			},
			"ask_limit_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for limit on launch.",
			},
			"ask_tags_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for tags on launch.",
			},
			"ask_skip_tags_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for skip tags on launch.",
			},
			"ask_job_type_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for job type on launch.",
			},
			"ask_verbosity_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for verbosity on launch.",
			},
			"ask_inventory_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for inventory on launch.",
			},
			"ask_credential_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for credential on launch.",
			},
			"ask_execution_environment_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for execution environment on launch.",
			},
			"ask_labels_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for labels on launch.",
			},
			"ask_forks_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for forks on launch.",
			},
			"ask_job_slice_count_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for job slice count on launch.",
			},
			"ask_timeout_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for timeout on launch.",
			},
			"ask_instance_groups_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt user for instance groups on launch.",
			},
			"survey_enabled": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Enable a survey on the job template.",
			},
			"become_enabled": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "If enabled, run this playbook as an administrator.",
			},
			"diff_mode": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "If enabled, show the changes made by Ansible tasks, where supported. This is equivalent to Ansible's `--diff` mode.",
			},
			"allow_simultaneous": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Allow simultaneous runs of the job template.",
			},
			"custom_virtualenv": schema.StringAttribute{
				Optional:    true,
				Description: "Local absolute file path containing a custom Python virtualenv to use. Only compatible with older versions of Tower. Deprecated, will be removed in the future",
			},
			"job_slice_count": schema.Int32Attribute{
				Optional:    true,
				Default:     int32default.StaticInt32(1),
				Computed:    true,
				Description: "Divide the work done by this job template into the specified number of job slices, each running the same tasks against a portion of the inventory.",
			},
			"webhook_service": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Service that webhook requests will be accepted from. Either `\"\"`, `bitbucket_dc`, `github`, or `gitlab`",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"", "github", "gitlab"}...),
				},
			},
			"webhook_credential": schema.StringAttribute{
				Optional:    true,
				Description: "Personal Access Token for posting back the status to the service API",
			},
			"prevent_instance_group_fallback": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "If enabled, the job template will prevent adding any inventory or organization instance groups to the list of preferred instances groups to run on. Note: If this setting is enabled and you provided an empty list, the global instance groups will be applied.",
			},
		},
	}
}

func (r JobTemplateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data JobTemplateModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If inventory is set, return without warning.
	if !data.Inventory.IsNull() {
		return
	}

	// If ask_inventory_on_launch is true, return without warning.
	if data.AskInventoryOnLaunch.ValueBool() {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("inventory"),
		"Missing Attribute Configuration",
		"Expected inventory to be specified or ask_inventory_on_launch = true.",
	)
}

func (r *JobTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*providerClient)

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
	var data JobTemplateModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData JobTemplateAPIModel
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
	if !(data.AskLabelsOnLaunch.IsNull()) {
		bodyData.AskLabelsOnLaunch = data.AskLabelsOnLaunch.ValueBool()
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

	url := "job_templates/"
	returnedData, statusCode, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{200, 201}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	if statusCode == 201 {
		resp.Diagnostics.AddWarning(
			"Job Template warning",
			fmt.Sprintf("The playbook file %v was not found in project %v, but the job template was created.", data.Playbook.ValueString(), int(data.Project.ValueInt32())))
	}

	returnedValues := []string{"id"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobTemplateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("job_templates/%d/", id)
	body, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "")
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

	var responseData JobTemplateAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable unmarshal response body into object",
			fmt.Sprintf("Error =  %v. ", err.Error()))
		return
	}

	if !data.Name.IsNull() || responseData.Name != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_type"), responseData.JobType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Inventory.IsNull() || responseData.Inventory != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inventory"), responseData.Inventory)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Project.IsNull() || responseData.Project != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project"), responseData.Project)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Playbook.IsNull() || responseData.Playbook != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("playbook"), responseData.Playbook)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_branch"), responseData.ScmBranch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("forks"), responseData.Forks)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("limit"), responseData.Limit)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("verbosity"), responseData.Verbosity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("extra_vars"), responseData.ExtraVars)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_tags"), responseData.JobTags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("force_handlers"), responseData.ForceHandlers)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("skip_tags"), responseData.SkipTags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("start_at_task"), responseData.StartAtTask)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeout"), responseData.Timeout)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("use_fact_cache"), responseData.UseFactCache)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ExecutionEnvironment.IsNull() || responseData.ExecutionEnvironment != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("execution_environment"), responseData.ExecutionEnvironment)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host_config_key"), responseData.HostConfigKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_scm_branch_on_launch"), responseData.AskScmBranchOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_diff_mode_on_launch"), responseData.AskDiffModeOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_variables_on_launch"), responseData.AskVariablesOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_limit_on_launch"), responseData.AskLimitOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_tags_on_launch"), responseData.AskTagsOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_skip_tags_on_launch"), responseData.AskSkipTagsOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_job_type_on_launch"), responseData.AskJobTypeOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_verbosity_on_launch"), responseData.AskVerbosityOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_inventory_on_launch"), responseData.AskInventoryOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_credential_on_launch"), responseData.AskCredentialOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_execution_environment_on_launch"), responseData.AskExecutionEnvironmenOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_labels_on_launch"), responseData.AskLabelsOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_forks_on_launch"), responseData.AskForksOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_job_slice_count_on_launch"), responseData.AskJobSliceCountOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_timeout_on_launch"), responseData.AskTimeoutOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_instance_groups_on_launch"), responseData.AskInstanceGroupsOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("survey_enabled"), responseData.SurveyEnabled)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("become_enabled"), responseData.BecomeEnabled)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("diff_mode"), responseData.DiffMode)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allow_simultaneous"), responseData.AllowSimultaneous)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.CustomVirtualEnv.IsNull() || responseData.CustomVirtualEnv != nil {
		if data.CustomVirtualEnv.ValueString() == "" && responseData.CustomVirtualEnv == nil {
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_slice_count"), responseData.JobSliceCount)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_service"), responseData.WebhookService)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.WebhookCredential.IsNull() || responseData.WebhookCredential != nil {
		if data.WebhookCredential.ValueString() == "" && responseData.WebhookCredential == nil {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_credential"), "")...)
		} else {
			if customWebhook, ok := responseData.WebhookCredential.(string); ok {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_credential"), customWebhook)...)
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prevent_instance_group_fallback"), responseData.PreventInstanceGroupFallback)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *JobTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JobTemplateModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData JobTemplateAPIModel
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
	bodyData.AskLabelsOnLaunch = data.AskLabelsOnLaunch.ValueBool()
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

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("job_templates/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data JobTemplateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}
	url := fmt.Sprintf("job_templates/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *JobTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
