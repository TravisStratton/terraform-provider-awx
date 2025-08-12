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

func TestAccExecutionEnvironmentResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	resource1 := ExecutionEnvironmentAPIModel{
		Name:        "test-ee-" + acctest.RandString(5),
		Description: "test execution environment1",
		Image:       "quay.io/ansible/awx-ee:latest",
		Pull:        "always",
	}

	resource2 := ExecutionEnvironmentAPIModel{
		Name:        "test-ee-" + acctest.RandString(5),
		Description: "test execution environment2",
		Image:       "quay.io/ansible/awx-ee:latest",
		Pull:        "never",
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExecutionEnvironmentResource1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource1.Image),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource1.Pull),
					),
				},
			},
			{
				ResourceName:      "awx_execution_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExecutionEnvironmentResource2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource2.Image),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource2.Pull),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_execution_environment.test", configprefix.Prefix),
						tfjsonpath.New("credential"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccExecutionEnvironmentResource1Config(resource ExecutionEnvironmentAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_execution_environment" "test" {
  name        	= "%s"
  description 	= "%s"
  image   		= "%s"
  pull 			= "%s"
}
  `, resource.Name, resource.Description, resource.Image, resource.Pull))
}

func testAccExecutionEnvironmentResource2Config(resource ExecutionEnvironmentAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
data "awx_credential_type" "test" {
  name = "Container Registry"
  kind = "registry"
}
resource "awx_organization" "test" {
  name        = "%s"
}
resource "awx_credential" "test" {
  name            = "%s"
  organization    = awx_organization.test.id
  credential_type = data.awx_credential_type.test.id
  inputs = jsonencode({
	"host" : "quay.io",
	"username" : "test",
	"password" : "%s",
	"verify_ssl" : true
  })
}
resource "awx_execution_environment" "test" {
  name        	= "%s"
  description 	= "%s"
  image   		= "%s"
  pull 			= "%s"
  credential	= awx_credential.test.id
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Image, resource.Pull))
}
