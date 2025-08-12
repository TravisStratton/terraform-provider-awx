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

func TestAccInventoryDataSource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	resource1 := InventoryAPIModel{
		Name:         "test-inventory-" + acctest.RandString(5),
		Description:  "test description 1",
		Organization: 1,
		Variables:    "{\"foo\":\"bar\"}",
	}
	resource2 := InventoryAPIModel{
		Name:         "test-inventory-" + acctest.RandString(5),
		Description:  "test description 3",
		Organization: 1,
		Variables:    "{\"foo\":\"baz\"}",
		Kind:         "smart",
		HostFilter:   "name__icontains=localhost",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read normal inventory by ID
			{
				Config: testAccInventoryDataSource1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(resource1.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("host_filter"),
						knownvalue.Null(),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test1", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			// Read smart inventory by ID
			{
				Config: testAccInventoryDataSource2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(resource2.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource2.Kind),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("host_filter"),
						knownvalue.StringExact(resource2.HostFilter),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccInventoryDataSource1Config(resource InventoryAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test1" {
  name        			= "%s"
}
resource "awx_inventory" "test1" {
  name         = "%s"
  description  = "%s"
  organization = awx_organization.test1.id
  variables    = jsonencode(%s)
}
data "awx_inventory" "test1" {
  id = awx_inventory.test1.id
}
`, acctest.RandString(5), resource.Name, resource.Description, resource.Variables))
}

func testAccInventoryDataSource2Config(resource InventoryAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test2" {
  name        			= "%s"
}
resource "awx_inventory" "test2" {
  name         	= "%s"
  description  	= "%s"
  organization  = awx_organization.test2.id
  variables    	= jsonencode(%s)
  kind			= "%s"
  host_filter	= "%s"
}
data "awx_inventory" "test2" {
  id = awx_inventory.test2.id
}
`, acctest.RandString(5), resource.Name, resource.Description, resource.Variables, resource.Kind, resource.HostFilter))
}
