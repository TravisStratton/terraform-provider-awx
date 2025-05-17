package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NotificationTemplatesResource{}
var _ resource.ResourceWithImportState = &NotificationTemplatesResource{}

func NewNotificationTemplatesResource() resource.Resource {
	return &NotificationTemplatesResource{}
}

type NotificationTemplatesResource struct {
	client *AwxClient
}

type NotificationTemplatesResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Organization              types.Int32  `tfsdk:"organization"`
	NotificationType          types.String `tfsdk:"notification_type"`
	NotificationConfiguration types.String `tfsdk:"notification_configuration"`
	Messages                  types.String `tfsdk:"messages"`
}

type NotificationTemplateAPI struct {
	Id                        int    `json:"id"`
	Name                      string `json:"name"`
	Description               string `json:"description,omitempty"`
	Organization              int    `json:"organization"`
	NotificationType          string `json:"notification_type"`
	NotificationConfiguration any    `json:"notification_configuration,omitempty"`
	Messages                  any    `json:"messages,omitempty"`
}

type SlackConfiguration struct {
	Channels  []string `json:"channels"`
	HexColors string   `json:"hex_color"`
	Token     string   `json:"token"`
}

type WebhookConfiguration struct {
	Url                    string         `json:"url"`
	Headers                map[string]any `json:"headers"`
	Password               string         `json:"password"`
	Username               string         `json:"username"`
	HttpMethod             string         `json:"http_method"`
	DisableSslVerification bool           `json:"disable_ssl_verification"`
}

type MessageValue struct {
	Body    string `json:"body"`
	Message string `json:"message"`
}

type Messages struct {
	Error            MessageValue            `json:"error"`
	Started          MessageValue            `json:"started"`
	Success          MessageValue            `json:"success"`
	WorkflowApproval map[string]MessageValue `json:"workflow_approval"`
}

func (r *NotificationTemplatesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_template"
}

func (r *NotificationTemplatesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage a notification template. These can be attached, by ID, to job templates, as an example usage.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Notification template name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Computed:    true,
				Description: "Defaults to `\"\"`",
			},
			"organization": schema.Int32Attribute{
				Required:    true,
				Description: "Organization ID for the notification template.",
			},
			"notification_type": schema.StringAttribute{
				Required:    true,
				Description: "Only `slack` and `webhook` are currently supported in this provider. Choose from: `email`, `grafan`, `irc`, `mattermost`, `pagerduty`, `rocketchat`, `slack`, `twilio`, `webhook`.",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"slack", "webhook"}...),
				},
			},
			"notification_configuration": schema.StringAttribute{
				Optional:    true,
				Description: "json. This value depends on the `notification_type` chosen. But, the value should be json. E.g. `notification_configuration = jsonencode(blah blah blah)`. The AWX Tower API never returns a value for Token. So, this provider is coded to ignore changes to that field.",
			},
			"messages": schema.StringAttribute{
				Optional:    true,
				Description: "json",
			},
		},
	}
}

func (r *NotificationTemplatesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationTemplatesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationTemplatesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData NotificationTemplateAPI
	bodyData.Name = data.Name.ValueString()
	bodyData.Description = data.Description.ValueString()
	bodyData.Organization = int(data.Organization.ValueInt32())
	bodyData.NotificationType = data.NotificationType.ValueString()

	if !data.NotificationConfiguration.IsNull() {
		fieldToBytes := []byte(data.NotificationConfiguration.ValueString())

		var notifConfig any

		if data.NotificationType.ValueString() == "slack" {
			notifConfig = new(SlackConfiguration)
		} else {
			notifConfig = new(WebhookConfiguration)
		}

		err := json.Unmarshal(fieldToBytes, &notifConfig)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to move Notification Config into json object",
				fmt.Sprintf("Error = %s ", err.Error()))
			return
		}

		bodyData.NotificationConfiguration = notifConfig
	}
	if !data.Messages.IsNull() {
		fieldToBytes := []byte(data.Messages.ValueString())

		messageData := new(Messages)

		err := json.Unmarshal(fieldToBytes, &messageData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to move Messages into json object",
				fmt.Sprintf("Error = %s ", err.Error()))
			return
		}

		bodyData.Messages = messageData
	}

	url := "notification_templates/"
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

