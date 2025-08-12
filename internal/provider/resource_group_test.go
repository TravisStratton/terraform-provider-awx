package provider

import (
	"fmt"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccGroupResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}

	Group1 := GroupAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example 1",
		Variables:   "{\"foo\":\"bar\"}",
	}

	Group2 := GroupAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example 2",
		Variables:   "{\"baz\":\"qux\"}",
	}

	ReplacementInventory := InventoryAPIModel{
		Name:        "test-inventory-" + acctest.RandString(5),
		Description: "Example Replacement",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupResourceConfig(Group1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(Group1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(Group1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(Group1.Variables),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      "awx_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupResource2Config(Group2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(Group2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(Group2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(Group2.Variables),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
				},
			},
			// change an existing inventory on an existing group and verify plan marks for re-create
			{
				Config: testAccSecondInvResourceConfig(ReplacementInventory) + testAccInvPlanRecreateConfig(Group2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("awx_group.test2", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func testAccSecondInvResourceConfig(resource InventoryAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
	resource "awx_organization" "new-inv-test-org" {
		name = "test-organizatio-%s"
	}
	resource "awx_inventory" "new-inventory" {
		name = "%s"
		organization = awx_organization.new-inv-test-org.id
	}
		`, acctest.RandString(5), resource.Name))
}

// touches same group created by testAccGroupResource2Config(), but just changes the inventory value,
//
//	doing this to verify that the replace functionality works when updating this.
func testAccInvPlanRecreateConfig(resource GroupAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_group" "test2" {
  name        = "%s"
  description = "%s"
  inventory   = awx_inventory.new-inventory.id
  variables   = jsonencode(%s)
}
  `, resource.Name, resource.Description, resource.Variables))
}

func testAccGroupResourceConfig(resource GroupAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test" {
  name        = "test-organization-%s"
  description = "test"
}
resource "awx_inventory" "test" {
  name         = "test-inventory-%s"
  description  = "test"
  organization = awx_organization.test.id
}
resource "awx_group" "test" {
  name        = "%s"
  description = "%s"
  inventory   = awx_inventory.test.id
  variables   = jsonencode(%s)
}
  `, acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Variables))
}

func testAccGroupResource2Config(resource GroupAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test2" {
  name        = "test-organization-%s"
  description = "test"
}
resource "awx_inventory" "test2" {
  name         = "test-inventory-%s"
  description  = "test"
  organization = awx_organization.test2.id
}
resource "awx_group" "test2" {
  name        = "%s"
  description = "%s"
  inventory   = awx_inventory.test2.id
  variables   = jsonencode(%s)
}
  `, acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Variables))
}
