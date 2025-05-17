package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CredentialTypeResource{}
var _ resource.ResourceWithImportState = &CredentialTypeResource{}

func NewCredentialTypeResource() resource.Resource {
	return &CredentialTypeResource{}
}

type CredentialTypeResource struct {
	client *AwxClient
}

func (r *CredentialTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential_type"
}

func (r *CredentialTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an AWX credential type.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential type ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Credential type name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Credential type description.",
				Optional:    true,
			},
			"inputs": schema.StringAttribute{
				Optional:    true,
				Description: "Enter inputs using JSON syntax wrapped with `jsonencode()`. Refer to the Ansible Controller documentation for example syntax. Default value is `\"---\"`",
			},
			"injectors": schema.StringAttribute{
				Optional:    true,
				Description: "Enter injectors using either JSON syntax with `jsonencode()`. Refer to the Ansible Controller documentation for example syntax. Default value is `\"---\"`",
			},
			"kind": schema.StringAttribute{
				Default:     stringdefault.StaticString("cloud"),
				Optional:    true,
				Computed:    true,
				Description: "Either `cloud` or `net` but probably `cloud`.",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"cloud", "net"}...),
				},
			},
		},
	}
}

func (r *CredentialTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CredentialTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CredentialTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData CredentialTypeAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !data.Inputs.IsNull() && data.Inputs.ValueString() != "" {
		inputsDataMap := new(map[string]any)
		err := json.Unmarshal([]byte(data.Inputs.ValueString()), &inputsDataMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal map to json",
				fmt.Sprintf("Unable to process inputs: %+v. ", data.Inputs))
			return
		}

		bodyData.Inputs = inputsDataMap
	}
	if !data.Injectors.IsNull() && data.Injectors.ValueString() != "" {
		injectorsDataMap := new(map[string]any)
		err := json.Unmarshal([]byte(data.Injectors.ValueString()), &injectorsDataMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal map to json",
				fmt.Sprintf("Unable to process injectors: %+v. ", data.Injectors))
			return
		}

		bodyData.Injectors = injectorsDataMap
	}
	if !(data.Kind.IsNull()) {
		bodyData.Kind = data.Kind.ValueString()
	}

	url := "credential_types/"
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

func (r *CredentialTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CredentialTypeModel

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

	url := fmt.Sprintf("credential_types/%d/", id)
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

	var responseData CredentialTypeAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
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

	if !data.Inputs.IsNull() || responseData.Inputs != "" {
		rawInputs := responseData.Inputs
		rawInputsType := reflect.TypeOf(rawInputs)

		if rawInputsType.Kind() == reflect.Map {
			rawInputs, ok := rawInputs.(map[string]any)
			if !ok {
				resp.Diagnostics.AddError("Unable to cast", "Unable to cast Inputs as map[string]any")
				return
			}

			if len(rawInputs) != 0 {
				tmpInputsMap := make(map[string]any, len(rawInputs))
				for k, v := range rawInputs {
					tmpInputsMap[k] = v
				}
				tmpInputsJson, err := json.Marshal(tmpInputsMap)
				if err != nil {
					resp.Diagnostics.AddError("marshall issue", "Unable to marshall Inputs into json for storage.")
					return
				}
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inputs"), string(tmpInputsJson))...)
				if resp.Diagnostics.HasError() {
					resp.Diagnostics.AddError("String issue.", "Unable to convert Inputs json to string and storage.")
					return
				}
			}
		}
	}

	if !data.Injectors.IsNull() || responseData.Injectors != "" {
		rawInjectors := responseData.Injectors
		rawInjectorsType := reflect.TypeOf(rawInjectors)
		if rawInjectorsType.Kind() == reflect.Map {
			rawInjectors, ok := rawInjectors.(map[string]any)
			if !ok {
				resp.Diagnostics.AddError("Unable to cast", "Unable to cast Injectors as map[string]any")
				return
			}

			if len(rawInjectors) != 0 {
				tmpInjectorsMap := make(map[string]any, len(rawInjectors))
				for k, v := range rawInjectors {
					tmpInjectorsMap[k] = v
				}
				tmpInjectorsJson, err := json.Marshal(tmpInjectorsMap)
				if err != nil {
					resp.Diagnostics.AddError("marshall issue", "Unable to marshall Injectors into json for storage.")
					return
				}
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("injectors"), string(tmpInjectorsJson))...)
				if resp.Diagnostics.HasError() {
					resp.Diagnostics.AddError("String issue.", "Unable to convert Injectors json to string and storage.")
					return
				}
			}
		}
	}

	if !data.Kind.IsNull() || responseData.Kind != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kind"), responseData.Kind)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *CredentialTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CredentialTypeModel

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

	var bodyData CredentialTypeAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.Inputs.IsNull()) {
		inputsMap := new(map[string]any)
		err = json.Unmarshal([]byte(data.Inputs.ValueString()), &inputsMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable unmarshal map to json",
				fmt.Sprintf("Unable to convert id: %+v. ", data.Inputs))
			return
		}
		bodyData.Inputs = inputsMap
	}
	if !(data.Injectors.IsNull()) {
		injectorsMap := new(map[string]any)
		err = json.Unmarshal([]byte(data.Injectors.ValueString()), &injectorsMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable unmarshal map to json",
				fmt.Sprintf("Unable to convert id: %+v. ", data.Injectors))
			return
		}
		bodyData.Injectors = injectorsMap
	}
	if !(data.Kind.IsNull()) {
		bodyData.Kind = data.Kind.ValueString()
	}

	url := fmt.Sprintf("credential_types/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CredentialTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CredentialTypeModel

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

	url := fmt.Sprintf("credential_types/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *CredentialTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
