package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WorkflowJobTemplatesNodeAlwaysResource{}
var _ resource.ResourceWithImportState = &WorkflowJobTemplatesNodeAlwaysResource{}

func NewWorkflowJobTemplatesNodeAlwaysResource() resource.Resource {
	return &WorkflowJobTemplatesNodeAlwaysResource{}
}

type WorkflowJobTemplatesNodeAlwaysResource struct {
	client *AwxClient
}

type WorkflowJobTemplatesNodeAlwaysResourceModel struct {
	Id        types.String `tfsdk:"id"`
	AlwaysIds types.Set    `tfsdk:"always_ids"`
}

func (r *WorkflowJobTemplatesNodeAlwaysResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_job_template_node_always"
}

func (r *WorkflowJobTemplatesNodeAlwaysResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Specify a node ID and then a list of node IDs that should run when this one ends in success.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the containing workflow job template node.",
			},
			"always_ids": schema.SetAttribute{
				Required:    true,
				Description: "An unordered list of Node IDs attached to this workflow template node that should run on successful completion of this node.",
				ElementType: types.Int32Type,
			},
		},
	}
}

func (r *WorkflowJobTemplatesNodeAlwaysResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowJobTemplatesNodeAlwaysResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkflowJobTemplatesNodeAlwaysResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
	}

	url := fmt.Sprintf("workflow_job_template_nodes/%d/always_nodes/", id)

	var relatedIds []int

	diags := data.AlwaysIds.ElementsAs(ctx, &relatedIds, false)
	if diags.HasError() {
		return
	}

	for _, val := range relatedIds {

		var bodyData ChildAssocBody
		bodyData.Id = val
		bodyData.Associate = true

		_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate child.", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *WorkflowJobTemplatesNodeAlwaysResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkflowJobTemplatesNodeAlwaysResourceModel

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
	url := fmt.Sprintf("workflow_job_template_nodes/%d/always_nodes/", id)

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

	var responseData JTCredentialAPIRead

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable unmarshal response body into object",
			fmt.Sprintf("Error =  %v. ", err.Error()))
		return
	}

	tfRelatedIds := make([]int, 0, responseData.Count)

	for _, v := range responseData.Results {
		tfRelatedIds = append(tfRelatedIds, v.Id)
	}

	listValue, diags := types.SetValueFrom(ctx, types.Int32Type, tfRelatedIds)
	if diags.HasError() {
		return
	}
	data.AlwaysIds = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowJobTemplatesNodeAlwaysResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkflowJobTemplatesNodeAlwaysResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Converting ID to Int failed", fmt.Sprintf("Converting the job template id %s to int failed.", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("workflow_job_template_nodes/%d/always_nodes/", id)

	body, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	var responseData JTChildAPIRead

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable unmarshal response body into object",
			fmt.Sprintf("Error =  %v. ", err.Error()))
		return
	}

	ApiTfChildIds := make([]int, 0, responseData.Count)

	for _, v := range responseData.Results {
		ApiTfChildIds = append(ApiTfChildIds, v.Id)
	}

	var PlanChildIds []int
	diags := data.AlwaysIds.ElementsAs(ctx, &PlanChildIds, false)
	if diags.HasError() {
		return
	}

	// diassociate any chyildren found currently via API call that
	//  are no longer in the plan
	for _, v := range ApiTfChildIds {
		if !slices.Contains(PlanChildIds, v) {
			var bodyData ChildDissasocBody
			bodyData.Id = v

			_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
			if err != nil {
				resp.Diagnostics.AddError("Failed to disassociate child.", err.Error())
				return
			}
		}
	}
	// associate any children found in plan that weren't shown in API response
	for _, v := range PlanChildIds {
		if !slices.Contains(ApiTfChildIds, v) {
			var bodyData ChildAssocBody
			bodyData.Id = v
			bodyData.Associate = true

			_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
			if err != nil {
				resp.Diagnostics.AddError("Failed to associate child.", err.Error())
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowJobTemplatesNodeAlwaysResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkflowJobTemplatesNodeAlwaysResourceModel

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

	url := fmt.Sprintf("workflow_job_template_nodes/%d/always_nodes/", id)

	var RelatedIds []int

	diags := data.AlwaysIds.ElementsAs(ctx, &RelatedIds, false)
	if diags.HasError() {
		return
	}

	for _, val := range RelatedIds {

		var bodyData ChildDissasocBody

		bodyData.Id = val
		bodyData.Disassociate = true

		_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
		if err != nil {
			resp.Diagnostics.AddError("Failed to disassociate child.", err.Error())
			return
		}
	}
}

func (r *WorkflowJobTemplatesNodeAlwaysResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
