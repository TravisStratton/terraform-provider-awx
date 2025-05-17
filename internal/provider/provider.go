package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure Provider satisfies various provider interfaces.
var _ provider.Provider = &awxProvider{}
var _ provider.ProviderWithFunctions = &awxProvider{}

// awxProvider defines the provider implementation.
type awxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// awxProviderModel describes the provider data model.
type awxProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Platform types.String `tfsdk:"platform"`
}

func (p *awxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awx"
	resp.Version = p.version
}

func (p *awxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This is a Terraform Provider for managing resources in AWX/Tower or Ansible Automation Platform (AAP).",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "URL for AWX (i.e. https://tower.example.com)",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "AWX access token (instead of username/password)",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "AWX username (instead of token)",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "AWX password (instead of token)",
				Optional:    true,
			},
			"platform": schema.StringAttribute{
				Description: "Does the endpoint point to an Ansible Automation Platform (AAP) version 2.5, verion 2.4, or AWX/Tower environment? Acceptable values are `awx`, `aap2.4`, or `aap2.5`. A default value of `awx` will be assumed if this field is not set. You can also set this using the TOWER_PLATFORM environment variable.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("aap2.4", "aap2.5", "awx"),
				},
			},
		},
	}
}

func (p *awxProvider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		providervalidator.Conflicting(
			path.MatchRoot("token"),
			path.MatchRoot("username"),
		),
		providervalidator.Conflicting(
			path.MatchRoot("token"),
			path.MatchRoot("password"),
		),
		providervalidator.RequiredTogether(
			path.MatchRoot("username"),
			path.MatchRoot("password"),
		),
	}
}

func (p *awxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var (
		token, endpoint, username, password, auth, platform string
	)

	var data awxProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	} else {
		endpoint = os.Getenv("TOWER_HOST")
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing API Endpoint Configuration",
			"While configuring the provider, the API endpoint hostname was not found in "+
				"the TOWER_HOST environment variable or provider "+
				"configuration block endpoint attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	} else {
		// strip trailing '/' character from endpoint if present
		endpointRunes := []rune(endpoint)
		lastChar := endpointRunes[len(endpointRunes)-1]
		if lastChar == '/' {
			endpoint = string(endpointRunes[:len(endpointRunes)-1])
		}
	}

	envToken, tokenExists := os.LookupEnv("TOWER_OAUTH_TOKEN")
	envUsername, userExists := os.LookupEnv("TOWER_USERNAME")
	envPassword, passwordExists := os.LookupEnv("TOWER_PASSWORD")

	// Get token if password/username not set
	if data.Token.IsNull() && data.Username.IsNull() && data.Password.IsNull() && tokenExists {
		token = envToken
	}

	if data.Token.IsNull() && data.Username.IsNull() && data.Password.IsNull() && !tokenExists && userExists && passwordExists {
		username = envUsername
		password = envPassword
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	if (token != "" && (username != "" || password != "")) || (token == "" && (username == "" || password == "")) {
		resp.Diagnostics.AddError(
			"Provider Configuration Error",
			"Specify a token (TOWER_OAUTH_TOKEN) OR username/password (TOWER_USERNAME/TOWER_PASSWORD).")
		return
	}

	if token != "" {
		auth = "Bearer" + " " + token
	} else {
		authString := username + ":" + password
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))
		auth = "Basic" + " " + encodedAuth
	}

	httpclient := &http.Client{
		Timeout: 30 * time.Second,
	}

	client := new(AwxClient)

	client.client = httpclient
	client.endpoint = endpoint
	client.auth = auth

	if !data.Platform.IsNull() {
		platform = data.Platform.ValueString()
		os.Setenv("TOWER_PLATFORM", platform)
	}

	envPlatform, platformExists := os.LookupEnv("TOWER_PLATFORM")

	if platformExists {
		platform = envPlatform
	}

	if platform == "" {
		platform = "awx"
	}

	client.platform = platform

	if client.platform == "awx" || client.platform == "aap2.4" {
		client.urlPrefix = "/api/v2/"
	} else { // aap2.5
		client.urlPrefix = "/api/controller/v2/"
	}

	url := "me/"

	_, _, err := client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"tower authentication failure",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *awxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExecutionEnvironmentResource,
		NewCredentialResource,
		NewCredentialInputSourcesResource,
		NewCredentialTypeResource,
		NewHostResource,
		NewInstanceGroupResource,
		NewInventoryResource,
		NewInventorySourceResource,
		NewJobTemplateCredentialResource,
		NewJobTemplateInstanceGroupsResource,
		NewJobTemplateLabelsResource,
		NewJobTemplateNotifTemplErrResource,
		NewJobTemplateNotifTemplStartedResource,
		NewJobTemplateNotifTemplSuccessResource,
		NewJobTemplateResource,
		NewJobTemplateSurveyResource,
		NewLabelsResource,
		NewNotificationTemplatesResource,
		NewOrganizationResource,
		NewProjectResource,
		NewScheduleResource,
		NewUserResource,
		NewWorkflowJobTemplatesResource,
		NewWorkflowJobTemplatesJobNodeResource,
		NewWorkflowJobTemplatesNodeLabelResource,
		NewWorkflowJobTemplatesNodeSuccessResource,
		NewWorkflowJobTemplatesNodeFailureResource,
		NewWorkflowJobTemplatesNodeAlwaysResource,
		NewWorkflowJobTemplateApprovalNodeResource,
	}
}

func (p *awxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCredentialDataSource,
		NewCredentialTypeDataSource,
		NewExecutionEnvironmentDataSource,
		NewHostDataSource,
		NewInventoryDataSource,
		NewInventorySourceDataSource,
		NewInstanceGroupDataSource,
		NewJobTemplateDataSource,
		NewOrganizationDataSource,
		NewProjectDataSource,
		NewScheduleDataSource,
		NewUserDataSource,
	}
}

func (p *awxProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		//NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &awxProvider{
			version: version,
		}
	}
}
