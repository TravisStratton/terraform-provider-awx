package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type JTChildAPIRead struct {
	Count   int           `json:"count"`
	Results []ChildResult `json:"results"`
}

type ChildResult struct {
	Id int `json:"id"`
}

type ChildAssocBody struct {
	Id        int  `json:"id"`
	Associate bool `json:"associate"`
}

type ChildDissasocBody struct {
	Id           int  `json:"id"`
	Disassociate bool `json:"disassociate"`
}

type JTLabelsAPIRead struct {
	Count        int           `json:"count"`
	LabelResults []LabelResult `json:"results"`
}

type LabelResult struct {
	Id int `json:"id"`
}

type LabelDissasocBody struct {
	Id           int  `json:"id"`
	Disassociate bool `json:"disassociate"`
}

type CredentialModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Organization   types.Int32  `tfsdk:"organization"`
	Team           types.Int32  `tfsdk:"team"`
	User           types.Int32  `tfsdk:"user"`
	CredentialType types.Int32  `tfsdk:"credential_type"`
	Kind           types.String `tfsdk:"kind"`
	Inputs         types.String `tfsdk:"inputs"`
}

type CredentialAPIModel struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Organization   int    `json:"organization,omitempty"`
	Team           int    `json:"team,omitempty"`
	User           int    `json:"user,omitempty"`
	CredentialType int    `json:"credential_type"`
	Kind           string `json:"kind,omitempty"`
	Inputs         any    `json:"inputs,omitempty"`
}

type CredentialTypeModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Inputs      types.String `tfsdk:"inputs"`
	Injectors   types.String `tfsdk:"injectors"`
	Kind        types.String `tfsdk:"kind"`
}

type CredentialTypeAPIModel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Inputs      any    `json:"inputs,omitempty"`
	Injectors   any    `json:"injectors,omitempty"`
	Kind        string `json:"kind"`
}

type ExecutionEnvironmentModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Image        types.String `tfsdk:"image"`
	Pull         types.String `tfsdk:"pull"`
	Organization types.Int32  `tfsdk:"organization"`
	Credential   types.Int32  `tfsdk:"credential"`
}

type ExecutionEnvironmentAPIModel struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Image        string `json:"image"`
	Pull         string `json:"pull"`
	Organization int    `json:"organization,omitempty"`
	Credential   int    `json:"credential,omitempty"`
}

type HostModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Inventory   types.Int32  `tfsdk:"inventory"`
	Variables   types.String `tfsdk:"variables"`
}

type HostAPIModel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Inventory   int    `json:"inventory"`
	Variables   string `json:"variables,omitempty"`
}

type GroupModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Inventory   types.Int32  `tfsdk:"inventory"`
	Variables   types.String `tfsdk:"variables"`
}

type GroupAPIModel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Inventory   int    `json:"inventory"`
	Variables   string `json:"variables,omitempty"`
}

type InstanceGroupModel struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	IsContainerGroup         types.Bool   `tfsdk:"is_container_group"`
	MaxConcurrentJobs        types.Int32  `tfsdk:"max_concurrent_jobs"`
	MaxForks                 types.Int32  `tfsdk:"max_forks"`
	PodSpecOverride          types.String `tfsdk:"pod_spec_override"`
	PolicyInstancePercentage types.Int32  `tfsdk:"policy_instance_percentage"`
	PolicyInstanceMinimum    types.Int32  `tfsdk:"policy_instance_minimum"`
	Credential               types.Int32  `tfsdk:"credential"`
}

type InstanceGroupAPIModel struct {
	Id                       int    `json:"id"`
	Name                     string `json:"name"`
	IsContainerGroup         bool   `json:"is_container_group"`
	MaxConcurrentJobs        int    `json:"max_concurrent_jobs"`
	MaxForks                 int    `json:"max_forks"`
	PodSpecOverride          any    `json:"pod_spec_override,omitempty"`
	PolicyInstancePercentage int    `json:"policy_instance_percentage"`
	PolicyInstanceMinimum    int    `json:"policy_instance_minimum"`
	Credential               int    `json:"credential,omitempty"`
}

type InventoryModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Organization types.Int32  `tfsdk:"organization"`
	Variables    types.String `tfsdk:"variables"`
	Kind         types.String `tfsdk:"kind"`
	HostFilter   types.String `tfsdk:"host_filter"`
}

type InventoryAPIModel struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Organization int    `json:"organization"`
	Variables    string `json:"variables,omitempty"`
	Kind         string `json:"kind,omitempty"`
	HostFilter   string `json:"host_filter,omitempty"`
}

type InventorySourceModel struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Inventory            types.Int32  `tfsdk:"inventory"`
	Source               types.String `tfsdk:"source"`
	Credential           types.Int32  `tfsdk:"credential"`
	Description          types.String `tfsdk:"description"`
	ExecutionEnvironment types.Int32  `tfsdk:"execution_environment"`
	SourcePath           types.String `tfsdk:"source_path"`
	EnabledValue         types.String `tfsdk:"enabled_value"`
	EnabledVar           types.String `tfsdk:"enabled_var"`
	HostFilter           types.String `tfsdk:"host_filter"`
	OverwriteVars        types.Bool   `tfsdk:"overwrite_vars"`
	Overwrite            types.Bool   `tfsdk:"overwrite"`
	SourceVars           types.String `tfsdk:"source_vars"`
	SourceProject        types.Int32  `tfsdk:"source_project"`
	ScmBranch            types.String `tfsdk:"scm_branch"`
	UpdateCacheTimeout   types.Int32  `tfsdk:"update_cache_timeout"`
	UpdateOnLaunch       types.Bool   `tfsdk:"update_on_launch"`
	Verbosity            types.Int32  `tfsdk:"verbosity"`
}

type InventorySourceAPIModel struct {
	Id                   int    `json:"id"`
	Name                 string `json:"name"`
	Inventory            int    `json:"inventory"`
	Source               string `json:"source"`
	Credential           int    `json:"credential,omitempty"`
	Description          string `json:"description,omitempty"`
	ExecutionEnvironment any    `json:"execution_environment"`
	SourcePath           string `json:"source_path,omitempty"`
	EnabledValue         string `json:"enabled_value,omitempty"`
	EnabledVar           string `json:"enabled_var,omitempty"`
	HostFilter           string `json:"host_filter,omitempty"`
	OverwriteVars        bool   `json:"overwrite_vars,omitempty"`
	Overwrite            bool   `json:"overwrite,omitempty"`
	SourceVars           string `json:"source_vars,omitempty"`
	SourceProject        int    `json:"source_project,omitempty"`
	ScmBranch            string `json:"scm_branch,omitempty"`
	UpdateCacheTimeout   int    `json:"update_cache_timeout,omitempty"`
	UpdateOnLaunch       bool   `json:"update_on_launch,omitempty"`
	Verbosity            int    `json:"verbosity,omitempty"`
}

