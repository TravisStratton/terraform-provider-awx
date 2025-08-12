package provider

import (
	"fmt"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccGroupHostResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGrpHstOrgInv() + testAccGrpHst1stPass() + testAccGrpHst1stPassGrp2(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link", configprefix.Prefix),
						tfjsonpath.New("group_id"),
						fmt.Sprintf("%s_group.group-example", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link", configprefix.Prefix),
						tfjsonpath.New("host_id"),
						fmt.Sprintf("%s_host.host-1", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link-2", configprefix.Prefix),
						tfjsonpath.New("group_id"),
						fmt.Sprintf("%s_group.group-example", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link-2", configprefix.Prefix),
						tfjsonpath.New("host_id"),
						fmt.Sprintf("%s_host.host-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp2-host-link", configprefix.Prefix),
						tfjsonpath.New("group_id"),
						fmt.Sprintf("%s_group.group-example-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp2-host-link", configprefix.Prefix),
						tfjsonpath.New("host_id"),
						fmt.Sprintf("%s_host.host-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccGrpHstOrgInv() + testAccGrpHst1stPass() + testAccGrpHst2ndPassGrp2(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("awx_group_host.grp2-host-link", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func testAccGrpHstOrgInv() string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "example" {
  name        = "%s"
  description = "example"
}

resource "awx_inventory" "example" {
  name         = "%s"
  description  = "example"
  organization = awx_organization.example.id
}	
	`, acctest.RandString(5), acctest.RandString(5)))
}

func testAccGrpHst1stPass() string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_group" "group-example" {
  name        = "%s"
  description = "Example with jsonencoded variables."
  inventory   = awx_inventory.example.id
  variables = jsonencode(
    {
      foo = "bar"
      baz = "qux"
    }
  )
}

resource "awx_host" "host-1" {
  name = "%s"
  inventory = awx_inventory.example.id
}

resource "awx_host" "host-2" {
  name = "%s"
  inventory = awx_inventory.example.id
}

resource "awx_group_host" "grp-host-link" {
  group_id = awx_group.group-example.id
  host_id = awx_host.host-1.id
}

resource "awx_group_host" "grp-host-link-2" {
  group_id = awx_group.group-example.id
  host_id = awx_host.host-2.id
}

	`, acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)))
}

func testAccGrpHst1stPassGrp2() string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_group" "group-example-2" {
  name        = "%s"
  description = "A second group example."
  inventory   = awx_inventory.example.id
}

resource "awx_group_host" "grp2-host-link" {
  group_id = awx_group.group-example-2.id
  host_id = awx_host.host-2.id
}	
	`, acctest.RandString(5)))
}

func testAccGrpHst2ndPassGrp2() string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_group" "group-example-2" {
  name        = "%s"
  description = "A second group example."
  inventory   = awx_inventory.example.id
}

resource "awx_group_host" "grp2-host-link" {
  group_id = awx_group.group-example-2.id
  host_id = awx_host.host-1.id
}	
	`, acctest.RandString(5)))
}
