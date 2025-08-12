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

var _ datasource.DataSource = &GroupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

type GroupDataSource struct {
	client *providerClient
}

func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get group datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Group ID",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Group name. To lookup by name you must also include `inventory`",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Group description.",
				Computed:    true,
			},
			"inventory": schema.Int32Attribute{
				Description: "Id of inventory that this group belongs to. Used to look up groups by `name`",
				Optional:    true,
			},
			"variables": schema.StringAttribute{
				Description: "Group variables in JSON or YAML format.",
				Computed:    true,
			},
		},
	}
}

func (d GroupDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("name"),
			path.MatchRoot("inventory"),
		),
	}
}

func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*providerClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = configureData
}

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	if !data.Id.IsNull() {
		id, err := strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable convert id from string to int.",
				fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
			return
		}
		url = fmt.Sprintf("groups/%d/", id)
	}
	if !data.Name.IsNull() && !data.Inventory.IsNull() {
		name := urlParser.QueryEscape(data.Name.ValueString())
		url = fmt.Sprintf("groups/?name=%s&inventory=%d", name, data.Inventory.ValueInt32())
	}

	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "")
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

	var responseData GroupAPIModel

	if !data.Id.IsNull() && data.Name.IsNull() && data.Inventory.IsNull() {
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error =  %v.", err.Error()))
			return
		}
	}
	if data.Id.IsNull() && !data.Name.IsNull() && !data.Inventory.IsNull() {
		nameResult := struct {
			Count   int             `json:"count"`
			Results []GroupAPIModel `json:"results"`
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
				"Incorrect number of groups returned by name",
				fmt.Sprintf("Unable to read group as API returned %v groups.", nameResult.Count))
			return
		}
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", responseData.Id))
	data.Name = types.StringValue(responseData.Name)
	data.Inventory = types.Int32Value(int32(responseData.Inventory))

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}
	if responseData.Variables != "" {
		data.Variables = types.StringValue(responseData.Variables)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
