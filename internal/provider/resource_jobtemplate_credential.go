// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &JobTemplateCredentialResource{}
var _ resource.ResourceWithImportState = &JobTemplateCredentialResource{}

func NewJobTemplateCredentialResource() resource.Resource {
	return &JobTemplateCredentialResource{}
}

// JobTemplateCredentialResource defines the resource implementation.
type JobTemplateCredentialResource struct {
	client *AwxClient
}

// JobTemplateCredentialResourceModel describes the resource data model.
type JobTemplateCredentialResourceModel struct {
	JobTemplateId types.String `tfsdk:"job_template_id"`
	CredentialIds types.List   `tfsdk:"credential_ids"`
}

type JTCredentialAPIRead struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}

type Result struct {
	Id int `json:"id"`
}

func (r *JobTemplateCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jobtemplate_credential"
}

func (r *JobTemplateCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `The /api/v2/job_templates/{id}/credentials/ returns all credential objects associated to the template. But, when asked to associate a credential or \
                              dissassociate a credential, you must post a request once per credential ID.Therefore, I couldn't find a way to limit this resource to the 'one api call' \
                              principle. Instead, the terraform schema stores a list of associated credential ids. And, when creating or deleting or updated, it will make one api call PER \
                              list element. This allows the import function to work by only needing to pass in one job template ID to fill out the entire resource. If this was not done this way \
                              then when someone tries to to use the terraform plan -generate-config-out=./file.tf functionality it will create the resource block correctly. Otherwise, the \
                              -generate-config-out function would have to generate several resource blocks per tempalte id and it's not set up to do that, per my current awareness. As I'm writing this \
                              provider specifically so we can use the -generate-config-out option, I felt this was worth the price of breaking this principle. The downside seems to be that this means \
							  if one of the list element's api calls succeeds, but a subsequent list element's fails, the success of the first element's call is not magially un-done. \
							  So you'll perpas have to use refresh state functions in tf cli to resolve.`,
		Description: `The /api/v2/job_templates/{id}/credentials/ returns all credential objects associated to the template. But, when asked to associate a credential or \
					  dissassociate a credential, you must post a request once per credential ID.Therefore, I couldn't find a way to limit this resource to the 'one api call' \
					  principle. Instead, the terraform schema stores a list of associated credential ids. And, when creating or deleting or updated, it will make one api call PER \
					  list element. This allows the import function to work by only needing to pass in one job template ID to fill out the entire resource. If this was not done this way \
					  then when someone tries to to use the terraform plan -generate-config-out=./file.tf functionality it will create the resource block correctly. Otherwise, the \
					  -generate-config-out function would have to generate several resource blocks per tempalte id and it's not set up to do that, per my current awareness. As I'm writing this \
					  provider specifically so we can use the -generate-config-out option, I felt this was worth the price of breaking this principle. The downside seems to be that this means \
					  if one of the list element's api calls succeeds, but a subsequent list element's fails, the success of the first element's call is not magially un-done. \
					  So you'll perpas have to use refresh state functions in tf cli to resolve.`,
		Attributes: map[string]schema.Attribute{
			"job_template_id": schema.StringAttribute{
				Required:            true,
				Description:         "The ID of the containing Job Template.",
				MarkdownDescription: "The ID of the containing Job Template",
			},
			"credential_ids": schema.ListAttribute{
				Required:            true,
				Description:         "An ordered list of credential IDs associated to a particular Job Template.",
				MarkdownDescription: "An ordered list of credential IDs associated to a particular Job Template.",
				ElementType:         types.Int32Type,
			},
		},
	}
}

func (r *JobTemplateCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	configureData := req.ProviderData.(*AwxClient)

	r.client = configureData
}

func (r *JobTemplateCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobTemplateCredentialResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// set url for create HTTP request
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

	for _, val := range credIds {

		var bodyData Result
		bodyData.Id = val

		err := r.client.AssocJobTemplCredential(ctx, id, bodyData)
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate credential.", err.Error())
			return
		}
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobTemplateCredentialResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// set url for create HTTP request
	id, err := strconv.Atoi(data.JobTemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Converting ID to Int failed", fmt.Sprintf("Converting the job template id %s to int failed.", data.JobTemplateId.ValueString()))
		return
	}

	url := r.client.endpoint + fmt.Sprintf("/api/v2/job_templates/%d/credentials/", id)

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
	}
	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError(
			"Bad request status code.",
			fmt.Sprintf("Expected 200, got %v. ", httpResp.StatusCode))
		return
	}

	var responseData JTCredentialAPIRead

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

	tfCredIds := make([]int, 0, responseData.Count)

	for _, v := range responseData.Results {
		if data.CredentialIds.IsNull() {
			tfCredIds = append(tfCredIds, v.Id)
		} else {
			//todo
			return
		}
	}

	listValue, diags := types.ListValueFrom(ctx, types.Int32Type, tfCredIds)
	if diags.HasError() {
		return
	}

	data.CredentialIds = listValue

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Left intentinally "blank" (as initialized by clone of template scaffold) as these resources is replace by schema plan modifiers
func (r *JobTemplateCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JobTemplateCredentialResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobTemplateCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data JobTemplateCredentialResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// // set url for create HTTP request
	// id, err := strconv.Atoi(data.Id.ValueString())
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable convert id from string to int",
	// 		fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
	// }

	// url := r.client.endpoint + fmt.Sprintf("/api/v2/job_templates/%d/survey_spec", id)

	// // create HTTP request
	// httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to generate delete request",
	// 		fmt.Sprintf("Unable to gen url: %v. ", url))
	// }

	// httpReq.Header.Add("Content-Type", "application/json")
	// httpReq.Header.Add("Authorization", "Bearer"+" "+r.client.token)

	// httpResp, err := r.client.client.Do(httpReq)
	// if err != nil {
	// 	resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete got error: %s", err))
	// }
	// if httpResp.StatusCode != 200 {
	// 	resp.Diagnostics.AddError(
	// 		"Bad request status code.",
	// 		fmt.Sprintf("Expected 200, got %v. ", httpResp.StatusCode))

	// }
}

func (r *JobTemplateCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("job_template_id"), req, resp)
}