package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CredentialDataSource{}

func NewCredentialDataSource() datasource.DataSource {
	return &CredentialDataSource{}
}

type CredentialDataSource struct {
	client *providerClient
}

func (d *CredentialDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (d *CredentialDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get credential datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential ID.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Credential name.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Credential description.",
				Computed:    true,
			},
			"kind": schema.StringAttribute{
				Description: "Credential kind.",
				Computed:    true,
			},
			"organization": schema.Int32Attribute{
				Description: "ID of organization which owns this credential. One and only one of `organization`, `team`, or `user` must be set.",
				Computed:    true,
			},
			"team": schema.Int32Attribute{
				Description: "ID of team which owns this credential. One and only one of `organization`, `team`, or `user` must be set.",
				Computed:    true,
			},
			"user": schema.Int32Attribute{
				Description: "ID of user which owns this credential. One and only one of `organization`, `team`, or `user` must be set.",
				Computed:    true,
			},
			"credential_type": schema.Int32Attribute{
				Description: "ID of the credential type.",
				Computed:    true,
			},
			"inputs": schema.StringAttribute{
				Description: "Credential inputs.",
				Computed:    true,
			},
		},
	}
}

func (d *CredentialDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CredentialModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int.",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}

	url = fmt.Sprintf("credentials/%d/", id)
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

	var responseData CredentialAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into object",
			fmt.Sprintf("Error =  %v.", err.Error()))
		return
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Name = types.StringValue(responseData.Name)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.Organization != 0 {
		data.Organization = types.Int32Value(int32(responseData.Organization))
	}

	data.Kind = types.StringValue(responseData.Kind)
	data.CredentialType = types.Int32Value(int32(responseData.CredentialType))

	jsonInputs, err := json.Marshal(responseData.Inputs)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Convert to string and print
	jsonString := string(jsonInputs)

	data.Inputs = types.StringValue(jsonString)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
