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

var _ datasource.DataSource = &ExecutionEnvironmentDataSource{}

func NewExecutionEnvironmentDataSource() datasource.DataSource {
	return &ExecutionEnvironmentDataSource{}
}

type ExecutionEnvironmentDataSource struct {
	client *AwxClient
}

func (d *ExecutionEnvironmentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_execution_environment"
}

func (d *ExecutionEnvironmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get execution environment datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Execution Environment ID.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Execution Environment name.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Execution Environment description.",
				Computed:    true,
			},
			"image": schema.StringAttribute{
				Description: "The full image location, including the container registry, image name, and version tag.",
				Computed:    true,
			},
			"pull": schema.StringAttribute{
				Description: "`always` always pull container before running, `missing` only pull the image if not pressent before running, `never` never pull container before running.",
				Computed:    true,
			},
			"organization": schema.Int32Attribute{
				Description: "Leave this field blank to make the execution environment globally available.",
				Computed:    true,
			},
			"credential": schema.Int32Attribute{
				Description: "Credential to authenticate with a protected container registry.",
				Computed:    true,
			},
		},
	}
}

func (d ExecutionEnvironmentDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *ExecutionEnvironmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*AwxClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = configureData
}

func (d *ExecutionEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExecutionEnvironmentModel

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
		url = fmt.Sprintf("execution_environments/%d/", id)
	}
	if !data.Name.IsNull() {
		// set url for read by name HTTP request
		name := urlParser.QueryEscape(data.Name.ValueString())
		url = fmt.Sprintf("execution_environments/?name=%s", name)
	}

	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404})
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

	if !data.Id.IsNull() && data.Name.IsNull() {
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error =  %v.", err.Error()))
			return
		}
	}
	// If looking up by name, check that there is only one response and extract it.
	if data.Id.IsNull() && !data.Name.IsNull() {
		nameResult := struct {
			Count   int                            `json:"count"`
			Results []ExecutionEnvironmentAPIModel `json:"results"`
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
				"Incorrect number of execution_environments returned by name",
				fmt.Sprintf("Unable to read execution_environment as API returned %v execution_environments.", nameResult.Count))
			return
		}
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Name = types.StringValue(responseData.Name)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.Image != "" {
		data.Image = types.StringValue(responseData.Image)
	}

	if responseData.Pull != "" {
		data.Pull = types.StringValue(responseData.Pull)
	}

	if responseData.Organization != 0 {
		data.Organization = types.Int32Value(int32(responseData.Organization))
	}

	if responseData.Credential != 0 {
		data.Credential = types.Int32Value(int32(responseData.Credential))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
