package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &InventorySourceResource{}
var _ resource.ResourceWithImportState = &InventorySourceResource{}

func NewInventorySourceResource() resource.Resource {
	return &InventorySourceResource{}
}

type InventorySourceResource struct {
	client *AwxClient
}

func (r *InventorySourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inventory_source"
}

func (r *InventorySourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an AWX inventory_source.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Inventory Source ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Inventory Source name.",
				Required:    true,
			},
			"inventory": schema.Int32Attribute{
				Description: "Inventory ID for the inventory source to be attached to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Description: "Type of SCM resource. Options: `scm`, `ec2`, `gce`, `azure_rm`, `vmware`, `satellite6`, `openstack`, `rhv`, `controller`, `insights`, `terraform`, `openshift_virtualization`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"scm", "ec2", "gce", "azure_rm", "vmware", "satellite6", "openstack", "rhv", "controller", "insights", "terraform", "openshift_virtualization"}...),
				},
			},
			"description": schema.StringAttribute{
				Description: "InventorySource description.",
				Optional:    true,
			},
			"credential": schema.Int32Attribute{
				Description: "Inventory source credential ID.",
				Optional:    true,
			},
			"execution_environment": schema.Int32Attribute{
				Description: "The ID of the execution environment this inventory source.",
				Optional:    true,
			},
			"source_path": schema.StringAttribute{
				Description: "(Inventory file) - The inventory file to be synced by this source.",
				Optional:    true,
			},
			"enabled_value": schema.StringAttribute{
				Description: "This field is ignored unless an Enabled Variable is set. If the enabled variable matches this value, the host will be enabled on import.",
				Optional:    true,
			},
			"enabled_var": schema.StringAttribute{
				Description: "Retrieve the enabled state from the given dict of host variables. The enabled variable may be specified using dot notation, e.g: 'foo.bar'",
				Optional:    true,
			},
			"host_filter": schema.StringAttribute{
				Description: "Regular expression where only matching host names will be imported. The filter is applied as a post-processing step after any inventory plugin filters are applied.",
				Optional:    true,
			},
			"overwrite": schema.BoolAttribute{
				Description: "If checked, any hosts and groups that were previously present on the external source but are now removed will be removed from the inventory. Hosts and groups that were not managed by the inventory source will be promoted to the next manually created group or if there is no manually created group to promote them into, they will be left in the `all` default group for the inventory. When not checked, local child hosts and groups not found on the external source will remain untouched by the inventory update process.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"overwrite_vars": schema.BoolAttribute{
				Description: "If checked, all variables for child groups and hosts will be removed and replaced by those found on the external source. When not checked, a merge will be performed, combining local variables with those found on the external source.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"update_on_launch": schema.BoolAttribute{
				Description: "Each time a job runs using this inventory, refresh the inventory from the selected source before executing job tasks.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"source_vars": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString("---"),
				Computed:    true,
				Description: "Default value is `\"---\"`",
			},
			"source_project": schema.Int32Attribute{
				Description: "The ID of the source project.",
				Optional:    true,
			},
			"scm_branch": schema.StringAttribute{
				Optional:    true,
				Description: "Branch to use on inventory sync. Project default used if blank. Only allowed if project allow_override field is set to true.",
			},
			"update_cache_timeout": schema.Int32Attribute{
				Description: "Time in seconds to consider an inventory sync to be current. During job runs and callbacks the task system will evaluate the timestamp of the latest sync. If it is older than Cache Timeout, it is not considered current, and a new inventory sync will be performed.",
				Optional:    true,
				Default:     int32default.StaticInt32(30),
				Computed:    true,
			},
			"verbosity": schema.Int32Attribute{
				Description: "Control the level of output Ansible will produce for inventory source update jobs. `0 - Warning`, `1 - Info`, `2 - Debug`",
				Optional:    true,
				Default:     int32default.StaticInt32(1),
				Computed:    true,
				Validators: []validator.Int32{
					int32validator.Between(0, 2),
				},
			},
		},
	}
}

func (r InventorySourceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data InventorySourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Required attribute for scm.
	if data.Source.ValueString() == "scm" {
		if data.SourcePath.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("source_path"),
				"Missing Attribute Configuration",
				"Source Control Type requires source_path to be set",
			)
		}
		if data.SourceProject.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("source_project"),
				"Missing Attribute Configuration",
				"Source Control Type requires source_project to be set",
			)
		}
	}

	// Not allowed for all sources except scm.
	if data.Source.ValueString() != "scm" {
		if data.SourcePath.ValueString() != "" || !data.SourcePath.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("source_path"),
				"Attribute Configuration Error",
				"source_path can only be used for scm sources",
			)
		}
		if !data.SourceProject.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("source_project"),
				"Attribute Configuration Error",
				"source_project can only be used for scm sources",
			)
		}
		if data.ScmBranch.ValueString() != "" || !data.ScmBranch.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("scm_branch"),
				"Attribute Configuration Error",
				"scm_branch can only be used for scm sources",
			)
		}

	}
}

