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

func TestAccGroupDataSource(t *testing.T) {
	group := GroupAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example group for datasource test",
		Variables:   "{\"foo\":\"bar\"}",
		Inventory:   0, // will be set in config
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig(group),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(group.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(group.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(group.Variables),
					),
				},
			},
			// Lookup by name and inventory
			{
				Config: testAccGroupDataSourceConfigByName(group),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.by_name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(group.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.by_name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(group.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.by_name", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(group.Variables),
					),
				},
			},
		},
	})
}

func testAccGroupDataSourceConfig(resource GroupAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "example" {
  name        = "test-organization-%s"
  description = "test"
}
resource "awx_inventory" "example" {
  name         = "test-inventory-%s"
  description  = "test"
  organization = awx_organization.example.id
}
resource "awx_group" "test" {
  name        = "%s"
  description = "%s"
  inventory   = awx_inventory.example.id
  variables   = jsonencode(%s)
}
data "awx_group" "test" {
  id = awx_group.test.id
}
`, acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Variables))
}

func testAccGroupDataSourceConfigByName(resource GroupAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "example" {
  name        = "test-organization-%s"
  description = "test"
}
resource "awx_inventory" "example" {
  name         = "test-inventory-%s"
  description  = "test"
  organization = awx_organization.example.id
}
resource "awx_group" "test" {
  name        = "%s"
  description = "%s"
  inventory   = awx_inventory.example.id
  variables   = jsonencode(%s)
}
data "awx_group" "by_name" {
  name      = awx_group.test.name
  inventory = awx_inventory.example.id
}
`, acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Variables))
}