type JobTemplateModel struct {
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
	StartAtTask                    types.String `tfsdk:"start_at_task"`
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
	AskLabelsOnLaunch              types.Bool   `tfsdk:"ask_labels_on_launch"`
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

type JobTemplateAPIModel struct {
	Id                             int    `json:"id"`
	Name                           string `json:"name,omitempty"`
	Description                    string `json:"description"`
	JobType                        string `json:"job_type,omitempty"`
	Inventory                      int    `json:"inventory,omitempty"`
	Project                        int    `json:"project,omitempty"`
	Playbook                       string `json:"playbook,omitempty"`
	ScmBranch                      string `json:"scm_branch,omitempty"`
	Forks                          int    `json:"forks,omitempty"`
	Limit                          string `json:"limit,omitempty"`
	Verbosity                      int    `json:"verbosity,omitempty"`
	ExtraVars                      string `json:"extra_vars,omitempty"`
	JobTags                        string `json:"job_tags,omitempty"`
	ForceHandlers                  bool   `json:"force_handlers,omitempty"`
	SkipTags                       string `json:"skip_tags,omitempty"`
	StartAtTask                    string `json:"start_at_task,omitempty"`
	Timeout                        int    `json:"timeout,omitempty"`
	UseFactCache                   bool   `json:"use_fact_cache,omitempty"`
	ExecutionEnvironment           int    `json:"execution_environment,omitempty"`
	HostConfigKey                  string `json:"host_config_key,omitempty"`
	AskScmBranchOnLaunch           bool   `json:"ask_scm_branch_on_launch,omitempty"`
	AskDiffModeOnLaunch            bool   `json:"ask_diff_mode_on_launch,omitempty"`
	AskVariablesOnLaunch           bool   `json:"ask_variables_on_launch,omitempty"`
	AskLimitOnLaunch               bool   `json:"ask_limit_on_launch,omitempty"`
	AskTagsOnLaunch                bool   `json:"ask_tags_on_launch,omitempty"`
	AskSkipTagsOnLaunch            bool   `json:"ask_skip_tags_on_launch,omitempty"`
	AskJobTypeOnLaunch             bool   `json:"ask_job_type_on_launch,omitempty"`
	AskVerbosityOnLaunch           bool   `json:"ask_verbosity_on_launch,omitempty"`
	AskInventoryOnLaunch           bool   `json:"ask_inventory_on_launch,omitempty"`
	AskCredentialOnLaunch          bool   `json:"ask_credential_on_launch,omitempty"`
	AskExecutionEnvironmenOnLaunch bool   `json:"ask_execution_environment_on_launch,omitempty"`
	AskLabelsOnLaunch              bool   `json:"ask_labels_on_launch,omitempty"`
	AskForksOnLaunch               bool   `json:"ask_forks_on_launch,omitempty"`
	AskJobSliceCountOnLaunch       bool   `json:"ask_job_slice_count_on_launch,omitempty"`
	AskTimeoutOnLaunch             bool   `json:"ask_timeout_on_launch,omitempty"`
	AskInstanceGroupsOnLaunch      bool   `json:"ask_instance_groups_on_launch,omitempty"`
	SurveyEnabled                  bool   `json:"survey_enabled,omitempty"`
	BecomeEnabled                  bool   `json:"become_enabled,omitempty"`
	DiffMode                       bool   `json:"diff_mode,omitempty"`
	AllowSimultaneous              bool   `json:"allow_simultaneous,omitempty"`
	CustomVirtualEnv               any    `json:"custom_virtualenv,omitempty"` //blank is returned by api as "custom_virtual": null (not "")
	JobSliceCount                  int    `json:"job_slice_count,omitempty"`
	WebhookService                 string `json:"webhook_service,omitempty"`
	WebhookCredential              any    `json:"webhook_credential,omitempty"` //blank is returned by api as "webhook_credentials": null (not "")
	PreventInstanceGroupFallback   bool   `json:"prevent_instance_group_fallback,omitempty"`
}

type LabelModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Organization types.Int32  `tfsdk:"organization"`
}

type LabelAPIModel struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Organization int    `json:"organization"`
}

type NotificationTemplateModel struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Organization              types.Int32  `tfsdk:"organization"`
	NotificationType          types.String `tfsdk:"notification_type"`
	NotificationConfiguration types.String `tfsdk:"notification_configuration"`
	Messages                  types.String `tfsdk:"messages"`
}

type NotificationTemplateAPIModel struct {
	Id                        int    `json:"id"`
	Name                      string `json:"name"`
	Description               string `json:"description,omitempty"`
	Organization              int    `json:"organization"`
	NotificationType          string `json:"notification_type"`
	NotificationConfiguration any    `json:"notification_configuration,omitempty"`
	Messages                  any    `json:"messages,omitempty"`
}

type SlackConfiguration struct {
	Channels  []string `json:"channels"`
	HexColors string   `json:"hex_color"`
	Token     string   `json:"token"`
}

type WebhookConfiguration struct {
	Url                    string         `json:"url"`
	Headers                map[string]any `json:"headers"`
	Password               string         `json:"password"`
	Username               string         `json:"username"`
	HttpMethod             string         `json:"http_method"`
	DisableSslVerification bool           `json:"disable_ssl_verification"`
}

