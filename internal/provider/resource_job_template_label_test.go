package provider

import (
	"fmt"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccJobTemplateLabel_basic(t *testing.T) {
	testingJobTemplateName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)

	stringListComparer := &compareStringInList{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateLabel1Config(testingJobTemplateName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_label.test_label_1", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_label.test", configprefix.Prefix),
						tfjsonpath.New("label_ids"),
						stringListComparer,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_label.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						compare.ValuesSame(),
					),
				},
			},
			{
				ResourceName:                         "awx_job_template_label.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateJobTemplateID("awx_job_template_label.test"),
				ImportStateVerifyIdentifierAttribute: ("job_template_id"),
			},
		},
	})
}

func testAccJobTemplateLabel1Config(jobTemplateName string) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test" {
  name        = "%s"
}

resource "awx_project" "test" {
	name         = "%s"
	organization = awx_organization.test.id
	allow_override = true
	scm_type = "git"
	scm_url = "https://github.com/fakerepo"
}	

resource "awx_job_template" "test" {
	name = "%s"
	playbook = "hello_world.yml"
	ask_inventory_on_launch = true
	project = awx_project.test.id
}

resource "awx_label" "test_label_1" {
	name = "testlabel1"
	organization = awx_organization.test.id
}

resource "awx_job_template_label" "test" {
	job_template_id = awx_job_template.test.id
	label_ids = [awx_label.test_label_1.id]
}

  `, acctest.RandString(5), acctest.RandString(5), jobTemplateName))
}
