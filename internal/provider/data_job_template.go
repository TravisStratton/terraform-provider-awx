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

var _ datasource.DataSource = &JobTemplateDataSource{}

func NewJobTemplateDataSource() datasource.DataSource {
	return &JobTemplateDataSource{}
}

type JobTemplateDataSource struct {
	client *AwxClient
}

func (d *JobTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_template"
}

func (d *JobTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get job_template datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Job template ID.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Job template name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Job template description.",
			},
			"job_type": schema.StringAttribute{
				Description: "Acceptable values are a choice of: `run`, `check`. For job templates, select run to execute the playbook. Select check to only check playbook syntax, test environment setup, and report problems without executing the playbook.",
				Computed:    true,
			},
			"inventory": schema.Int32Attribute{
				Computed:    true,
				Description: "ID number of the inventory to associate with the job template. Supply this or set `ask_inventory_on_launch = true`.",
			},
			"project": schema.Int32Attribute{
				Computed:    true,
				Description: "ID number of the project to associate with the job template",
			},
			"playbook": schema.StringAttribute{
				Computed:    true,
				Description: "Playbook name to be executed by this job",
			},
			"scm_branch": schema.StringAttribute{
				Computed:    true,
				Description: "Branch to use in job run. Project default used if blank. Only allowed if project allow_override field is set to true.",
			},
			"forks": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of parallel or simultaneous processes to use while executing the playbook. An empty value, or a value less than 1 will use the Ansible default which is usually 5. The default number of forks can be overwritten with a change to ansible.cfg.",
			},
			"limit": schema.StringAttribute{
				Computed:    true,
				Description: "Provide a host pattern to further constrain the list of hosts that will be managed or affected by the playbook. Multiple patterns are allowed.",
			},
			"verbosity": schema.Int32Attribute{
				Computed:    true,
				Description: "Control the level of output ansible will produce as the playbook executes. `0 - Normal`, `1 - Verbose`, `2 - More Verbose`, `3 - Debug`, `4 - r.client.auth Debug`, `5 - WinRM Debug`",
			},
			"extra_vars": schema.StringAttribute{
				Computed:    true,
				Description: "Specify `extra_vars` for the template.",
			},
			"job_tags": schema.StringAttribute{
				Computed:    true,
				Description: "Tags are useful when you have a large playbook, and you want to run a specific part of a play or task. Use commas to separate multiple tags.",
			},
			"force_handlers": schema.BoolAttribute{
				Computed:    true,
				Description: "Enable forcing playbook handlers to run even if a task fails.",
			},
			"skip_tags": schema.StringAttribute{
				Computed:    true,
				Description: "Skip tags are useful when you have a large playbook, and you want to skip specific parts of a play or task. Use commas to separate multiple tags.",
			},
			"start_at_task": schema.StringAttribute{
				Computed:    true,
				Description: "Start the playbook at the task matching this name.",
			},
			"timeout": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of time (in seconds) to run before the job is canceled. Defaults to 0 for no job timeout.",
			},
			"use_fact_cache": schema.BoolAttribute{
				Computed:    true,
				Description: "Enable use of fact caching for the job template.",
			},
			"execution_environment": schema.Int32Attribute{
				Computed:    true,
				Description: "Execution Environment ID to use for the job template.",
			},
			"host_config_key": schema.StringAttribute{
				Computed:    true,
				Description: "Allow provisioning callbacks using this host config key.",
			},
			"ask_scm_branch_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for scm branch on launch.",
			},
			"ask_diff_mode_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user to enable diff mode (show changes) to files when supported by modules.",
			},
			"ask_variables_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for variables on launch.",
			},
			"ask_limit_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for limit on launch.",
			},
			"ask_tags_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for tags on launch.",
			},
			"ask_skip_tags_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for skip tags on launch.",
			},
			"ask_job_type_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for job type on launch.",
			},
			"ask_verbosity_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for verbosity on launch.",
			},
			"ask_inventory_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for inventory on launch.",
			},
			"ask_credential_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for credential on launch.",
			},
			"ask_execution_environment_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for execution environment on launch.",
			},
			"ask_labels_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for labels on launch.",
			},
			"ask_forks_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for forks on launch.",
			},
			"ask_job_slice_count_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for job slice count on launch.",
			},
			"ask_timeout_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for timeout on launch.",
			},
			"ask_instance_groups_on_launch": schema.BoolAttribute{
				Computed:    true,
				Description: "Prompt user for instance groups on launch.",
			},
			"survey_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Enable a survey on the job template.",
			},
			"become_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If enabled, run this playbook as an administrator.",
			},
			"diff_mode": schema.BoolAttribute{
				Computed:    true,
				Description: "If enabled, show the changes made by Ansible tasks, where supported. This is equivalent to Ansible's `--diff` mode.",
			},
			"allow_simultaneous": schema.BoolAttribute{
				Computed:    true,
				Description: "Allow simultaneous runs of the job template.",
			},
			"custom_virtualenv": schema.StringAttribute{
				Computed:    true,
				Description: "Local absolute file path containing a custom Python virtualenv to use. Only compatible with older versions of AWX/Tower. Deprecated, will be removed in the future",
			},
			"job_slice_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Divide the work done by this job template into the specified number of job slices, each running the same tasks against a portion of the inventory.",
			},
			"webhook_service": schema.StringAttribute{
				Computed:    true,
				Description: "Service that webhook requests will be accepted from. Either `\"\"`, `bitbucket_dc`, `github`, or `gitlab`",
			},
			"webhook_credential": schema.StringAttribute{
				Computed:    true,
				Description: "Personal Access Token for posting back the status to the service API",
			},
			"prevent_instance_group_fallback": schema.BoolAttribute{
				Computed:    true,
				Description: "If enabled, the job template will prevent adding any inventory or organization instance groups to the list of preferred instances groups to run on. Note: If this setting is enabled and you provided an empty list, the global instance groups will be applied.",
			},
		},
	}
}

