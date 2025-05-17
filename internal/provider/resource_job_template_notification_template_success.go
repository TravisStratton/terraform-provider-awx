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

var _ resource.Resource = &JobTemplateNotifTemplSuccessResource{}
var _ resource.ResourceWithImportState = &JobTemplateNotifTemplSuccessResource{}

func NewJobTemplateNotifTemplSuccessResource() resource.Resource {
	return &JobTemplateNotifTemplSuccessResource{}
}

type JobTemplateNotifTemplSuccessResource struct {
	client *AwxClient
}

type JobTemplateNotifTemplSuccessResourceModel struct {
	JobTemplateId    types.String `tfsdk:"job_template_id"`
	NotifTEmplateIDs types.Set    `tfsdk:"notif_template_ids"`
}

func (r *JobTemplateNotifTemplSuccessResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_template_notification_template_success"
}

func (r *JobTemplateNotifTemplSuccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Associate awx_notification_template(s) to a job template.",
		Attributes: map[string]schema.Attribute{
			"job_template_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the containing Job Template.",
			},
			"notif_template_ids": schema.SetAttribute{
				Required:    true,
				Description: "An unordered list of `awx_notification_template` IDs associated to a particular Job Template.",
				ElementType: types.Int32Type,
			},
		},
	}
}

func (r *JobTemplateNotifTemplSuccessResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobTemplateNotifTemplSuccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobTemplateNotifTemplSuccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.JobTemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.JobTemplateId.ValueString()))
	}

	url := fmt.Sprintf("job_templates/%d/notification_templates_success/", id)

	var relatedIds []int

	diags := data.NotifTEmplateIDs.ElementsAs(ctx, &relatedIds, false)
	if diags.HasError() {
		return
	}

	for _, val := range relatedIds {

		var bodyData ChildResult
		bodyData.Id = val

		_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204})
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate child.", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateNotifTemplSuccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobTemplateNotifTemplSuccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.JobTemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Converting ID to Int failed", fmt.Sprintf("Converting the job template id %s to int failed.", data.JobTemplateId.ValueString()))
		return
	}

	url := fmt.Sprintf("job_templates/%d/notification_templates_success/", id)

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

	var responseData JTChildAPIRead

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
	data.NotifTEmplateIDs = listValue
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateNotifTemplSuccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JobTemplateNotifTemplSuccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.JobTemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Converting ID to Int failed", fmt.Sprintf("Converting the job template id %s to int failed.", data.JobTemplateId.ValueString()))
		return
	}

	url := fmt.Sprintf("job_templates/%d/notification_templates_success/", id)

	body, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200})
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
	diags := data.NotifTEmplateIDs.ElementsAs(ctx, &PlanChildIds, false)
	if diags.HasError() {
		return
	}

	// diassociate any chyildren found currently via API call that
	//  are no longer in the plan
	for _, v := range ApiTfChildIds {
		if !slices.Contains(PlanChildIds, v) {
			var bodyData ChildDissasocBody
			bodyData.Id = v

			_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204})
			if err != nil {
				resp.Diagnostics.AddError("Failed to disassociate child.", err.Error())
				return
			}
		}
	}
	// associate any children found in plan that weren't shown in API response
	for _, v := range PlanChildIds {
		if !slices.Contains(ApiTfChildIds, v) {
			var bodyData ChildResult
			bodyData.Id = v

			_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204})
			if err != nil {
				resp.Diagnostics.AddError("Failed to associate child.", err.Error())
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateNotifTemplSuccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data JobTemplateNotifTemplSuccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.JobTemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.JobTemplateId.ValueString()))
	}

	url := fmt.Sprintf("job_templates/%d/notification_templates_success/", id)

	var RelatedIds []int

	diags := data.NotifTEmplateIDs.ElementsAs(ctx, &RelatedIds, false)
	if diags.HasError() {
		return
	}

	for _, val := range RelatedIds {

		var bodyData ChildDissasocBody

		bodyData.Id = val
		bodyData.Disassociate = true

		_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204})
		if err != nil {
			resp.Diagnostics.AddError("Failed to disassociate child.", err.Error())
			return
		}
	}

}

func (r *JobTemplateNotifTemplSuccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("job_template_id"), req, resp)
}
