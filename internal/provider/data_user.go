package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlParser "net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *providerClient
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get user datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User ID. Either `id` or `username` required for datasource.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "The username of the user. Either `id` or `username` required for datasource.",
			},
			"first_name": schema.StringAttribute{
				Computed:    true,
				Description: "User's first name.",
			},
			"last_name": schema.StringAttribute{
				Computed:    true,
				Description: "User's last name.",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "User's email.",
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Description: "User's password. API will return blank or encrypted.",
			},
			"is_superuser": schema.BoolAttribute{
				Computed:    true,
				Description: "Designates that this user has all permissions without explicitly assigning them.",
			},
			"is_system_auditor": schema.BoolAttribute{
				Computed:    true,
				Description: "User is a system wide auditor.",
			},
		},
	}
}

func (d UserDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("username"),
		),
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*providerClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = configureData
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	if !data.Id.IsNull() {
		// set url for read by id HTTP request
		id, err := strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable convert id from string to int.",
				fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
			return
		}
		url = fmt.Sprintf("users/%d/", id)
	}
	if !data.Username.IsNull() {
		// set url for read by username HTTP request
		name := urlParser.QueryEscape(data.Username.ValueString())
		url = fmt.Sprintf("users/?username=%s", name)
	}

	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "gateway")
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

	if !data.Id.IsNull() && data.Username.IsNull() {
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error =  %v.", err.Error()))
			return
		}
	}
	// If looking up by usernamename, check that there is only one response and extract it.
	if data.Id.IsNull() && !data.Username.IsNull() {
		nameResult := struct {
			Count   int            `json:"count"`
			Results []UserAPIModel `json:"results"`
		}{}
		err = json.Unmarshal(body, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count == 1 {
			responseData = nameResult.Results[0]
		} else {
			resp.Diagnostics.AddError(
				"Incorrect number of users returned by name",
				fmt.Sprintf("Unable to read user as API returned %v users.", nameResult.Count))
			return
		}
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Username = types.StringValue(responseData.Username)
	data.Password = types.StringValue(responseData.Password)
	data.IsSuperuser = types.BoolValue(responseData.IsSuperuser)
	data.IsSystemAuditor = types.BoolValue(responseData.IsSystemAuditor)

	if responseData.FirstName != "" {
		data.FirstName = types.StringValue(responseData.FirstName)
	}

	if responseData.LastName != "" {
		data.LastName = types.StringValue(responseData.LastName)
	}

	if responseData.Email != "" {
		data.Email = types.StringValue(responseData.Email)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
