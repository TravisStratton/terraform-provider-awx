resource "awx_notification_template" "example-slack-type" {
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

resource "awx_notification_template" "example-webhook-type" {
  name              = "example2"
  notification_type = "webhook"
  organization      = 1
  notification_configuration = jsonencode({
    url = "https://webhooktarget.com"
    headers = {
      httpheader1 = "travis"
      httpheader2 = 2
    }
    password                 = "thepassword"
    username                 = "user-abc"
    http_method              = "POST"
    disable_ssl_verification = true
  })
}
