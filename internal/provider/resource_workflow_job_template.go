package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WorkflowJobTemplatesResource{}
var _ resource.ResourceWithImportState = &WorkflowJobTemplatesResource{}

func NewWorkflowJobTemplatesResource() resource.Resource {
	return &WorkflowJobTemplatesResource{}
}

type WorkflowJobTemplatesResource struct {
	client *AwxClient
}

func (r *WorkflowJobTemplatesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_job_template"
}

func (r *WorkflowJobTemplatesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A workflow job template resource.",

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
			"extra_vars": schema.StringAttribute{
				Optional: true,
				Default:  stringdefault.StaticString("---"),
				Computed: true,
			},
			"organization": schema.Int32Attribute{
				Required: true,
			},
			"survey_enabled": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Defaults to `false`.",
			},
			"allow_simultaneous": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Defaults to `false`.",
			},
			"ask_variables_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Defaults to `false`.",
			},
			"inventory": schema.Int32Attribute{
				Optional:    true,
				Description: "Inventory ID of the inventory containing the hosts you want this job to manage.",
			},
			"limit": schema.StringAttribute{
				Optional:    true,
				Description: "Provide a host pattern to further constrain the list of hosts that will be managed or affected by the playbook. Multiple patterns are allowed.",
			},
			"scm_branch": schema.StringAttribute{
				Optional:    true,
				Description: "Select a branch for the workflow. This branch is applied to all job template nodes that prompt for a branch.",
			},
			"ask_inventory_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt on launch for Inventory. Defaults to `false`.",
			},
			"ask_scm_branch_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt on launch for Source Control Branch. Defaults to `false`.",
			},
			"ask_limit_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt on launch for Limit. Defaults to `false`.",
			},
			"webhook_service": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Defaults to `\"\"`.",
			},
			"webhook_credential": schema.StringAttribute{
				Optional: true,
			},
			"ask_labels_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt on launch for Labels. Defaults to `false`.",
			},
			"ask_skip_tags_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt on launch for Skip Tags. Defaults to `false`.",
			},
			"ask_tags_on_launch": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
				Description: "Prompt on launch for Ask Tags. Defaults to `false`.",
			},
			"skip_tags": schema.StringAttribute{
				Optional:    true,
				Description: "Tags are useful when you have a large playbook, and you want to run a specific part of a play or task. Use commas to separate multiple tags.",
			},
			"job_tags": schema.StringAttribute{
				Optional:    true,
				Description: "Skip tags are useful when you have a large playbook, and you want to skip specific parts of a play or task. Use commas to separate multiple tags.",
			},
		},
	}
}

