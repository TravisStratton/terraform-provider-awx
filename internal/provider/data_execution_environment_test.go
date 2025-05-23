package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccExecutionEnvironmentDataSource(t *testing.T) {
	resource1 := ExecutionEnvironmentAPIModel{
		Name:        "test-ee-" + acctest.RandString(5),
		Description: "test execution environment",
		Image:       "quay.io/ansible/awx-ee:latest",
		Pull:        "always",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing
			{
				Config: testAccExecutionEnvironmentDataSourceConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-id",
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-id",
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource1.Image),
					),
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-id",
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource1.Pull),
					),
				},
			},
			// Read by name testing
			{
				Config: testAccExecutionEnvironmentDataSourceConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-name",
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-name",
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource1.Image),
					),
					statecheck.ExpectKnownValue(
						"data.awx_execution_environment.test-name",
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource1.Pull),
					),
				},
			},
		},
	})
}

func testAccExecutionEnvironmentDataSourceConfig(resource ExecutionEnvironmentAPIModel) string {
	return fmt.Sprintf(`
resource "awx_execution_environment" "test" {
  name        	= "%s"
  description 	= "%s"
  image   		= "%s"
  pull 			= "%s"
}
data "awx_execution_environment" "test-id" {
  id = awx_execution_environment.test.id
}
data "awx_execution_environment" "test-name" {
  name = awx_execution_environment.test.name
}
`, resource.Name, resource.Description, resource.Image, resource.Pull)
}