func (d JobTemplateDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *JobTemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*AwxClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = configureData
}

func (d *JobTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JobTemplateModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	if !data.Id.IsNull() {
		// set url for read by id HTTP request
		id, err := strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable convert id from string to int.",
				fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
			return
		}
		url = fmt.Sprintf("job_templates/%d/", id)
	}
	if !data.Name.IsNull() {
		// set url for read by name HTTP request
		name := urlParser.QueryEscape(data.Name.ValueString())
		url = fmt.Sprintf("job_templates/?name=%s", name)
	}

	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404})
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

	if !data.Id.IsNull() && data.Name.IsNull() {
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error =  %v.", err.Error()))
			return
		}
	}
	// If looking up by name, check that there is only one response and extract it.
	if data.Id.IsNull() && !data.Name.IsNull() {
		nameResult := struct {
			Count   int                   `json:"count"`
			Results []JobTemplateAPIModel `json:"results"`
		}{}
		err = json.Unmarshal(body, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count == 1 {
			responseData = nameResult.Results[0]
		} else {
			resp.Diagnostics.AddError(
				"Incorrect number of execution_environments returned by name",
				fmt.Sprintf("Unable to read execution_environment as API returned %v execution_environments.", nameResult.Count))
			return
		}
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Name = types.StringValue(responseData.Name)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}
	if responseData.JobType != "" {
		data.JobType = types.StringValue(responseData.JobType)
	}
	if responseData.Inventory != 0 {
		data.Inventory = types.Int32Value(int32(responseData.Inventory))
	}
	if responseData.Project != 0 {
		data.Project = types.Int32Value(int32(responseData.Project))
	}
	if responseData.Playbook != "" {
		data.Playbook = types.StringValue(responseData.Playbook)
	}
	if responseData.ScmBranch != "" {
		data.ScmBranch = types.StringValue(responseData.ScmBranch)
	}
	if responseData.Forks != 0 {
		data.Forks = types.Int32Value(int32(responseData.Forks))
	}
	if responseData.Limit != "" {
		data.Limit = types.StringValue(responseData.Limit)
	}
	if responseData.Verbosity != 0 {
		data.Verbosity = types.Int32Value(int32(responseData.Verbosity))
	}
	if responseData.ExtraVars != "" {
		data.ExtraVars = types.StringValue(responseData.ExtraVars)
	}
	if responseData.JobTags != "" {
		data.JobTags = types.StringValue(responseData.JobTags)
	}
	if responseData.ForceHandlers {
		data.ForceHandlers = types.BoolValue(responseData.ForceHandlers)
	}
	if responseData.SkipTags != "" {
		data.SkipTags = types.StringValue(responseData.SkipTags)
	}
	if responseData.StartAtTask != "" {
		data.StartAtTask = types.StringValue(responseData.StartAtTask)
	}
	if responseData.Timeout != 0 {
		data.Timeout = types.Int32Value(int32(responseData.Timeout))
	}
	if responseData.UseFactCache {
		data.UseFactCache = types.BoolValue(responseData.UseFactCache)
	}
	if responseData.ExecutionEnvironment != 0 {
		data.ExecutionEnvironment = types.Int32Value(int32(responseData.ExecutionEnvironment))
	}
	if responseData.HostConfigKey != "" {
		data.HostConfigKey = types.StringValue(responseData.HostConfigKey)
	}
	if responseData.AskScmBranchOnLaunch {
		data.AskScmBranchOnLaunch = types.BoolValue(responseData.AskScmBranchOnLaunch)
	}
	if responseData.AskDiffModeOnLaunch {
		data.AskDiffModeOnLaunch = types.BoolValue(responseData.AskDiffModeOnLaunch)
	}
	if responseData.AskVariablesOnLaunch {
		data.AskVariablesOnLaunch = types.BoolValue(responseData.AskVariablesOnLaunch)
	}
	if responseData.AskLimitOnLaunch {
		data.AskLimitOnLaunch = types.BoolValue(responseData.AskLimitOnLaunch)
	}
	if responseData.AskTagsOnLaunch {
		data.AskTagsOnLaunch = types.BoolValue(responseData.AskTagsOnLaunch)
	}
	if responseData.AskSkipTagsOnLaunch {
		data.AskSkipTagsOnLaunch = types.BoolValue(responseData.AskSkipTagsOnLaunch)
	}
	if responseData.AskJobTypeOnLaunch {
		data.AskJobTypeOnLaunch = types.BoolValue(responseData.AskJobTypeOnLaunch)
	}
	if responseData.AskVerbosityOnLaunch {
		data.AskVerbosityOnLaunch = types.BoolValue(responseData.AskVerbosityOnLaunch)
	}
	if responseData.AskInventoryOnLaunch {
		data.AskInventoryOnLaunch = types.BoolValue(responseData.AskInventoryOnLaunch)
	}
	if responseData.AskCredentialOnLaunch {
		data.AskCredentialOnLaunch = types.BoolValue(responseData.AskCredentialOnLaunch)
	}
	if responseData.AskExecutionEnvironmenOnLaunch {
		data.AskExecutionEnvironmenOnLaunch = types.BoolValue(responseData.AskExecutionEnvironmenOnLaunch)
	}
	if responseData.AskLabelsOnLaunch {
		data.AskLabelsOnLaunch = types.BoolValue(responseData.AskLabelsOnLaunch)
	}
	if responseData.AskForksOnLaunch {
		data.AskForksOnLaunch = types.BoolValue(responseData.AskForksOnLaunch)
	}
	if responseData.AskJobSliceCountOnLaunch {
		data.AskJobSliceCountOnLaunch = types.BoolValue(responseData.AskJobSliceCountOnLaunch)
	}
	if responseData.AskTimeoutOnLaunch {
		data.AskTimeoutOnLaunch = types.BoolValue(responseData.AskTimeoutOnLaunch)
	}
	if responseData.AskInstanceGroupsOnLaunch {
		data.AskInstanceGroupsOnLaunch = types.BoolValue(responseData.AskInstanceGroupsOnLaunch)
	}
	if responseData.SurveyEnabled {
		data.SurveyEnabled = types.BoolValue(responseData.SurveyEnabled)
	}
	if responseData.BecomeEnabled {
		data.BecomeEnabled = types.BoolValue(responseData.BecomeEnabled)
	}
	if responseData.DiffMode {
		data.DiffMode = types.BoolValue(responseData.DiffMode)
	}
	if responseData.AllowSimultaneous {
		data.AllowSimultaneous = types.BoolValue(responseData.AllowSimultaneous)
	}
	if responseData.CustomVirtualEnv == nil {
		data.CustomVirtualEnv = types.StringValue("")
	} else if responseData.CustomVirtualEnv != "" {
		if stringVirtualEnv, ok := responseData.CustomVirtualEnv.(string); ok {
			data.CustomVirtualEnv = types.StringValue(stringVirtualEnv)
		} else {
			resp.Diagnostics.AddError(
				"Invalid Type",
				"Expected responseData.CustomVirtualEnv to be a string",
			)
			return
		}
	}
	if responseData.JobSliceCount != 0 {
		data.JobSliceCount = types.Int32Value(int32(responseData.JobSliceCount))
	}
	if responseData.WebhookService != "" {
		data.WebhookService = types.StringValue(responseData.WebhookService)
	}
	if responseData.WebhookCredential == nil {
		data.WebhookCredential = types.StringValue("")
	} else if responseData.WebhookCredential != "" {
		if stringVirtualEnv, ok := responseData.WebhookCredential.(string); ok {
			data.WebhookCredential = types.StringValue(stringVirtualEnv)
		} else {
			resp.Diagnostics.AddError(
				"Invalid Type",
				"Expected responseData.WebhookCredential to be a string",
			)
			return
		}
	}
	if responseData.PreventInstanceGroupFallback {
		data.PreventInstanceGroupFallback = types.BoolValue(responseData.PreventInstanceGroupFallback)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
