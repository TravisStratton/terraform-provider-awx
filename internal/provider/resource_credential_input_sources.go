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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CredentialInputSourcesResource{}
var _ resource.ResourceWithImportState = &CredentialInputSourcesResource{}

func NewCredentialInputSourcesResource() resource.Resource {
	return &CredentialInputSourcesResource{}
}

type CredentialInputSourcesResource struct {
	client *AwxClient
}

func (r *CredentialInputSourcesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential_input_sources"
}

func (r *CredentialInputSourcesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage Credential Input Sources.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential Input Source ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Credential input source description.",
				Optional:    true,
			},
			"input_field_name": schema.StringAttribute{
				Description: "Credential input source input field name.",
				Required:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Credential input source metadata",
				Required:    true,
				ElementType: types.StringType,
			},
			"target_credential": schema.Int32Attribute{
				Description: "ID of credential to which this input source should be attached.",
				Required:    true,
			},
			"source_credential": schema.Int32Attribute{
				Description: "ID of the source credential.",
				Required:    true,
			},
		},
	}
}

func (r *CredentialInputSourcesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*AwxClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AwxClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = configureData
}

func (r *CredentialInputSourcesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CredentialInputSourcesModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData CredentialInputSourcesAPIModel

	bodyData.InputFieldName = data.InputFieldName.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	bodyData.SourceCredential = int(data.SourceCredential.ValueInt32())
	bodyData.TargetCredential = int(data.TargetCredential.ValueInt32())

	var metadataMap map[string]string

	resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadataMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bodyData.Metadata = metadataMap

	url := "credential_input_sources/"
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

func (r *CredentialInputSourcesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CredentialInputSourcesModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id))
		return
	}

	url := fmt.Sprintf("credential_input_sources/%d/", id)
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

	var responseData CredentialInputSourcesAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into result object",
			fmt.Sprintf("Error:  %v.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("input_field_name"), responseData.InputFieldName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("metadata"), responseData.Metadata)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_credential"), responseData.TargetCredential)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_credential"), responseData.SourceCredential)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

}

func (r *CredentialInputSourcesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CredentialInputSourcesModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData CredentialInputSourcesAPIModel

	bodyData.InputFieldName = data.InputFieldName.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	bodyData.SourceCredential = int(data.SourceCredential.ValueInt32())
	bodyData.SourceCredential = int(data.TargetCredential.ValueInt32())

	var metadataMap map[string]string

	resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadataMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bodyData.Metadata = metadataMap

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id))
		return
	}

	url := fmt.Sprintf("credential_input_sources/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *CredentialInputSourcesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CredentialInputSourcesModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("credential_input_sources/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

}

func (r *CredentialInputSourcesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
