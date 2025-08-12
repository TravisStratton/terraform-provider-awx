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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &InstanceGroupResource{}
var _ resource.ResourceWithImportState = &InstanceGroupResource{}

func NewInstanceGroupResource() resource.Resource {
	return &InstanceGroupResource{}
}

type InstanceGroupResource struct {
	client *providerClient
}

func (r *InstanceGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_group"
}

func (r *InstanceGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an Automation Controller instance_group.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "InstanceGroup ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Instance Group name.",
				Required:    true,
			},
			"credential": schema.Int32Attribute{
				Description: "Credential ID to authenticate with Kubernetes or OpenShift for ContainerGroups. Must be of type `OpenShift` or `Kubernetes API Bearer Token`",
				Optional:    true,
			},
			"is_container_group": schema.BoolAttribute{
				Description: "Signifies that this InstanceGroup should act as a ContainerGroup. If no credential is specified, the underlying Pod’s ServiceAccount will be used.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"max_concurrent_jobs": schema.Int32Attribute{
				Description: "Maximum number of jobs to run concurrently on this group. Zero means no limit will be enforced.",
				Computed:    true,
				Optional:    true,
				Default:     int32default.StaticInt32(0),
			},
			"max_forks": schema.Int32Attribute{
				Description: "Maximum number of forks to allow across all jobs running concurrently on this group. Zero means no limit will be enforced.",
				Computed:    true,
				Optional:    true,
				Default:     int32default.StaticInt32(0),
			},
			"pod_spec_override": schema.StringAttribute{
				Description: "A custom Kubernetes or OpenShift Pod specification in json for ContainerGroups.",
				Optional:    true,
			},
			"policy_instance_percentage": schema.Int32Attribute{
				Description: "Minimum percentage of all instances that will be automatically assigned to this group when new instances come online.",
				Computed:    true,
				Optional:    true,
				Default:     int32default.StaticInt32(0),
			},
			"policy_instance_minimum": schema.Int32Attribute{
				Description: "Minimum number of instances that will be automatically assigned to this group when new instances come online.",
				Computed:    true,
				Optional:    true,
				Default:     int32default.StaticInt32(0),
			},
		},
	}
}

func (r InstanceGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data InstanceGroupModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.IsContainerGroup.ValueBool() {
		if !data.Credential.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("credential"),
				"Invalid Attribute Configuration",
				"Cannot specify credential for Instance Group",
			)
		}
		if data.PodSpecOverride.ValueString() != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("pod_spec_override"),
				"Invalid Attribute Configuration",
				"Cannot specify pod_spec_override for Instance Group",
			)
		}
	} else {
		if data.PolicyInstancePercentage.ValueInt32() != 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("policy_instance_percentage"),
				"Invalid Attribute Configuration",
				"Cannot specify policy_instance_percentage for Container Group",
			)
		}
		if data.PolicyInstanceMinimum.ValueInt32() != 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("policy_instance_minimum"),
				"Invalid Attribute Configuration",
				"Cannot specify policy_instance_minimum for Container Group",
			)
		}
	}
}

func (r *InstanceGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InstanceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceGroupModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData InstanceGroupAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.IsContainerGroup = data.IsContainerGroup.ValueBool()
	bodyData.MaxConcurrentJobs = int(data.MaxConcurrentJobs.ValueInt32())
	bodyData.MaxForks = int(data.MaxForks.ValueInt32())
	bodyData.PolicyInstanceMinimum = int(data.PolicyInstanceMinimum.ValueInt32())
	bodyData.PolicyInstancePercentage = int(data.PolicyInstancePercentage.ValueInt32())

	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}

	if data.PodSpecOverride.ValueString() != "" {
		bodyData.PodSpecOverride = data.PodSpecOverride.ValueString()
	}

	url := "instance_groups/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "")
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

func (r *InstanceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceGroupModel

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

	url := fmt.Sprintf("instance_groups/%d/", id)
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

	var responseData InstanceGroupAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is_container_group"), responseData.IsContainerGroup)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("max_concurrent_jobs"), responseData.MaxConcurrentJobs)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("max_forks"), responseData.MaxForks)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_instance_percentage"), responseData.PolicyInstancePercentage)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_instance_minimum"), responseData.PolicyInstanceMinimum)...)

	if !data.Credential.IsNull() || responseData.Credential != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("credential"), responseData.Credential)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if responseData.PodSpecOverride != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pod_spec_override"), responseData.PodSpecOverride)...)
	}
}

func (r *InstanceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InstanceGroupModel

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

	var bodyData InstanceGroupAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.IsContainerGroup = data.IsContainerGroup.ValueBool()
	bodyData.MaxConcurrentJobs = int(data.MaxConcurrentJobs.ValueInt32())
	bodyData.MaxForks = int(data.MaxForks.ValueInt32())
	bodyData.PolicyInstanceMinimum = int(data.PolicyInstanceMinimum.ValueInt32())
	bodyData.PolicyInstancePercentage = int(data.PolicyInstancePercentage.ValueInt32())

	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}

	if data.PodSpecOverride.ValueString() != "" {
		bodyData.PodSpecOverride = data.PodSpecOverride.ValueString()
	}

	url := fmt.Sprintf("instance_groups/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceGroupModel

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

	url := fmt.Sprintf("instance_groups/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *InstanceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
