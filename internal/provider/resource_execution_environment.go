package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ExecutionEnvironmentResource{}
var _ resource.ResourceWithImportState = &ExecutionEnvironmentResource{}

func NewExecutionEnvironmentResource() resource.Resource {
	return &ExecutionEnvironmentResource{}
}

type ExecutionEnvironmentResource struct {
	client *AwxClient
}

func (r *ExecutionEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_execution_environment"
}

func (r *ExecutionEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an AWX Execution Environment.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Execution Environment ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Execution Environment name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Execution Environment description.",
				Optional:    true,
			},
			"image": schema.StringAttribute{
				Description: "The full image location, including the container registry, image name, and version tag.",
				Required:    true,
			},
			"pull": schema.StringAttribute{
				Description: "`always` always pull container before running, `missing` only pull the image if not pressent before running, `never` never pull container before running.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"always", "missing", "never"}...),
				},
			},
			"organization": schema.Int32Attribute{
				Description: "Leave this field blank to make the execution environment globally available.",
				Optional:    true,
			},
			"credential": schema.Int32Attribute{
				Description: "Credential to authenticate with a protected container registry.",
				Optional:    true,
			},
		},
	}
}

func (r *ExecutionEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExecutionEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExecutionEnvironmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData ExecutionEnvironmentAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.Image = data.Image.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.Pull.IsNull()) {
		bodyData.Pull = data.Pull.ValueString()
	}
	if !(data.Organization.IsNull()) {
		bodyData.Organization = int(data.Organization.ValueInt32())
	}
	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}

	url := "execution_environments/"
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

func (r *ExecutionEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ExecutionEnvironmentModel

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

	url := fmt.Sprintf("execution_environments/%d/", id)
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

	var responseData ExecutionEnvironmentAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("image"), responseData.Image)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Pull.IsNull() || responseData.Pull != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pull"), responseData.Pull)...)
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

	if !data.Credential.IsNull() || responseData.Credential != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("credential"), responseData.Credential)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *ExecutionEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExecutionEnvironmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data))
		return
	}

	var bodyData ExecutionEnvironmentAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.Image = data.Image.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.Pull.IsNull()) {
		bodyData.Pull = data.Pull.ValueString()
	}
	if !(data.Organization.IsNull()) {
		bodyData.Organization = int(data.Organization.ValueInt32())
	}
	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}

	url := fmt.Sprintf("execution_environments/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExecutionEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ExecutionEnvironmentModel

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

	url := fmt.Sprintf("execution_environments/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *ExecutionEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
