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

func TestAccWkflwJobTemplJobNodeFailureResource(t *testing.T) {
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
				Config: testAccWkflwJobTemplJobNodeFailureResource1Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test1", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
				},
			},
			{
				ResourceName:      "awx_workflow_job_template_node_failure.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWkflwJobTemplJobNodeFailureResource2Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test3", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test4", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test5", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccWkflwJobTemplJobNodeFailureResource1Config() string {
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
resource "awx_workflow_job_template_job_node" "test1" {
  unified_job_template     	= awx_job_template.test.id
  workflow_job_template_id 	= awx_workflow_job_template.test.id
  inventory 				= awx_inventory.test.id
}
resource "awx_workflow_job_template_job_node" "test2" {
  unified_job_template     	= awx_job_template.test.id
  workflow_job_template_id 	= awx_workflow_job_template.test.id
  inventory 				= awx_inventory.test.id
}
resource "awx_workflow_job_template_node_failure" "test" {
  id              = awx_workflow_job_template_job_node.test1.id
  failure_ids = [awx_workflow_job_template_job_node.test2.id]
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)))
}

func testAccWkflwJobTemplJobNodeFailureResource2Config() string {
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
  name      = "%s"
  inventory = awx_inventory.test.id
  project   = awx_project.test.id
  playbook  = "test.yml"
}
resource "awx_workflow_job_template" "test" {
  name                     = "%s"
  inventory                = awx_inventory.test.id
  organization             = awx_organization.test.id
}
resource "awx_workflow_job_template_job_node" "test3" {
  unified_job_template     	= awx_job_template.test.id
  workflow_job_template_id 	= awx_workflow_job_template.test.id
  inventory 				= awx_inventory.test.id
}
resource "awx_workflow_job_template_job_node" "test4" {
  unified_job_template     	= awx_job_template.test.id
  workflow_job_template_id 	= awx_workflow_job_template.test.id
  inventory 				= awx_inventory.test.id
}
resource "awx_workflow_job_template_job_node" "test5" {
  unified_job_template     	= awx_job_template.test.id
  workflow_job_template_id 	= awx_workflow_job_template.test.id
  inventory 				= awx_inventory.test.id
}
resource "awx_workflow_job_template_node_failure" "test" {
  id              = awx_workflow_job_template_job_node.test3.id
  failure_ids = [awx_workflow_job_template_job_node.test4.id, awx_workflow_job_template_job_node.test5.id]
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)))
}
