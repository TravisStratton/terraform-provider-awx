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

var _ datasource.DataSource = &NotificationTemplateDataSource{}

func NewNotificationTemplateDataSource() datasource.DataSource {
	return &NotificationTemplateDataSource{}
}

type NotificationTemplateDataSource struct {
	client *providerClient
}

func (d *NotificationTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_template"
}

func (d *NotificationTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get notification template datasource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Notification tempalate ID.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Notification template name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Defaults to `\"\"`",
			},
			"organization": schema.Int32Attribute{
				Computed:    true,
				Description: "Organization ID for the notification template.",
			},
			"notification_type": schema.StringAttribute{
				Computed:    true,
				Description: "Only `slack` and `webhook` are currently supported in this provider. Choose from: `email`, `grafan`, `irc`, `mattermost`, `pagerduty`, `rocketchat`, `slack`, `twilio`, `webhook`.",
			},
			"notification_configuration": schema.StringAttribute{
				Computed:    true,
				Description: "json. This value depends on the `notification_type` chosen. But, the value should be json. E.g. `notification_configuration = jsonencode(blah blah blah)`. The automation controller API never returns a value for Token. So, this provider is coded to ignore changes to that field.",
			},
			"messages": schema.StringAttribute{
				Computed:    true,
				Description: "json",
			},
		},
	}
}

func (d NotificationTemplateDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *NotificationTemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationTemplateModel

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
		url = fmt.Sprintf("notification_templates/%d/", id)
	}
	if !data.Name.IsNull() {
		// set url for read by name HTTP request
		name := urlParser.QueryEscape(data.Name.ValueString())
		url = fmt.Sprintf("notification_templates/?name=%s", name)
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

	var responseData NotificationTemplateAPIModel

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
			Results []NotificationTemplateAPIModel `json:"results"`
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
				"Incorrect number of notification_templates returned by name",
				fmt.Sprintf("Unable to read execution_environment as API returned %v notification_templates.", nameResult.Count))
			return
		}
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

	if responseData.NotificationType != "" {
		data.NotificationType = types.StringValue(responseData.NotificationType)
	}

	if responseData.Organization != 0 {
		data.Organization = types.Int32Value(int32(responseData.Organization))
	}

	var responseSlackConifg SlackConfiguration

	if responseData.NotificationType == "slack" {

		jsonData, err := json.Marshal(responseData.NotificationConfiguration)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
				"Unable to marshal data into json."+err.Error(),
			)
			return
		}

		err = json.Unmarshal(jsonData, &responseSlackConifg)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
				"Unable to unmarshal state data notification configuration into a go type for interogation."+err.Error(),
			)
			return
		}

	}

	var responseWebhookConfig WebhookConfiguration

	if responseData.NotificationType == "webhook" {
		jsonData, err := json.Marshal(responseData.NotificationConfiguration)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
				"Unable to marshal response notification configuration into json for interogation."+err.Error(),
			)
			return
		}
		err = json.Unmarshal(jsonData, &responseWebhookConfig)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
				"Unable to unmarshal response notification configuration into a go type for interogation."+err.Error(),
			)
			return
		}
	}

	if responseData.NotificationType == "webhook" {

		jsonData, err := json.Marshal(responseWebhookConfig)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
				"Unable to marshal data into json."+err.Error(),
			)
			return
		}
		if string(jsonData) != "" {
			data.NotificationConfiguration = types.StringValue(string(jsonData))
		}
	}

	if responseData.NotificationType == "slack" {

		jsonData, err := json.Marshal(responseSlackConifg)
		if err != nil {
			resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
				"Unable to marshal data into json."+err.Error(),
			)
			return
		}
		if string(jsonData) != "" {
			data.NotificationConfiguration = types.StringValue(string(jsonData))
		}
	}

	messages := new(Messages)

	if responseData.Messages != nil {

		foundOneMessageVal := false
		resp_msgs, ok := responseData.Messages.(map[string]any)
		if !ok {
			resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
				"Unexpected error in datasource_notification_templates. resp_msgs not the right type",
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
							resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
								"Unexpected error in datasource_notification_templates with. error msg_message is not the right type",
							)
						}
					} else {
						resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
							"Unexpected error in datasource_notification_templates with. error msg_body is not the right type",
						)
					}
				} else {
					resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
						"Unexpected error in datasource_notification_templates with. error msg is not the right type",
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
							resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
								"Unexpected error in datasource_notification_templates with. started msg_message is not the right type",
							)
						}
					} else {
						resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
							"Unexpected error in datasource_notification_templates with. started msg_body is not the right type",
						)
					}
				} else {
					resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
						"Unexpected error in datasource_notification_templates with. started msg is not the right type",
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
							resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
								"Unexpected error in datasource_notification_templates with. started msg_message is not the right type",
							)
						}
					} else {
						resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
							"Unexpected error in datasource_notification_templates with. started msg_body is not the right type",
						)
					}
				} else {
					resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
						"Unexpected error in datasource_notification_templates with. started msg is not the right type",
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
									resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
										"Unexpected error in esource_notification_templates with. val msg_message is not the right type",
									)
								}
							} else {
								resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
									"Unexpected error in esource_notification_templates with. val msg_body is not the right type",
								)
							}
						} else {
							resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
								"Unexpected error in esource_notification_templates with. val msg is not the right type",
							)
							return
						}

					}
					messages.WorkflowApproval = wka
				} else {
					resp.Diagnostics.AddError("Unexpected error in datasource_notification_templates",
						"Unexpected error in datasource_notification_templates. wkaParent is not the right type",
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
			if string(msgJson) != "" {
				data.Messages = types.StringValue(string(msgJson))
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
