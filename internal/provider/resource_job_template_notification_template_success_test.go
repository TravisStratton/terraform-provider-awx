package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccJobTemplNotifSuccessResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	StringListCompare := &compareStringInList{}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplNotifSuccess1ResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						"awx_job_template.test",
						tfjsonpath.New("id"),
						"awx_job_template_notification_template_success.test",
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						"awx_notification_template.test1",
						tfjsonpath.New("id"),
						"awx_job_template_notification_template_success.test",
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
			{
				ResourceName:      "awx_job_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobTemplNotifSuccess2ResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						"awx_job_template.test",
						tfjsonpath.New("id"),
						"awx_job_template_notification_template_success.test",
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						"awx_notification_template.test2",
						tfjsonpath.New("id"),
						"awx_job_template_notification_template_success.test",
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						"awx_notification_template.test3",
						tfjsonpath.New("id"),
						"awx_job_template_notification_template_success.test",
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccJobTemplNotifSuccess1ResourceConfig() string {
	return fmt.Sprintf(`
resource "awx_organization" "test" {
  name        = "%s"
}

resource "awx_inventory" "test" {
  name         = "%s"
  organization = awx_organization.test.id
}

resource "awx_project" "test" {
  name         		= "%s"
  organization 		= awx_organization.test.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "awx_job_template" "test" {
  name        = "%s"
  job_type    = "run"
  inventory   = awx_inventory.test.id
  project     = awx_project.test.id
  playbook    = "test.yml"
}
resource "awx_notification_template" "test1" {
  name              = "%s"
  notification_type = "slack"
  organization      = awx_organization.test.id
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
    success = {
      body    = ""
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    started = {
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
resource "awx_job_template_notification_template_success" "test" {
  job_template_id    = awx_job_template.test.id
  notif_template_ids = [awx_notification_template.test1.id]
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5))
}

func testAccJobTemplNotifSuccess2ResourceConfig() string {
	return fmt.Sprintf(`
resource "awx_organization" "test" {
  name        = "%s"
}

resource "awx_inventory" "test" {
  name         = "%s"
  organization = awx_organization.test.id
}

resource "awx_project" "test" {
  name         		= "%s"
  organization 		= awx_organization.test.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "awx_job_template" "test" {
  name        = "%s"
  job_type    = "run"
  inventory   = awx_inventory.test.id
  project     = awx_project.test.id
  playbook    = "test.yml"
}
resource "awx_notification_template" "test2" {
  name              = "%s"
  notification_type = "slack"
  organization      = awx_organization.test.id
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
    success = {
      body    = ""
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    started = {
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
resource "awx_notification_template" "test3" {
  name              = "%s"
  notification_type = "slack"
  organization      = awx_organization.test.id
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
    success = {
      body    = ""
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    started = {
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
resource "awx_job_template_notification_template_success" "test" {
  job_template_id    = awx_job_template.test.id
  notif_template_ids = [awx_notification_template.test2.id, awx_notification_template.test3.id]
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5))
}