type MessageValue struct {
	Body    string `json:"body"`
	Message string `json:"message"`
}

type Messages struct {
	Error            MessageValue            `json:"error"`
	Started          MessageValue            `json:"started"`
	Success          MessageValue            `json:"success"`
	WorkflowApproval map[string]MessageValue `json:"workflow_approval"`
}

type OrganizationModel struct {
	Id             types.String `tfsdk:"id"`
	Aap25GatewayId types.Int32  `tfsdk:"aap25_gateway_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	DefaultEnv     types.Int32  `tfsdk:"default_environment"`
	MaxHosts       types.Int32  `tfsdk:"max_hosts"`
}

type contextKey string

type OrganizationAPIModel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	DefaultEnv  int    `json:"default_environment,omitempty"`
	MaxHosts    int    `json:"max_hosts,omitempty"`
}

type ProjectModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Organization       types.Int32  `tfsdk:"organization"`
	ScmType            types.String `tfsdk:"scm_type"`
	Description        types.String `tfsdk:"description"`
	AllowOverride      types.Bool   `tfsdk:"allow_override"`
	Credential         types.Int32  `tfsdk:"credential"`
	DefaultEnv         types.Int32  `tfsdk:"default_environment"`
	LocalPath          types.String `tfsdk:"local_path"`
	ScmBranch          types.String `tfsdk:"scm_branch"`
	ScmClean           types.Bool   `tfsdk:"scm_clean"`
	ScmDelOnUpdate     types.Bool   `tfsdk:"scm_delete_on_update"`
	ScmRefSpec         types.String `tfsdk:"scm_refspec"`
	ScmTrackSubmodules types.Bool   `tfsdk:"scm_track_submodules"`
	ScmUpdOnLaunch     types.Bool   `tfsdk:"scm_update_on_launch"`
	ScmUrl             types.String `tfsdk:"scm_url"`
	Timeout            types.Int32  `tfsdk:"timeout"`
}

type ProjectAPIModel struct {
	Id                 int    `json:"id"`
	Name               string `json:"name"`
	Organization       int    `json:"organization"`
	ScmType            string `json:"scm_type"`
	Description        string `json:"description,omitempty"`
	AllowOverride      bool   `json:"allow_override,omitempty"`
	Credential         int    `json:"credential,omitempty"`
	DefaultEnv         int    `json:"default_environment,omitempty"`
	LocalPath          string `json:"local_path,omitempty"`
	ScmBranch          string `json:"scm_branch,omitempty"`
	ScmClean           bool   `json:"scm_clean,omitempty"`
	ScmDelOnUpdate     bool   `json:"scm_delete_on_update,omitempty"`
	ScmRefSpec         string `json:"scm_refspec,omitempty"`
	ScmTrackSubmodules bool   `json:"scm_track_submodules,omitempty"`
	ScmUpdOnLaunch     bool   `json:"scm_update_on_launch,omitempty"`
	ScmUrl             string `json:"scm_url,omitempty"`
	Timeout            int    `json:"timeout"`
}

type ScheduleModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	UnifiedJobTemplate types.Int32  `tfsdk:"unified_job_template"`
	Rrule              types.String `tfsdk:"rrule"`
	Enabled            types.Bool   `tfsdk:"enabled"`
}

type ScheduleAPIModel struct {
	Id                 int    `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	UnifiedJobTemplate int    `json:"unified_job_template"`
	Rrule              string `json:"rrule"`
	Enabled            bool   `json:"enabled"`
}

type UserModel struct {
	Id              types.String `tfsdk:"id"`
	Username        types.String `tfsdk:"username"`
	FirstName       types.String `tfsdk:"first_name"`
	LastName        types.String `tfsdk:"last_name"`
	Email           types.String `tfsdk:"email"`
	IsSuperuser     types.Bool   `tfsdk:"is_superuser"`
	IsSystemAuditor types.Bool   `tfsdk:"is_system_auditor"`
	Password        types.String `tfsdk:"password"`
}

