package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WorkflowJobTemplatesJobNodeResource{}
var _ resource.ResourceWithImportState = &WorkflowJobTemplatesJobNodeResource{}

func NewWorkflowJobTemplatesJobNodeResource() resource.Resource {
	return &WorkflowJobTemplatesJobNodeResource{}
}

type WorkflowJobTemplatesJobNodeResource struct {
	client *providerClient
}

type WorkflowJobTemplatesJobNodeResourceModel struct {
	Id                     types.String `tfsdk:"id"`
	WorkflowJobId          types.Int32  `tfsdk:"workflow_job_template_id"`
	UnifiedJobTemplateId   types.Int32  `tfsdk:"unified_job_template"`
	Inventory              types.Int32  `tfsdk:"inventory"`
	ExtraData              types.String `tfsdk:"extra_data"`
	ScmBranch              types.String `tfsdk:"scm_branch"`
	JobType                types.String `tfsdk:"job_type"`
	JobTags                types.String `tfsdk:"job_tags"`
	SkipTags               types.String `tfsdk:"skip_tags"`
	Limit                  types.String `tfsdk:"limit"`
	DiffMode               types.Bool   `tfsdk:"diff_mode"`
	Verbosity              types.Int32  `tfsdk:"verbosity"`
	AllParentsMustConverge types.Bool   `tfsdk:"all_parents_must_converge"`
	Identifier             types.String `tfsdk:"identifier"`
}

type WorkflowJobTemplateNodeAPIModel struct {
	WorkflowJobId          int    `json:"workflow_job_template"`
	UnifiedJobTemplateId   int    `json:"unified_job_template"`
	Inventory              any    `json:"inventory"`
	ExtraData              any    `json:"extra_data,omitempty"`
	ScmBranch              string `json:"scm_branch,omitempty"`
	JobType                string `json:"job_type,omitempty"`
	JobTags                string `json:"job_tags,omitempty"`
	SkipTags               string `json:"skip_tags,omitempty"`
	Limit                  string `json:"limit,omitempty"`
	DiffMode               any    `json:"diff_mode,omitempty"`
	Verbosity              int    `json:"verbosity,omitempty"`
	AllParentsMustConverge bool   `json:"all_parents_must_converge"`
	Identifier             string `json:"identifier,omitempty"`
}

func (r *WorkflowJobTemplatesJobNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_job_template_job_node"
}

func (r *WorkflowJobTemplatesJobNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A node has its own ID, but must be associated to a workflow job template ID. And, the node must specify which actual job template this node will run, which is set using the attribute unified_job_template.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique ID for this node",
			},
			"workflow_job_template_id": schema.Int32Attribute{
				Required:    true,
				Description: "This is the ID of the workflow job template to which this node is embedded.",
			},
			"unified_job_template": schema.Int32Attribute{
				Required:    true,
				Description: "This is the ID of the actual job template that should be executed as this node.",
			},
			"inventory": schema.Int32Attribute{
				Optional:    true,
				Description: "ID of the Inventory applied as a prompt, if job template prompts for inventory.",
			},
			"extra_data": schema.StringAttribute{
				Optional:    true,
				Description: "Variables to apply at launch time. JSON Key/value pairs, wrap in `jsonencode()`.  Will only be accepted if job template prompts for vars or has a survey asking for those vars.",
			},
			"scm_branch": schema.StringAttribute{
				Optional:    true,
				Description: "SCM branch applied as a prompt, if job template prompts for SCM branch.",
			},
			"job_type": schema.StringAttribute{
				Optional:    true,
				Description: "Job type applied as a prompt, if job template prompts for job type.",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"run", "check"}...),
				},
			},
			"job_tags": schema.StringAttribute{
				Optional:    true,
				Description: "Job tags applied as a prompt, if job template prompts for job tags.",
			},
			"skip_tags": schema.StringAttribute{
				Optional:    true,
				Description: "Tags to skip, applied as a prompt, if job template prompts for job tags.",
			},
			"limit": schema.StringAttribute{
				Optional:    true,
				Description: "Limit to act on, applied as a prompt, if job template prompts for limit.",
			},
			"diff_mode": schema.BoolAttribute{
				Optional:    true,
				Description: "Run diff mode, applied as a prompt, if job template prompts for diff mode.",
			},
			"verbosity": schema.Int32Attribute{
				Optional:    true,
				Description: "Verbosity applied as a prompt, if job template prompts for verbosity. Control the level of output ansible will produce as the playbook executes. `0 - Normal`, `1 - Verbose`, `2 - More Verbose`, `3 - Debug`, `4 - r.client.auth Debug`, `5 - WinRM Debug`",
				Validators: []validator.Int32{
					int32validator.Between(0, 5),
				},
			},
			"all_parents_must_converge": schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If enabled then the node will only run if all of the parent nodes have met the criteria to reach this node. Defaults to `false`.",
			},
			"identifier": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for this node, set automatically by API when creating a new one.",
			},
		},
	}
}

func (r *WorkflowJobTemplatesJobNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowJobTemplatesJobNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowJobTemplatesJobNodeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData WorkflowJobTemplateNodeAPIModel

	if !data.WorkflowJobId.IsNull() {
		bodyData.WorkflowJobId = int(data.WorkflowJobId.ValueInt32())
	}
	if !data.UnifiedJobTemplateId.IsNull() {
		bodyData.UnifiedJobTemplateId = int(data.UnifiedJobTemplateId.ValueInt32())
	}
	if !data.Inventory.IsNull() {
		bodyData.Inventory = int(data.Inventory.ValueInt32())
	}
	if !data.ExtraData.IsNull() {
		extraDataMap := new(map[string]any)
		err := json.Unmarshal([]byte(data.ExtraData.ValueString()), &extraDataMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable unmarshal map to json",
				fmt.Sprintf("Unable to convert id: %+v. ", data.ExtraData))
			return
		}

		bodyData.ExtraData = extraDataMap
	}
	if !data.ScmBranch.IsNull() {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !data.JobType.IsNull() {
		bodyData.JobType = data.JobType.ValueString()
	}
	if !data.JobTags.IsNull() {
		bodyData.JobTags = data.JobTags.ValueString()
	}
	if !data.SkipTags.IsNull() {
		bodyData.SkipTags = data.SkipTags.ValueString()
	}
	if !data.Limit.IsNull() {
		bodyData.Limit = data.Limit.ValueString()
	}
	if !data.DiffMode.IsNull() {
		bodyData.DiffMode = data.DiffMode.ValueBool()
	}
	if !data.Verbosity.IsNull() {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}
	if !data.AllParentsMustConverge.IsNull() {
		bodyData.AllParentsMustConverge = data.AllParentsMustConverge.ValueBool()
	}
	if !data.Identifier.IsNull() {
		bodyData.Identifier = data.Identifier.ValueString()
	}

	url := "workflow_job_template_nodes/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues := []string{"id", "identifier"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))
	data.Identifier = types.StringValue(fmt.Sprintf("%v", returnedData["identifier"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowJobTemplatesJobNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowJobTemplatesJobNodeResourceModel

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

	url := fmt.Sprintf("workflow_job_template_nodes/%d/", id)
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

	var responseData WorkflowJobTemplateNodeAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable unmarshal response body into object",
			fmt.Sprintf("Error =  %v. ", err.Error()))
		return
	}

	if !data.WorkflowJobId.IsNull() || responseData.WorkflowJobId != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workflow_job_template_id"), responseData.WorkflowJobId)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.UnifiedJobTemplateId.IsNull() || responseData.UnifiedJobTemplateId != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("unified_job_template"), responseData.UnifiedJobTemplateId)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.Inventory.IsNull() || responseData.Inventory != nil {
		inventory, ok := responseData.Inventory.(float64)
		if !ok {
			resp.Diagnostics.AddError("read of inventory failed", fmt.Sprintf("unable to cast inventory %v to float64", responseData.Inventory))
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inventory"), int32(inventory))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	rawExtraData := responseData.ExtraData

	rawType := reflect.TypeOf(rawExtraData)

	if rawType.Kind() == reflect.Map {

		rawExtraData, ok := rawExtraData.(map[string]any)

		if !ok {
			resp.Diagnostics.AddError("unable to cast", "Unable to cast Extra Data as map[string]any")
			return
		}

		if len(rawExtraData) != 0 {
			tempMap := make(map[string]any, len(rawExtraData))
			for k, v := range rawExtraData {
				tempMap[k] = v
			}
			tempJson, err := json.Marshal(tempMap)
			if err != nil {
				resp.Diagnostics.AddError("marshall issue", "Unable to marshall extra data into json for storage.")
				return
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("extra_data"), string(tempJson))...)
			if resp.Diagnostics.HasError() {
				return
			}

		}

	}

	if rawType.Kind() == reflect.String {

		if extraDataProper, ok := responseData.ExtraData.(string); ok {
			if !data.ExtraData.IsNull() || (extraDataProper != "" && extraDataProper != "{}") {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("extra_data"), responseData.ExtraData)...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		} else {
			resp.Diagnostics.AddError("fail to cast any as string",
				"ExtraData couldnt be interpreted as a string.",
			)
			return
		}
	}

	if !data.ScmBranch.IsNull() || responseData.ScmBranch != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_branch"), responseData.ScmBranch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.JobType.IsNull() || responseData.JobType != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_type"), responseData.JobType)...)
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
	if !data.SkipTags.IsNull() || responseData.SkipTags != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("skip_tags"), responseData.SkipTags)...)
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

	diffModeType := reflect.TypeOf(responseData.DiffMode)

	if diffModeType != nil && diffModeType.Kind() == reflect.Bool {
		boolValue, ok := responseData.DiffMode.(bool)
		if !ok {
			resp.Diagnostics.AddError("unable to cast DiffMode as bool", "Unable to convert diff_mode to boolean")
			return
		}

		if !data.DiffMode.IsNull() && boolValue {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("diff_mode"), boolValue)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
	if !data.Verbosity.IsNull() || responseData.Verbosity != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("verbosity"), responseData.Verbosity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("all_parents_must_converge"), responseData.AllParentsMustConverge)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Identifier.IsNull() || responseData.Identifier != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("identifier"), responseData.Identifier)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

}

func (r *WorkflowJobTemplatesJobNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowJobTemplatesJobNodeResourceModel

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

	var bodyData WorkflowJobTemplateNodeAPIModel
	bodyData.WorkflowJobId = int(data.WorkflowJobId.ValueInt32())
	bodyData.UnifiedJobTemplateId = int(data.UnifiedJobTemplateId.ValueInt32())

	if !data.Inventory.IsNull() {
		bodyData.Inventory = int(data.Inventory.ValueInt32())
	}

	// Generate a go type that fits into the any var so that when ALL
	//  bodyData fields are set with go types, we call Marshall to generate entire JSON
	if !data.ExtraData.IsNull() {
		extraDataMap := new(map[string]any)
		err = json.Unmarshal([]byte(data.ExtraData.ValueString()), &extraDataMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable unmarshal map to json",
				fmt.Sprintf("Unable to convert id: %+v. ", data.ExtraData))
			return
		}
		bodyData.ExtraData = extraDataMap
	}

	bodyData.ScmBranch = data.ScmBranch.ValueString()
	bodyData.JobType = data.JobType.ValueString()
	bodyData.JobTags = data.JobTags.ValueString()
	bodyData.SkipTags = data.SkipTags.ValueString()
	bodyData.Limit = data.Limit.ValueString()
	bodyData.DiffMode = data.DiffMode.ValueBool()
	bodyData.Verbosity = int(data.Verbosity.ValueInt32())

	bodyData.AllParentsMustConverge = data.AllParentsMustConverge.ValueBool()
	bodyData.Identifier = data.Identifier.ValueString()

	url := fmt.Sprintf("workflow_job_template_nodes/%d/", id)
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues := []string{"identifier"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	data.Identifier = types.StringValue(fmt.Sprintf("%v", returnedData["identifier"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowJobTemplatesJobNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowJobTemplatesJobNodeResourceModel

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

	url := fmt.Sprintf("workflow_job_template_nodes/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *WorkflowJobTemplatesJobNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
