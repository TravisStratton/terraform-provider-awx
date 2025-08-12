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

func TestAccInventorySourceResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	inventory_source1 := InventorySourceAPIModel{
		Name:                 "test-inventory-source-" + acctest.RandString(5),
		Description:          "Example description 1",
		Source:               "scm",
		SourcePath:           "test",
		ExecutionEnvironment: 1,
		Overwrite:            true,
		OverwriteVars:        true,
		UpdateOnLaunch:       true,
	}

	inventory_source2 := InventorySourceAPIModel{
		Name:        "test-inventory-source-" + acctest.RandString(5),
		Description: "Example description 1",
		Source:      "scm",
		SourcePath:  "test",
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventorySourceResource1Config(inventory_source1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(inventory_source1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(inventory_source1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("source"),
						knownvalue.StringExact(inventory_source1.Source),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("source_path"),
						knownvalue.StringExact(inventory_source1.SourcePath),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("overwrite"),
						knownvalue.Bool(inventory_source1.Overwrite),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("overwrite_vars"),
						knownvalue.Bool(inventory_source1.OverwriteVars),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("update_on_launch"),
						knownvalue.Bool(inventory_source1.UpdateOnLaunch),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("source_project"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      "awx_inventory_source.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInventorySourceResource2Config(inventory_source2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(inventory_source2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(inventory_source2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("source"),
						knownvalue.StringExact(inventory_source2.Source),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("source_path"),
						knownvalue.StringExact(inventory_source2.SourcePath),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("overwrite"),
						knownvalue.Bool(inventory_source2.Overwrite),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("overwrite_vars"),
						knownvalue.Bool(inventory_source2.OverwriteVars),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("update_on_launch"),
						knownvalue.Bool(inventory_source2.UpdateOnLaunch),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.test", configprefix.Prefix),
						tfjsonpath.New("source_project"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccInventorySourceResource1Config(resource InventorySourceAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test" {
  name        = "%s"
}

resource "awx_project" "test" {
  name         = "%s"
  organization = awx_organization.test.id
  scm_type     = "git"
  scm_url      = "git@github.com:user/repo.git"
}

resource "awx_inventory" "test" {
  name         = "%s"
  organization = awx_organization.test.id
}

resource "awx_inventory_source" "test" {
  name             		= "%s"
  description	   		= "%s"
  inventory        		= awx_inventory.test.id
  source           		= "%s"
  source_project   		= awx_project.test.id
  source_path      		= "%s"
  overwrite        		= %v
  overwrite_vars   		= %v
  update_on_launch 		= %v
  execution_environment = %v
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Source, resource.SourcePath, resource.Overwrite, resource.OverwriteVars, resource.UpdateOnLaunch, resource.ExecutionEnvironment))
}

func testAccInventorySourceResource2Config(resource InventorySourceAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test" {
  name        = "%s"
}

resource "awx_project" "test" {
  name         = "%s"
  organization = awx_organization.test.id
  scm_type     = "git"
  scm_url      = "git@github.com:user/repo.git"
}

resource "awx_inventory" "test" {
  name         = "%s"
  organization = awx_organization.test.id
}

resource "awx_inventory_source" "test" {
  name             		= "%s"
  description	   		= "%s"
  inventory        		= awx_inventory.test.id
  source           		= "%s"
  source_project   		= awx_project.test.id
  source_path      		= "%s"
  overwrite        		= %v
  overwrite_vars   		= %v
  update_on_launch 		= %v
}
  `, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Source, resource.SourcePath, resource.Overwrite, resource.OverwriteVars, resource.UpdateOnLaunch))
}
