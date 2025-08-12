package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccNotificationTemplateDataSource(t *testing.T) {
	objectName := acctest.RandString(5)
	objectName2 := acctest.RandString(5)
	objectName3 := acctest.RandString(5)

	IdComparer := &compareTwoValuesAsStrings{}

	basicWebhookConifg := WebhookConfiguration{
		Url: "https://webhooktarget.com",
		Headers: map[string]any{
			"httpheader1": "testone",
			"httpheader2": 2,
		},
		Username:               "user-abc",
		HttpMethod:             "POST",
		DisableSslVerification: true,
	}

	temp, err := json.Marshal(basicWebhookConifg)
	if err != nil {
		t.Error("Unable to marshal basic webhook config")
		t.Fail()
	}

	webhookConfigString := string(temp)

	webhookConfigComparer := compareTwoWebhookConfigs{
		InitialValue: webhookConfigString,
	}

	webhookConfigComparer1 := statecheck.CompareValue(&webhookConfigComparer)

	basicSlackConfig := SlackConfiguration{
		Channels:  []string{"#channel1", "#channel2"},
		HexColors: "#000003",
	}

	temp, err = json.Marshal(basicSlackConfig)
	if err != nil {
		t.Error("Unable to marshal basic slack config")
		t.Fail()
	}

	slackConfigString := string(temp)

	slackConfigComparer := compareTwoSlackConfigs{
		InitialValue: slackConfigString,
	}

	slackConfigComparer1 := statecheck.CompareValue(&slackConfigComparer)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// test basic webhook case
			{
				Config: testAccNotifTmplWebhookDataSource1Config(objectName, webhookConfigString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_notification_template.example-webhook-type", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(objectName),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("data.%s_notification_template.example-webhook-type", configprefix.Prefix),
						tfjsonpath.New("organization"),
						fmt.Sprintf("%s_organization.example", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdComparer,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_notification_template.example-webhook-type", configprefix.Prefix),
						tfjsonpath.New("notification_type"),
						knownvalue.StringExact("webhook"),
					),
					webhookConfigComparer1.AddStateValue(
						fmt.Sprintf("data.%s_notification_template.example-webhook-type", configprefix.Prefix),
						tfjsonpath.New("notification_configuration"),
					),
				},
			},
			// test basic slack case
			{
				Config: testAccNotifTmplSlackDataSource2Config(objectName2, slackConfigString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_notification_template.example-slack-type", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(objectName2),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("data.%s_notification_template.example-slack-type", configprefix.Prefix),
						tfjsonpath.New("organization"),
						fmt.Sprintf("%s_organization.example2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdComparer,
					),
					slackConfigComparer1.AddStateValue(
						fmt.Sprintf("data.%s_notification_template.example-slack-type", configprefix.Prefix),
						tfjsonpath.New("notification_configuration"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_notification_template.example-slack-type", configprefix.Prefix),
						tfjsonpath.New("notification_type"),
						knownvalue.StringExact("slack"),
					),
				},
			},
			// a simple message field test case
			{
				Config: testAccNotifTmplSlackWithMessagesDataSourceConfig(objectName3, slackConfigString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_notification_template.example-slack-and-message", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(objectName3),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("data.%s_notification_template.example-slack-and-message", configprefix.Prefix),
						tfjsonpath.New("organization"),
						fmt.Sprintf("%s_organization.example", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdComparer,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_notification_template.example-slack-and-message", configprefix.Prefix),
						tfjsonpath.New("messages"),
						knownvalue.StringExact(`{"error":{"body":"","message":""},"started":{"body":"","message":"{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"},"success":{"body":"","message":""},"workflow_approval":{"approved":{"body":"","message":""},"denied":{"body":"","message":""},"running":{"body":"","message":""},"timed_out":{"body":"","message":""}}}`),
					),
				},
			},
		},
	})
}

func testAccNotifTmplWebhookDataSource1Config(objectName, notifConfig string) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "example" { 
	name = "%s" 
	description = "testing example" 
}

resource "awx_notification_template" "example-webhook-type" {
	name              = "%s"
	notification_type = "webhook"
	organization      = awx_organization.example.id
	notification_configuration = jsonencode(%s)
}
	
data "awx_notification_template" "example-webhook-type" {
	id = awx_notification_template.example-webhook-type.id
}`, acctest.RandStringFromCharSet(5, acctest.CharSetAlpha), objectName, notifConfig))
}

func testAccNotifTmplSlackDataSource2Config(objectName, notifConfig string) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "example2" { 
	name = "%s" 
	description = "testing example2" 
}

resource "awx_notification_template" "example-slack-type" {
	name              = "%s"
	notification_type = "slack"
	organization      = awx_organization.example2.id
  notification_configuration = jsonencode(%s)
}
data "awx_notification_template" "example-slack-type" {
	name = awx_notification_template.example-slack-type.name
}`, acctest.RandStringFromCharSet(5, acctest.CharSetAlpha), objectName, notifConfig))
}

func testAccNotifTmplSlackWithMessagesDataSourceConfig(objectName, notifConfig string) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "example" { 
	name = "%s" 
	description = "testing example3" 
}

resource "awx_notification_template" "example-slack-and-message" {
	name              = "%s"
	notification_type = "slack"
	organization      = awx_organization.example.id
  notification_configuration = jsonencode(%s)
  messages = jsonencode({
	  error = {
		body    = ""
		message = ""
	  }
	  started = {
		body    = ""
		message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
	  }
	  success = {
		body    = ""
		message = ""
	  }
	  workflow_approval = {
		approved = {
		  body    = ""
		  message = ""
		}
		denied = {
		  body    = ""
		  message = ""
		}
		running = {
		  body    = ""
		  message = ""
		}
		timed_out = {
		  body    = ""
		  message = ""
		}
	  }
	})
}
data "awx_notification_template" "example-slack-and-message" {
	id = awx_notification_template.example-slack-and-message.id
}`, acctest.RandStringFromCharSet(5, acctest.CharSetAlpha), objectName, notifConfig))
}
