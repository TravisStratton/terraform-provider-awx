// SPECIAL: This file may require repo or controller-specific things.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &OrganizationResource{}
var _ resource.ResourceWithImportState = &OrganizationResource{}

func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

type OrganizationResource struct {
	client *providerClient
}

func (r *OrganizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *OrganizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an Automation Controller organization.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aap25_gateway_id": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the organization.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Organization description.",
			},
			"default_environment": schema.Int32Attribute{
				Optional:    true,
				Description: "AWX/AAP2.4 only. The fallback execution environment that will be used for jobs inside of this organization if not explicitly assigned at the project, job template or workflow level.",
			},
			"max_hosts": schema.Int32Attribute{
				Optional:    true,
				Description: "AWX/AAP2.4 only Maximum number of hosts allowed to be managed by this organization.",
				Default:     int32default.StaticInt32(0),
				Computed:    true,
			},
		},
	}
}

func (r OrganizationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Disallow default_environment for >AAP2.5
	platform, ok := os.LookupEnv("TOWER_PLATFORM")
	if !ok {
		return
	}

	if platform == "awx" || platform == "aap2.4" || data.DefaultEnv.IsNull() {
	} else {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_environment"),
			"Invalid Attribute Configuration",
			"Attribute default_environment is not supported in this version of the provider.",
		)
		return
	}

	if platform == "awx" || platform == "aap2.4" || data.MaxHosts.IsNull() || data.MaxHosts.ValueInt32() == 0 {
		return
	} else {
		resp.Diagnostics.AddAttributeError(
			path.Root("max_hosts"),
			"Invalid Attribute Configuration",
			"Attribute max_hosts is not supported in this version of the provider.",
		)
		return
	}
}

func (r *OrganizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData OrganizationAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.DefaultEnv.IsNull()) {
		bodyData.DefaultEnv = int(data.DefaultEnv.ValueInt32())
	}
	if !(data.MaxHosts.IsNull()) {
		bodyData.MaxHosts = int(data.MaxHosts.ValueInt32())
	}

	url := "organizations/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	// organizations must be created using the /gateway/ instead of /controller/ api endpoint. But,
	//  the same org may get 2 different IDs and we need the ID from the controller in order to use
	//  the organization ID.
	id, ok := returnedData["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError(
			"unable to cast ID as float64",
			fmt.Sprintf("Value provided was: %v.", returnedData["id"]))
		return
	}
	data.Aap25GatewayId = types.Int32Value(int32(id))

	if r.client.platform == "aap2.5" {

		// overwrite returnedData with Get against org's /controller/ endpoint

		url := fmt.Sprintf("organizations/?name=%s", data.Name.ValueString())
		responseBodyData, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "controller")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error making API http request",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}
		// get id
		var nameResult JTChildAPIRead
		err = json.Unmarshal(responseBodyData, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into result object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count != 1 {
			resp.Diagnostics.AddError(
				"Org controller result count not 1.",
				fmt.Sprintf("Querying for org by name against controller endpoint resulted in result count of %d isntead of 1.", nameResult.Count))
			return
		}
		data.Id = types.StringValue(strconv.Itoa(nameResult.Results[0].Id))

	} else {
		data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationModel

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

	url := fmt.Sprintf("organizations/%d/", id)
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

	var responseData OrganizationAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	// if aap2.5 get the /gateway/ id and set the related field
	if r.client.platform == "aap2.5" {

		url := fmt.Sprintf("organizations/?name=%s", responseData.Name)
		responseBodyData, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "gateway")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error making API http request",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}

		var nameResult JTChildAPIRead
		err = json.Unmarshal(responseBodyData, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into result object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count != 1 {
			resp.Diagnostics.AddError(
				"Expected only one org result from gateway",
				fmt.Sprintf("Got count of %d instead.", nameResult.Count))
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aap25_gateway_id"), nameResult.Results[0].Id)...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		var id int
		var err error

		id, err = strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("can't convert Id to int", "unable to convert ID to int.")
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aap25_gateway_id"), id)...)
		if resp.Diagnostics.HasError() {
			return
		}
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

	if !data.DefaultEnv.IsNull() || responseData.DefaultEnv != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("default_environment"), responseData.DefaultEnv)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("max_hosts"), responseData.MaxHosts)...)
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	var err error

	if r.client.platform == "aap2.5" {
		id = int(data.Aap25GatewayId.ValueInt32())
	} else {
		id, err = strconv.Atoi(data.Id.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id))
		return
	}

	var bodyData OrganizationAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.DefaultEnv.IsNull()) {
		bodyData.DefaultEnv = int(data.DefaultEnv.ValueInt32())
	}
	if !(data.MaxHosts.IsNull()) {
		bodyData.MaxHosts = int(data.MaxHosts.ValueInt32())
	}

	url := fmt.Sprintf("organizations/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	var err error

	if r.client.platform == "aap2.5" {
		id = int(data.Aap25GatewayId.ValueInt32())
	} else {
		id, err = strconv.Atoi(data.Id.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
		return
	}
	url := fmt.Sprintf("organizations/%d/", id)

	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