func (r *WorkflowJobTemplatesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowJobTemplatesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowJobTemplatesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData WorkflowJobTemplateAPIModel

	if !data.Name.IsNull() {
		bodyData.Name = data.Name.ValueString()
	}
	if !data.Description.IsNull() {
		bodyData.Description = data.Description.ValueString()
	}
	if !data.ExtraVars.IsNull() {
		bodyData.ExtraVars = data.ExtraVars.ValueString()
	}
	if !data.Organization.IsNull() {
		bodyData.Organization = int(data.Organization.ValueInt32())
	}
	if !data.SurveyEnabled.IsNull() {
		bodyData.SurveyEnabled = data.SurveyEnabled.ValueBool()
	}
	if !data.AllowSimultaneous.IsNull() {
		bodyData.AllowSimultaneous = data.AllowSimultaneous.ValueBool()
	}
	if !data.AskVariablesOnLaunch.IsNull() {
		bodyData.AskVariablesOnLaunch = data.AskVariablesOnLaunch.ValueBool()
	}
	if !data.Inventory.IsNull() {
		bodyData.Inventory = int(data.Inventory.ValueInt32())
	}
	if !data.Limit.IsNull() {
		bodyData.Limit = data.Limit.ValueString()
	}
	if !data.ScmBranch.IsNull() {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !data.AskInventoryOnLaunch.IsNull() {
		bodyData.AskInventoryOnLaunch = data.AskInventoryOnLaunch.ValueBool()
	}
	if !data.AskScmBranchOnLaunch.IsNull() {
		bodyData.AskScmBranchOnLaunch = data.AskScmBranchOnLaunch.ValueBool()
	}
	if !data.AskLimitOnLaunch.IsNull() {
		bodyData.AskLimitOnLaunch = data.AskLimitOnLaunch.ValueBool()
	}
	if !data.WebhookService.IsNull() {
		bodyData.WebhookService = data.WebhookService.ValueString()
	}
	if !data.WebhookCredential.IsNull() {
		bodyData.WebhookCredential = data.WebhookCredential.ValueString()
	}
	if !data.AskLabelsOnLaunch.IsNull() {
		bodyData.AskLabelsOnLaunch = data.AskLabelsOnLaunch.ValueBool()
	}
	if !data.AskSkipTagsOnLaunch.IsNull() {
		bodyData.AskSkipTagsOnLaunch = data.AskSkipTagsOnLaunch.ValueBool()
	}
	if !data.AskTagsOnLaunch.IsNull() {
		bodyData.AskTagsOnLaunch = data.AskTagsOnLaunch.ValueBool()
	}
	if !data.SkipTags.IsNull() {
		bodyData.SkipTags = data.SkipTags.ValueString()
	}
	if !data.JobTags.IsNull() {
		bodyData.JobTags = data.JobTags.ValueString()
	}

	url := "/api/v2/workflow_job_templates/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
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

func (r *WorkflowJobTemplatesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowJobTemplatesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
	}

	url := fmt.Sprintf("/api/v2/workflow_job_templates/%d/", id)
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

	var responseData WorkflowJobTemplateAPIModel

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

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.ExtraVars.IsNull() || responseData.ExtraVars != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("extra_vars"), responseData.ExtraVars)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Organization.IsNull() || responseData.Organization != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), responseData.Organization)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("survey_enabled"), responseData.SurveyEnabled)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allow_simultaneous"), responseData.AllowSimultaneous)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_variables_on_launch"), responseData.AskVariablesOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Inventory.IsNull() || responseData.Inventory != nil {
		convertInt, ok := responseData.Inventory.(float64)
		if !ok {
			resp.Diagnostics.AddError("Unable to convert any to float64.", fmt.Sprintf("Unable to convert %v to float64.", responseData.Inventory))
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inventory"), int(convertInt))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.Limit.IsNull() || responseData.Limit != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("limit"), responseData.Limit)...)
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_inventory_on_launch"), responseData.AskInventoryOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_scm_branch_on_launch"), responseData.AskScmBranchOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_limit_on_launch"), responseData.AskLimitOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_service"), responseData.WebhookService)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.WebhookCredential.IsNull() || responseData.WebhookCredential != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("webhook_credential"), responseData.WebhookCredential)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_labels_on_launch"), responseData.AskLabelsOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_skip_tags_on_launch"), responseData.AskSkipTagsOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ask_tags_on_launch"), responseData.AskTagsOnLaunch)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.SkipTags.IsNull() || responseData.SkipTags != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("skip_tags"), responseData.SkipTags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.JobTags.IsNull() || responseData.JobTags != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_tags"), responseData.JobTags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

}

func (r *WorkflowJobTemplatesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowJobTemplatesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

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

	var bodyData WorkflowJobTemplateAPIModel
	bodyData.Name = data.Name.ValueString()
	if !data.Description.IsNull() {
		bodyData.Description = data.Description.ValueString()
	}
	if !data.ExtraVars.IsNull() {
		bodyData.ExtraVars = data.ExtraVars.ValueString()
	}
	bodyData.Organization = int(data.Organization.ValueInt32())
	bodyData.SurveyEnabled = data.SurveyEnabled.ValueBool()
	bodyData.AllowSimultaneous = data.AllowSimultaneous.ValueBool()
	bodyData.AskVariablesOnLaunch = data.AskVariablesOnLaunch.ValueBool()
	if !data.Inventory.IsNull() {
		bodyData.Inventory = int(data.Inventory.ValueInt32())
	}
	bodyData.Limit = data.Limit.ValueString()
	bodyData.ScmBranch = data.ScmBranch.ValueString()
	bodyData.AskInventoryOnLaunch = data.AskInventoryOnLaunch.ValueBool()
	bodyData.AskScmBranchOnLaunch = data.AskScmBranchOnLaunch.ValueBool()
	bodyData.AskLimitOnLaunch = data.AskLimitOnLaunch.ValueBool()
	bodyData.WebhookService = data.WebhookService.ValueString()
	bodyData.WebhookCredential = data.WebhookCredential.ValueString()
	bodyData.AskLabelsOnLaunch = data.AskLabelsOnLaunch.ValueBool()
	bodyData.AskSkipTagsOnLaunch = data.AskSkipTagsOnLaunch.ValueBool()
	bodyData.AskTagsOnLaunch = data.AskTagsOnLaunch.ValueBool()
	bodyData.SkipTags = data.SkipTags.ValueString()
	bodyData.JobTags = data.JobTags.ValueString()

	url := fmt.Sprintf("/api/v2/workflow_job_templates/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowJobTemplatesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowJobTemplatesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
	}

	url := fmt.Sprintf("/api/v2/workflow_job_templates/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *WorkflowJobTemplatesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
