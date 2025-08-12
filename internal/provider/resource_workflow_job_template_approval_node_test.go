package provider

import (
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

func TestAccWkflwJobTemplApprovalNodeResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	nodeName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWkflwJobTemplApprvlNodeResource1Config(nodeName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(nodeName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact("a description for testing"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(360),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test_timeoutdefault", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(0),
					),
				},
			},
			{
				ResourceName:      "awx_workflow_job_template_approval_node.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "awx_workflow_job_template_approval_node.test_timeoutdefault",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWkflwJobTemplApprvlNodeResource1Config(nodeName string) string {
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
  name      				= "%s"
  ask_inventory_on_launch 	= true
  project   				= awx_project.test.id
  playbook  				= "test.yml"
}
resource "awx_workflow_job_template" "test" {
  name                     = "%s"
  inventory                = awx_inventory.test.id
  organization             = awx_organization.test.id
}

resource "awx_workflow_job_template_approval_node" "test" {
  workflow_job_template_id 	= awx_workflow_job_template.test.id
  name = "%s"
  description = "a description for testing"
  timeout = 360
}

resource "awx_workflow_job_template_approval_node" "test_timeoutdefault" {
  workflow_job_template_id 	= awx_workflow_job_template.test.id
  name = "%s"
  description = "a description for testing"
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), nodeName, nodeName+"1"))
}
