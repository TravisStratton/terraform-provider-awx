package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var _ resource.Resource = &GroupHostResource{}

// var _ resource.ResourceWithImportState = &GroupHostResource{}

func NewGroupHostResource() resource.Resource {
	return &GroupHostResource{}
}

type GroupHostResource struct {
	client *providerClient
}

func (r *GroupHostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_host"
}

func (r *GroupHostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Add an existing host to an existing inventory group.`,
		Attributes: map[string]schema.Attribute{
			"group_id": schema.StringAttribute{
				Description: "Group ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_id": schema.StringAttribute{
				Description: "Host ID to add to the group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *GroupHostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupHostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupHostModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData GroupHostAssocAPIModel

	hostId, err := strconv.Atoi(data.HostId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert host id to int.", fmt.Sprintf("Unable to convert %v to int", data.HostId.ValueString()))
		return
	}

	bodyData.Id = hostId

	groupId, err := strconv.Atoi(data.GroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert group id to int.", fmt.Sprintf("Unable to convert %v to int", data.GroupId.ValueString()))
		return
	}

	url := fmt.Sprintf("groups/%d/hosts/", groupId)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *GroupHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupHostModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostId, err := strconv.Atoi(data.HostId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert host id to int in read.", fmt.Sprintf("Unable to convert %v to int", data.HostId.ValueString()))
		return
	}

	groupId, err := strconv.Atoi(data.GroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert group id to int in read.", fmt.Sprintf("Unable to convert %v to int", data.GroupId.ValueString()))
		return
	}

	url := fmt.Sprintf("groups/%d/hosts/?id=%d", groupId, hostId)
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

	result := struct {
		Count int `json:"count"`
	}{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into object",
			fmt.Sprintf("Error:  %v.", err.Error()))
		return
	}

	if result.Count == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	if result.Count != 1 {

		resp.Diagnostics.AddError(
			"Incorrect number of group hosts returned by Ids",
			fmt.Sprintf("Unable to read groups/id/hosts/?id=x as API returned %v results.", result.Count))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupHostModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// NOTE
	// because we have this resource scheme setup to require replace, the update method is intentially bare-minimum

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var bodyData ChildDissasocBody
	var data GroupHostModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostId, err := strconv.Atoi(data.HostId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert host id to int in Delete.", fmt.Sprintf("Unable to convert %v to int", data.HostId.ValueString()))
		return
	}

	groupId, err := strconv.Atoi(data.GroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert group id to int in Delete.", fmt.Sprintf("Unable to convert %v to int", data.GroupId.ValueString()))
		return
	}

	url := fmt.Sprintf("groups/%d/hosts/", groupId)

	bodyData.Id = hostId
	bodyData.Disassociate = true

	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

// see https://developer.hashicorp.com/terraform/plugin/framework/resources/import
// and
// https://developer.hashicorp.com/terraform/plugin/framework/resources/identity#importing-by-identity
// if we end up needing to add support for import to this resource type.
//
// Also, be sure to uncomment out the line at the top of this file related to Import if you implmeent import.
//
// This is not like imports we've done in this provider for other resources as there is no unique ID in the
// the system to designate this resources. Instead, it's a combo of the groupID & the HostID.

// func (r *GroupHostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
// }