func (r *InventorySourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InventorySourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InventorySourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData InventorySourceAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.Inventory = int(data.Inventory.ValueInt32())
	bodyData.Source = data.Source.ValueString()

	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.ExecutionEnvironment.IsNull()) {
		bodyData.ExecutionEnvironment = int(data.ExecutionEnvironment.ValueInt32())
	}
	if !(data.SourcePath.IsNull()) {
		bodyData.SourcePath = data.SourcePath.ValueString()
	}
	if !(data.EnabledValue.IsNull()) {
		bodyData.EnabledValue = data.EnabledValue.ValueString()
	}
	if !(data.EnabledVar.IsNull()) {
		bodyData.EnabledVar = data.EnabledVar.ValueString()
	}
	if !(data.HostFilter.IsNull()) {
		bodyData.HostFilter = data.HostFilter.ValueString()
	}
	if !(data.OverwriteVars.IsNull()) {
		bodyData.OverwriteVars = data.OverwriteVars.ValueBool()
	}
	if !(data.Overwrite.IsNull()) {
		bodyData.Overwrite = data.Overwrite.ValueBool()
	}
	if !(data.SourceVars.IsNull()) {
		bodyData.SourceVars = data.SourceVars.ValueString()
	}
	if !(data.SourceProject.IsNull()) {
		bodyData.SourceProject = int(data.SourceProject.ValueInt32())
	}
	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !(data.UpdateCacheTimeout.IsNull()) {
		bodyData.UpdateCacheTimeout = int(data.UpdateCacheTimeout.ValueInt32())
	}
	if !(data.UpdateOnLaunch.IsNull()) {
		bodyData.UpdateOnLaunch = data.UpdateOnLaunch.ValueBool()
	}
	if !(data.Verbosity.IsNull()) {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}

	url := "inventory_sources/"
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

func (r *InventorySourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InventorySourceModel

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
	url := fmt.Sprintf("inventory_sources/%d/", id)
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

	var responseData InventorySourceAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inventory"), responseData.Inventory)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source"), responseData.Source)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
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

	if !data.ExecutionEnvironment.IsNull() || responseData.ExecutionEnvironment != nil {
		execution_environment, ok := responseData.ExecutionEnvironment.(float64)
		if !ok {
			resp.Diagnostics.AddError("read of execution_environment failed", fmt.Sprintf("unable to cast execution_environment %v to float64", responseData.ExecutionEnvironment))
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("execution_environment"), int32(execution_environment))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.SourcePath.IsNull() || responseData.SourcePath != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_path"), responseData.SourcePath)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.EnabledValue.IsNull() || responseData.EnabledValue != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled_value"), responseData.EnabledValue)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.EnabledVar.IsNull() || responseData.EnabledVar != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled_var"), responseData.EnabledVar)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.HostFilter.IsNull() || responseData.HostFilter != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host_filter"), responseData.HostFilter)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("overwrite_vars"), responseData.OverwriteVars)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("overwrite"), responseData.Overwrite)...)

	if !data.SourceVars.IsNull() || responseData.SourceVars != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_vars"), responseData.SourceVars)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.SourceProject.IsNull() || responseData.SourceProject != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_project"), responseData.SourceProject)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.ScmBranch.IsNull() || responseData.ScmBranch != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_branch"), responseData.ScmBranch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.UpdateCacheTimeout.IsNull() || responseData.UpdateCacheTimeout != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("update_cache_timeout"), responseData.UpdateCacheTimeout)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("update_on_launch"), responseData.UpdateOnLaunch)...)

	if !data.Verbosity.IsNull() || responseData.Verbosity != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("verbosity"), responseData.Verbosity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *InventorySourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InventorySourceModel

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

	var bodyData InventorySourceAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.Inventory = int(data.Inventory.ValueInt32())
	bodyData.Source = data.Source.ValueString()

	if !(data.Credential.IsNull()) {
		bodyData.Credential = int(data.Credential.ValueInt32())
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.ExecutionEnvironment.IsNull()) {
		bodyData.ExecutionEnvironment = int(data.ExecutionEnvironment.ValueInt32())
	}
	if !(data.SourcePath.IsNull()) {
		bodyData.SourcePath = data.SourcePath.ValueString()
	}
	if !(data.EnabledValue.IsNull()) {
		bodyData.EnabledValue = data.EnabledValue.ValueString()
	}
	if !(data.EnabledVar.IsNull()) {
		bodyData.EnabledVar = data.EnabledVar.ValueString()
	}
	if !(data.HostFilter.IsNull()) {
		bodyData.HostFilter = data.HostFilter.ValueString()
	}
	if !(data.OverwriteVars.IsNull()) {
		bodyData.OverwriteVars = data.OverwriteVars.ValueBool()
	}
	if !(data.Overwrite.IsNull()) {
		bodyData.Overwrite = data.Overwrite.ValueBool()
	}
	if !(data.SourceVars.IsNull()) {
		bodyData.SourceVars = data.SourceVars.ValueString()
	}
	if !(data.SourceProject.IsNull()) {
		bodyData.SourceProject = int(data.SourceProject.ValueInt32())
	}
	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !(data.UpdateCacheTimeout.IsNull()) {
		bodyData.UpdateCacheTimeout = int(data.UpdateCacheTimeout.ValueInt32())
	}
	if !(data.UpdateOnLaunch.IsNull()) {
		bodyData.UpdateOnLaunch = data.UpdateOnLaunch.ValueBool()
	}
	if !(data.Verbosity.IsNull()) {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}

	url := fmt.Sprintf("inventory_sources/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InventorySourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InventorySourceModel

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

	url := fmt.Sprintf("inventory_sources/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *InventorySourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