func (r *NotificationTemplatesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationTemplatesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("notification_templates/%d/", id)
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
	var responseData NotificationTemplateAPI

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable unmarshal response body into object",
			fmt.Sprintf("Error =  %v. ", err.Error()))
		return
	}

	if !data.Name.IsNull() || responseData.Name != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Organization.IsNull() || responseData.Organization != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), responseData.Organization)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.NotificationType.IsNull() || responseData.NotificationType != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("notification_type"), responseData.NotificationType)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var responseSlackConifg SlackConfiguration

	if responseData.NotificationType == "slack" {

		jsonData, err := json.Marshal(responseData.NotificationConfiguration)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
				"Unable to marshal data into json."+err.Error(),
			)
			return
		}

		err = json.Unmarshal(jsonData, &responseSlackConifg)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
				"Unable to unmarshal state data notification configuration into a go type for interogation."+err.Error(),
			)
			return
		}

	}

	var responseWebhookConfig WebhookConfiguration

	if responseData.NotificationType == "webhook" {
		jsonData, err := json.Marshal(responseData.NotificationConfiguration)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
				"Unable to marshal response notification configuration into json for interogation."+err.Error(),
			)
			return
		}
		err = json.Unmarshal(jsonData, &responseWebhookConfig)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
				"Unable to unmarshal response notification configuration into a go type for interogation."+err.Error(),
			)
			return
		}
	}

	var stateSlackConfig SlackConfiguration

	if data.NotificationType.ValueString() == "slack" {

		err := json.Unmarshal([]byte(data.NotificationConfiguration.ValueString()), &stateSlackConfig)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
				"Unable to unmarshal state data notification configuration into a go type for interogation."+err.Error(),
			)
			return
		}
	}

	var stateWebhookConfig WebhookConfiguration

	if data.NotificationType.ValueString() == "webhook" {

		err = json.Unmarshal([]byte(data.NotificationConfiguration.ValueString()), &stateWebhookConfig)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
				"Unable to unmarshal state data notification configuration into a go type for interogation."+err.Error(),
			)
			return
		}
	}

	if responseData.NotificationType == "webhook" {

		// because the API always sends back $encrypted$ for secrets with an HTTP GET, use state value for compare instead
		responseWebhookConfig.Password = stateWebhookConfig.Password

		if !reflect.DeepEqual(stateWebhookConfig, responseWebhookConfig) {
			jsonData, err := json.Marshal(responseWebhookConfig)
			if err != nil {
				resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
					"Unable to marshal data into json."+err.Error(),
				)
				return
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("notification_configuration"), string(jsonData))...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	if responseData.NotificationType == "slack" {

		// because the API always sends back $encrypted$ for secrets with an HTTP GET, use state value for compare instead
		responseSlackConifg.Token = stateSlackConfig.Token

		if !reflect.DeepEqual(stateSlackConfig, responseSlackConifg) {
			jsonData, err := json.Marshal(responseSlackConifg)
			if err != nil {
				resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
					"Unable to marshal data into json."+err.Error(),
				)
				return
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("notification_configuration"), string(jsonData))...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	messages := new(Messages)

	if responseData.Messages != nil {

		foundOneMessageVal := false
		resp_msgs, ok := responseData.Messages.(map[string]any)
		if !ok {
			resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
				"Unexpected error in resource_notification_templates. resp_msgs not the right type",
			)
		}
		for k, v := range resp_msgs {
			if v == nil {
				continue
			} else {
				foundOneMessageVal = true
			}
			if k == "error" {

				if errorMsg, ok := v.(map[string]any); ok {
					if msg_body, ok := errorMsg["body"].(string); ok {
						if msg_message, ok := errorMsg["message"].(string); ok {
							messages.Error = MessageValue{Body: msg_body, Message: msg_message}
						} else {
							resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
								"Unexpected error in esource_notification_templates with. error msg_message is not the right type",
							)
						}
					} else {
						resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
							"Unexpected error in esource_notification_templates with. error msg_body is not the right type",
						)
					}
				} else {
					resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
						"Unexpected error in esource_notification_templates with. error msg is not the right type",
					)
					return
				}

			}
			if k == "started" {
				if startMsg, ok := v.(map[string]any); ok {
					if msg_body, ok := startMsg["body"].(string); ok {
						if msg_message, ok := startMsg["message"].(string); ok {
							messages.Started = MessageValue{Body: msg_body, Message: msg_message}
						} else {
							resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
								"Unexpected error in esource_notification_templates with. started msg_message is not the right type",
							)
						}
					} else {
						resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
							"Unexpected error in esource_notification_templates with. started msg_body is not the right type",
						)
					}
				} else {
					resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
						"Unexpected error in esource_notification_templates with. started msg is not the right type",
					)
					return
				}
			}
			if k == "success" {
				if successMsg, ok := v.(map[string]any); ok {
					if msg_body, ok := successMsg["body"].(string); ok {
						if msg_message, ok := successMsg["message"].(string); ok {
							messages.Success = MessageValue{Body: msg_body, Message: msg_message}
						} else {
							resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
								"Unexpected error in esource_notification_templates with. started msg_message is not the right type",
							)
						}
					} else {
						resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
							"Unexpected error in esource_notification_templates with. started msg_body is not the right type",
						)
					}
				} else {
					resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
						"Unexpected error in esource_notification_templates with. started msg is not the right type",
					)
					return
				}
			}
			if k == "workflow_approval" {
				if wkaParentValue, ok := v.(map[string]any); ok {
					wkaParent := wkaParentValue
					wka := make(map[string]MessageValue, len(wkaParent))
					for key, val := range wkaParent {

						if valMsg, ok := val.(map[string]any); ok {
							if msg_body, ok := valMsg["body"].(string); ok {
								if msg_message, ok := valMsg["message"].(string); ok {
									wka[key] = MessageValue{Body: msg_body, Message: msg_message}
								} else {
									resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
										"Unexpected error in esource_notification_templates with. val msg_message is not the right type",
									)
								}
							} else {
								resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
									"Unexpected error in esource_notification_templates with. val msg_body is not the right type",
								)
							}
						} else {
							resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
								"Unexpected error in esource_notification_templates with. val msg is not the right type",
							)
							return
						}

					}
					messages.WorkflowApproval = wka
				} else {
					resp.Diagnostics.AddError("Unexpected error in resource_notification_templates",
						"Unexpected error in resource_notification_templates. wkaParent is not the right type",
					)
				}
			}
		}

		msgJson, err := json.Marshal(messages)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to move Messages into json object",
				fmt.Sprintf("Error = %s ", err.Error()))
			return
		}

		if (!data.Messages.IsNull() || responseData.Messages != nil) && foundOneMessageVal {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("messages"), string(msgJson))...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
}

