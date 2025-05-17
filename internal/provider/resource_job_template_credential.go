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

var _ resource.Resource = &JobTemplateCredentialResource{}
var _ resource.ResourceWithImportState = &JobTemplateCredentialResource{}

func NewJobTemplateCredentialResource() resource.Resource {
	return &JobTemplateCredentialResource{}
}

type JobTemplateCredentialResource struct {
	client *AwxClient
}

type JobTemplateCredentialResourceModel struct {
	JobTemplateId types.String `tfsdk:"job_template_id"`
	CredentialIds types.Set    `tfsdk:"credential_ids"`
}

type JTCredentialAPIRead struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}

type Result struct {
	Id int `json:"id"`
}

type DissasocBody struct {
	Id           int  `json:"id"`
	Disassociate bool `json:"disassociate"`
}

func (r *JobTemplateCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_template_credential"
}

func (r *JobTemplateCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Associate credentials to a job template.",
		Attributes: map[string]schema.Attribute{
			"job_template_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the containing Job Template.",
			},
			"credential_ids": schema.SetAttribute{
				Required:    true,
				Description: "An unordered list of credential IDs associated to a particular Job Template.",
				ElementType: types.Int32Type,
			},
		},
	}
}

func (r *JobTemplateCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobTemplateCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobTemplateCredentialResourceModel

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

	url := fmt.Sprintf("job_templates/%d/credentials/", id)

	var credIds []int

	diags := data.CredentialIds.ElementsAs(ctx, &credIds, false)
	if diags.HasError() {
		return
	}

	for _, val := range credIds {

		var bodyData Result
		bodyData.Id = val

		_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204})
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate child.", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobTemplateCredentialResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.JobTemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Converting ID to Int failed", fmt.Sprintf("Converting the job template id %s to int failed.", data.JobTemplateId.ValueString()))
		return
	}

	url := fmt.Sprintf("job_templates/%d/credentials/", id)

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

	var responseData JTCredentialAPIRead

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable unmarshal response body into object",
			fmt.Sprintf("Error =  %v. ", err.Error()))
		return
	}

	tfCredIds := make([]int, 0, responseData.Count)

	for _, v := range responseData.Results {
		tfCredIds = append(tfCredIds, v.Id)
	}

	listValue, diags := types.SetValueFrom(ctx, types.Int32Type, tfCredIds)
	if diags.HasError() {
		return
	}
	data.CredentialIds = listValue
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JobTemplateCredentialResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.JobTemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Converting ID to Int failed", fmt.Sprintf("Converting the job template id %s to int failed.", data.JobTemplateId.ValueString()))
		return
	}

	url := fmt.Sprintf("job_templates/%d/credentials/", id)

	body, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
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

	ApiTfCredIds := make([]int, 0, responseData.Count)

	for _, v := range responseData.Results {
		ApiTfCredIds = append(ApiTfCredIds, v.Id)
	}

	var PlanCredIds []int
	diags := data.CredentialIds.ElementsAs(ctx, &PlanCredIds, false)
	if diags.HasError() {
		return
	}

	// diassociate any credentials found currently via API call that
	//  are no longer in the plan
	for _, v := range ApiTfCredIds {
		if !slices.Contains(PlanCredIds, v) {
			var bodyData DissasocBody
			bodyData.Id = v

			_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204})
			if err != nil {
				resp.Diagnostics.AddError("Failed to disassociate child.", err.Error())
				return
			}
		}
	}
	// associate any credentials found in plan that weren't shown in API response
	for _, v := range PlanCredIds {
		if !slices.Contains(ApiTfCredIds, v) {
			var bodyData Result
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

func (r *JobTemplateCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data JobTemplateCredentialResourceModel

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

	var credIds []int

	diags := data.CredentialIds.ElementsAs(ctx, &credIds, false)
	if diags.HasError() {
		return
	}

	url := fmt.Sprintf("job_templates/%d/credentials/", id)

	for _, val := range credIds {

		var bodyData DissasocBody

		bodyData.Id = val
		bodyData.Disassociate = true

		_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204})
		if err != nil {
			resp.Diagnostics.AddError("Failed to disassociate child.", err.Error())
			return
		}

	}

}

func (r *JobTemplateCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("job_template_id"), req, resp)
}
