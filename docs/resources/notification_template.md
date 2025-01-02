---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "awx_notification_template Resource - awx"
subcategory: ""
description: |-
  Manage a notification template. These can be attached, by ID, to job templates, as an example usage.
---

# awx_notification_template (Resource)

Manage a notification template. These can be attached, by ID, to job templates, as an example usage.

## Example Usage

```terraform
resource "awx_notification_template" "example" {
  name              = "example1"
  notification_type = "slack"
  organization      = 1
  notification_configuration = jsonencode({
    channels  = ["#channel1", "#channel1"]
    hex_color = ""
    token     = ""
  })
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `notification_type` (String) Only 'slack' is supported in this provider currently. Choose from: email, grafan, irc, mattermost, pagerduty, rocketchat, slack, twilio, webhook.
- `organization` (Number)

### Optional

- `description` (String)
- `messages` (String) json
- `notification_configuration` (String) json. This value depends on the notification_type chosen. But, the value should be json. E.g. notification_configuration = jsonencode(blah blah blah). The AWX Tower API never returns a value for Token. So, this provider is coded to ignore changes to that field.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
terraform import awx_notification_template.example 100
```