func (r *NotificationTemplatesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NotificationTemplatesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}

	var bodyData NotificationTemplateAPI

	bodyData.Name = data.Name.ValueString()

	if !data.Description.IsNull() && data.Description.ValueString() != "" {
		bodyData.Description = data.Description.ValueString()
	}
	bodyData.Organization = int(data.Organization.ValueInt32())
	bodyData.NotificationType = data.NotificationType.ValueString()

	fieldToBytes := []byte(data.NotificationConfiguration.ValueString())

	var notifConfig any

	if !data.NotificationConfiguration.IsNull() {
		fieldToBytes := []byte(data.NotificationConfiguration.ValueString())

		var notifConfig any

		if data.NotificationConfiguration.ValueString() == "slack" {
			notifConfig = new(SlackConfiguration)
		} else {
			notifConfig = new(WebhookConfiguration)
		}

		err := json.Unmarshal(fieldToBytes, &notifConfig)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to move Notification Config into json object",
				fmt.Sprintf("Error = %s ", err.Error()))
			return
		}

		bodyData.NotificationConfiguration = notifConfig
	}

	err = json.Unmarshal(fieldToBytes, &notifConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to move Notification Config into json object",
			fmt.Sprintf("Error = %s ", err.Error()))
		return
	}

	bodyData.NotificationConfiguration = notifConfig

	if !data.Messages.IsNull() {
		fieldToBytes = []byte(data.Messages.ValueString())

		messageData := new(Messages)

		err = json.Unmarshal(fieldToBytes, &messageData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to move Messages into json object",
				fmt.Sprintf("Error = %s ", err.Error()))
			return
		}

		bodyData.Messages = messageData
	}

	url := fmt.Sprintf("notification_templates/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationTemplatesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NotificationTemplatesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("notification_templates/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *NotificationTemplatesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