type UserAPIModel struct {
	Id              int    `json:"id"`
	Username        string `json:"username"`
	FirstName       string `json:"first_name,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	Email           string `json:"email,omitempty"`
	IsSuperuser     bool   `json:"is_superuser"`
	IsSystemAuditor bool   `json:"is_system_auditor"`
	Password        string `json:"password"`
}

type WorkflowJobTemplatesResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	ExtraVars            types.String `tfsdk:"extra_vars"`
	Organization         types.Int32  `tfsdk:"organization"`
	SurveyEnabled        types.Bool   `tfsdk:"survey_enabled"`
	AllowSimultaneous    types.Bool   `tfsdk:"allow_simultaneous"`
	AskVariablesOnLaunch types.Bool   `tfsdk:"ask_variables_on_launch"`
	Inventory            types.Int32  `tfsdk:"inventory"`
	Limit                types.String `tfsdk:"limit"`
	ScmBranch            types.String `tfsdk:"scm_branch"`
	AskInventoryOnLaunch types.Bool   `tfsdk:"ask_inventory_on_launch"`
	AskScmBranchOnLaunch types.Bool   `tfsdk:"ask_scm_branch_on_launch"`
	AskLimitOnLaunch     types.Bool   `tfsdk:"ask_limit_on_launch"`
	WebhookService       types.String `tfsdk:"webhook_service"`
	WebhookCredential    types.String `tfsdk:"webhook_credential"`
	AskLabelsOnLaunch    types.Bool   `tfsdk:"ask_labels_on_launch"`
	AskSkipTagsOnLaunch  types.Bool   `tfsdk:"ask_skip_tags_on_launch"`
	AskTagsOnLaunch      types.Bool   `tfsdk:"ask_tags_on_launch"`
	SkipTags             types.String `tfsdk:"skip_tags"`
	JobTags              types.String `tfsdk:"job_tags"`
}

type WorkflowJobTemplateAPIModel struct {
	Id                   int    `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	ExtraVars            string `json:"extra_vars"`
	Organization         int    `json:"organization"`
	SurveyEnabled        bool   `json:"survey_enabled"`
	AllowSimultaneous    bool   `json:"allow_simultaneous"`
	AskVariablesOnLaunch bool   `json:"ask_variables_on_launch"`
	Inventory            any    `json:"inventory"`
	Limit                string `json:"limit,omitempty"`
	ScmBranch            string `json:"scm_branch,omitempty"`
	AskInventoryOnLaunch bool   `json:"ask_inventory_on_launch"`
	AskScmBranchOnLaunch bool   `json:"ask_scm_branch_on_launch"`
	AskLimitOnLaunch     bool   `json:"ask_limit_on_launch"`
	WebhookService       string `json:"webhook_service"`
	WebhookCredential    string `json:"webhook_credential"`
	AskLabelsOnLaunch    bool   `json:"ask_labels_on_launch"`
	AskSkipTagsOnLaunch  bool   `json:"ask_skip_tags_on_launch"`
	AskTagsOnLaunch      bool   `json:"ask_tags_on_launch"`
	SkipTags             string `json:"skip_tags,omitempty"`
	JobTags              string `json:"job_tags,omitempty"`
}

type CredentialInputSourcesModel struct {
	Id               types.String `tfsdk:"id"`
	Description      types.String `tfsdk:"description"`
	InputFieldName   types.String `tfsdk:"input_field_name"`
	Metadata         types.Map    `tfsdk:"metadata"`
	TargetCredential types.Int32  `tfsdk:"target_credential"`
	SourceCredential types.Int32  `tfsdk:"source_credential"`
}

type CredentialInputSourcesAPIModel struct {
	Id               int               `json:"id,omitempty"`
	Description      string            `json:"description"`
	InputFieldName   string            `json:"input_field_name"`
	Metadata         map[string]string `json:"metadata"`
	TargetCredential int               `json:"target_credential"`
	SourceCredential int               `json:"source_credential"`
}
