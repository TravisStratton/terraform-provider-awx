package provider

import (
	"fmt"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccJobTemplNotifErrResource(t *testing.T) {
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
				Config: testAccJobTemplNotifErr1ResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.test1", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.test", configprefix.Prefix),
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
				Config: testAccJobTemplNotifErr2ResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.test", configprefix.Prefix),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.test3", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.test", configprefix.Prefix),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccJobTemplNotifErr1ResourceConfig() string {
	return configprefix.ReplaceText(fmt.Sprintf(`
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
resource "awx_job_template_notification_template_error" "test" {
  job_template_id    = awx_job_template.test.id
  notif_template_ids = [awx_notification_template.test1.id]
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)))
}

func testAccJobTemplNotifErr2ResourceConfig() string {
	return configprefix.ReplaceText(fmt.Sprintf(`
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
resource "awx_job_template_notification_template_error" "test" {
  job_template_id    = awx_job_template.test.id
  notif_template_ids = [awx_notification_template.test2.id, awx_notification_template.test3.id]
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)))
}
