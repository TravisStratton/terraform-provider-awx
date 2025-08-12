package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WorkflowJobTemplateApprovalNode{}
var _ resource.ResourceWithImportState = &WorkflowJobTemplateApprovalNode{}

func NewWorkflowJobTemplateApprovalNodeResource() resource.Resource {
	return &WorkflowJobTemplateApprovalNode{}
}

type WorkflowJobTemplateApprovalNode struct {
	client *providerClient
}

type WorkflowJobTemplateApprovalNodeModel struct {
	Id                    types.String `tfsdk:"id"`
	ApprovalTemplateId    types.Int32  `tfsdk:"approval_template_id"`
	WorkflowJobTemplateId types.Int32  `tfsdk:"workflow_job_template_id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Timeout               types.Int32  `tfsdk:"timeout"`
}

type WorkflowJobTmplNodeApprvCreateAPIModel struct {
	Description any    `json:"description,omitempty"`
	Name        string `json:"name"`
	Timeout     any    `json:"timeout,omitempty"`
}

func (r *WorkflowJobTemplateApprovalNode) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_job_template_approval_node"
}

func (r *WorkflowJobTemplateApprovalNode) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "To add an approval node/step to an existing workflow job template a new node will be created and a small template will be added to that node. You'll need to create an " + configprefix.Prefix + "_workflow_job_template_node_success resource (or always/failure) in order to make this approval node execute in the sequence you want your visualizer nodes to run.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique ID for this node.",
			},
			"approval_template_id": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Description: "The unique ID for this approval template.",
			},
			"workflow_job_template_id": schema.Int32Attribute{
				Required:    true,
				Description: "This the ID for the `Automation Controller_workflow_job_template` for which this approval node/template should be embedded.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of this template.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The name of this template.",
			},
			"timeout": schema.Int32Attribute{
				Optional:    true,
				Description: "The number of seconds for timeout.",
				Default:     int32default.StaticInt32(0),
				Computed:    true,
			},
		},
	}
}

func (r *WorkflowJobTemplateApprovalNode) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowJobTemplateApprovalNode) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowJobTemplateApprovalNodeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	/////////////////////////////////////////////////////
	// First create an empty node

	url := fmt.Sprintf("workflow_job_templates/%d/workflow_nodes/", data.WorkflowJobTemplateId.ValueInt32())

	newJTworkflowNode := struct {
		NodeType string `json:"node_type"`
	}{
		NodeType: "approval",
	}

	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, newJTworkflowNode, []int{201}, "")
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

	////////////////////////////////////////////////////
	// Now we create a new approval template on that node
	////////////////////////////////////////////////////
	var bodyData WorkflowJobTmplNodeApprvCreateAPIModel

	bodyData.Name = data.Name.ValueString()
	if !data.Description.IsNull() {
		bodyData.Description = data.Description.ValueString()
	}
	if !data.Timeout.IsNull() {
		bodyData.Timeout = int(data.Timeout.ValueInt32())
	}

	tempId := fmt.Sprintf("%v", returnedData["id"])
	tempIdInt, err := strconv.Atoi(tempId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting returned ID to an integer",
			fmt.Sprintf("Error converting %v to integer, with error %v", tempId, err.Error()),
		)
		return
	}

	url = fmt.Sprintf("workflow_job_template_nodes/%d/create_approval_template/", tempIdInt)

	returnedData, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues = []string{"id"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	tempId = fmt.Sprintf("%v", returnedData["id"])
	tempIdInt, err = strconv.Atoi(tempId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting returned ID to an integer",
			fmt.Sprintf("Error converting %v to integer, with error %v", tempId, err.Error()),
		)
		return
	}

	data.ApprovalTemplateId = types.Int32Value(int32(tempIdInt))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowJobTemplateApprovalNode) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowJobTemplateApprovalNodeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	/// read the node's workflow template ID first
	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}
	url := fmt.Sprintf("workflow_job_template_nodes/%d/", id)

	responseBody, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "")
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

	getNameFromResponse := struct {
		WorkflowJobTemplateId int `json:"workflow_job_template"`
		ApprovalTemplateId    int `json:"unified_job_template"`
	}{}

	err = json.Unmarshal(responseBody, &getNameFromResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get unmarshal http response",
			fmt.Sprintf("error was %v", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workflow_job_template_id"), getNameFromResponse.WorkflowJobTemplateId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("approval_template_id"), getNameFromResponse.ApprovalTemplateId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	/// now read the node's template's data

	url = fmt.Sprintf("workflow_approval_templates/%d/", getNameFromResponse.ApprovalTemplateId)

	responseBody, statusCode, err = r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "")
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

	var readAPIResponse WorkflowJobTmplNodeApprvCreateAPIModel

	err = json.Unmarshal(responseBody, &readAPIResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get unmarshal http response",
			fmt.Sprintf("error was %v", err))
		return
	}

	if !data.Name.IsNull() || readAPIResponse.Name != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), readAPIResponse.Name)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Description.IsNull() || readAPIResponse.Description != "" {
		descrString, ok := readAPIResponse.Description.(string)
		if !ok {
			resp.Diagnostics.AddError("couldn't convert any to string", "unable to convert any to string.")
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), descrString)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Timeout.IsNull() || readAPIResponse.Timeout != nil {
		timeout, ok := readAPIResponse.Timeout.(float64)
		if !ok {
			resp.Diagnostics.AddError("couldn't convert any to float64", "unable to convert any to float64.")
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeout"), int(timeout))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

}

func (r *WorkflowJobTemplateApprovalNode) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowJobTemplateApprovalNodeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData WorkflowJobTmplNodeApprvCreateAPIModel

	bodyData.Name = data.Name.ValueString()

	bodyData.Description = data.Description.ValueString()
	bodyData.Timeout = int(data.Timeout.ValueInt32())

	url := fmt.Sprintf("workflow_approval_templates/%d/", data.ApprovalTemplateId.ValueInt32())

	_, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, url, bodyData, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowJobTemplateApprovalNode) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowJobTemplateApprovalNodeModel

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

	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *WorkflowJobTemplateApprovalNode) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
