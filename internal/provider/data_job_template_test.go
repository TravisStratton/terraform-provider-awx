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

func TestAccJobTemplateDataSource(t *testing.T) {
	resource1 := JobTemplateAPIModel{
		Name:        "test-job-template-" + acctest.RandString(5),
		Description: "test description 1",
		JobType:     "run",
		Playbook:    "test.yml",
	}
	resource2 := JobTemplateAPIModel{
		Name:        "test-job-template-" + acctest.RandString(5),
		Description: "test description 2",
		JobType:     "check",
		Playbook:    "test2.yml",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read job_template by ID
			{
				Config: testAccJobTemplateDataSource1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-id", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-id", configprefix.Prefix),
						tfjsonpath.New("job_type"),
						knownvalue.StringExact(resource1.JobType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-id", configprefix.Prefix),
						tfjsonpath.New("playbook"),
						knownvalue.StringExact(resource1.Playbook),
					),
				},
			},
			// Read job_template by name
			{
				Config: testAccJobTemplateDataSource2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-name", configprefix.Prefix),
						tfjsonpath.New("job_type"),
						knownvalue.StringExact(resource2.JobType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_job_template.test-name", configprefix.Prefix),
						tfjsonpath.New("playbook"),
						knownvalue.StringExact(resource2.Playbook),
					),
				},
			},
		},
	})
}

func testAccJobTemplateDataSource1Config(resource JobTemplateAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test-id" {
  name        = "%s"
}

resource "awx_inventory" "test-id" {
  name         = "%s"
  organization = awx_organization.test-id.id
}

resource "awx_project" "test-id" {
  name         		= "%s"
  organization 		= awx_organization.test-id.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "awx_job_template" "test-id" {
  name        = "%s"
  description = "%s"
  job_type    = "%s"
  inventory   = awx_inventory.test-id.id
  project     = awx_project.test-id.id
  playbook    = "%s"
}
data "awx_job_template" "test-id" {
  id = awx_job_template.test-id.id
}
`, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.JobType, resource.Playbook))
}

func testAccJobTemplateDataSource2Config(resource JobTemplateAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test-name" {
  name        = "%s"
}

resource "awx_inventory" "test-name" {
  name         = "%s"
  organization = awx_organization.test-name.id
}

resource "awx_project" "test-name" {
  name         		= "%s"
  organization 		= awx_organization.test-name.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "awx_job_template" "test-name" {
  name        = "%s"
  description = "%s"
  job_type    = "%s"
  inventory   = awx_inventory.test-name.id
  project     = awx_project.test-name.id
  playbook    = "%s"
}
data "awx_job_template" "test-name" {
  id = awx_job_template.test-name.id
}
`, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.JobType, resource.Playbook))
}
