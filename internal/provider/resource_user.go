// SPECIAL: This file may require repo or controller-specific things.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *providerClient
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an Automation Controller user.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username of the user.",
			},
			"first_name": schema.StringAttribute{
				Optional:    true,
				Description: "User's first name.",
			},
			"last_name": schema.StringAttribute{
				Optional:    true,
				Description: "User's last name.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "User's email.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Description: "User's password. If the password is updated in automation controller, due to the api, terraform will not know that it has been changed.",
			},
			"is_superuser": schema.BoolAttribute{
				Optional:    true,
				Description: "Designates that this user has all permissions without explicitly assigning them. Only one of `is_superuser` or `is_system_auditor` is allowed.",
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"is_system_auditor": schema.BoolAttribute{
				Optional:    true,
				Description: "AWX/AAP2.4 only. User is a system wide auditor. Only one of `is_superuser` or `is_system_auditor` is allowed. AAP 2.5 changes this attribute to `is_platform_auditor`, but the API to set that to true does not work.",
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
		},
	}
}

func (r *UserResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("is_superuser"),
			path.MatchRoot("is_system_auditor"),
		),
	}
}

func (r UserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Disallow is_system_auditor for >AAP2.5
	platform, ok := os.LookupEnv("TOWER_PLATFORM")
	if !ok {
		return
	}

	if platform == "awx" || platform == "aap2.4" || data.IsSystemAuditor == basetypes.NewBoolValue(false) || data.IsSystemAuditor.IsNull() {
		return
	} else {
		resp.Diagnostics.AddAttributeError(
			path.Root("is_system_auditor"),
			"Invalid Attribute Configuration",
			"Attribute is_system_auditor is not supported in this version of the provider.",
		)
		return
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData UserAPIModel

	bodyData.Username = data.Username.ValueString()
	bodyData.Password = data.Password.ValueString()
	bodyData.IsSuperuser = data.IsSuperuser.ValueBool()
	bodyData.IsSystemAuditor = data.IsSystemAuditor.ValueBool()

	if !(data.FirstName.IsNull()) {
		bodyData.FirstName = data.FirstName.ValueString()
	}
	if !(data.LastName.IsNull()) {
		bodyData.LastName = data.LastName.ValueString()
	}
	if !(data.Email.IsNull()) {
		bodyData.Email = data.Email.ValueString()
	}

	url := "users/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserModel

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

	url := fmt.Sprintf("users/%d/", id)
	body, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "gateway")
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

	var responseData UserAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), responseData.Username)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is_superuser"), responseData.IsSuperuser)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is_system_auditor"), responseData.IsSystemAuditor)...)

	if !data.FirstName.IsNull() || responseData.FirstName != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("first_name"), responseData.FirstName)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.LastName.IsNull() || responseData.LastName != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("last_name"), responseData.LastName)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Email.IsNull() || responseData.Email != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), responseData.Email)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Always use current state of password to set resp.State as responseData.Password will not be valid
	var statePassword types.String
	diags := req.State.GetAttribute(ctx, path.Root("password"), &statePassword)
	if diags.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("password"), statePassword)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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

	var bodyData UserAPIModel

	bodyData.Username = data.Username.ValueString()
	bodyData.Password = data.Password.ValueString()
	bodyData.IsSuperuser = data.IsSuperuser.ValueBool()
	bodyData.IsSystemAuditor = data.IsSystemAuditor.ValueBool()

	if !(data.FirstName.IsNull()) {
		bodyData.FirstName = data.FirstName.ValueString()
	}
	if !(data.LastName.IsNull()) {
		bodyData.LastName = data.LastName.ValueString()
	}
	if !(data.Email.IsNull()) {
		bodyData.Email = data.Email.ValueString()
	}

	url := fmt.Sprintf("users/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserModel

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
	url := fmt.Sprintf("users/%d/", id)

	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
