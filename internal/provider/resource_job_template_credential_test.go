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

func TestAccJobTemplateCredentialResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateCredentialResource1Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_credential.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:                         "awx_job_template_credential.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateJobTemplateID("awx_job_template_credential.test"),
				ImportStateVerifyIdentifierAttribute: ("job_template_id"),
			},
			{
				Config: testAccJobTemplateCredentialResource2Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_credential.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccJobTemplateCredentialResource1Config() string {
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
data "awx_credential_type" "test" {
  name = "Machine"
  kind = "ssh"
}
resource "awx_credential" "test" {
  name            = "%s"
  description	  = "%s"
  organization    = awx_organization.test.id
  credential_type = data.awx_credential_type.test.id
  inputs = jsonencode({
    "password" : "%s",
    "username" : "%s"
  })
}
resource "awx_job_template" "test" {
  name        = "%s"
  job_type    = "run"
  inventory   = awx_inventory.test.id
  project     = awx_project.test.id
  playbook    = "%s"
}
resource "awx_job_template_credential" "test" {
  credential_ids  = [ awx_credential.test.id ]
  job_template_id = awx_job_template.test.id
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)))
}

func testAccJobTemplateCredentialResource2Config() string {
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
data "awx_credential_type" "test1" {
  name = "Machine"
  kind = "ssh"
}
resource "awx_credential" "test1" {
  name            = "%s"
  description	  = "%s"
  organization    = awx_organization.test.id
  credential_type = data.awx_credential_type.test1.id
  inputs = jsonencode({
    "password" : "%s",
    "username" : "%s"
  })
}
data "awx_credential_type" "test2" {
  name = "Amazon Web Services"
  kind = "cloud"
}
resource "awx_credential" "test2" {
  name            = "%s"
  description	  = "%s"
  organization    = awx_organization.test.id
  credential_type = data.awx_credential_type.test2.id
  inputs = jsonencode({
    "password" : "%s",
    "username" : "%s"
  })
}
resource "awx_job_template" "test" {
  name        = "%s"
  job_type    = "run"
  inventory   = awx_inventory.test.id
  project     = awx_project.test.id
  playbook    = "%s"
}
resource "awx_job_template_credential" "test" {
  credential_ids  = [ awx_credential.test1.id, awx_credential.test2.id ]
  job_template_id = awx_job_template.test.id
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)))
}
