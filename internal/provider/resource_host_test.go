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

func TestAccHostResource(t *testing.T) {
	host1 := HostAPIModel{
		Name:        "test-host-" + acctest.RandString(5),
		Description: "Example with jsonencoded variables for localhost",
		Variables:   "{\"foo\":\"bar\"}",
		Enabled:     true,
	}

	host2 := HostAPIModel{
		Name:        "test-host-" + acctest.RandString(5),
		Description: "Updated example with different variables",
		Variables:   "{\"baz\":\"qux\"}",
		Enabled:     false,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostResourceConfig(host1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(host1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(host1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(host1.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(host1.Enabled),
					),
				},
			},
			{
				ResourceName:      "awx_host.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHostResourceConfig(host2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(host2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(host2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(host2.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(host1.Enabled),
					),
				},
			},
		},
	})
}

func testAccHostResourceConfig(resource HostAPIModel) string {
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
resource "awx_host" "test" {
  name        = "%s"
  description = "%s"
  inventory   = awx_inventory.test.id
  variables   = jsonencode(%s)
}
  `, acctest.RandString(5), acctest.RandString(5), resource.Name, resource.Description, resource.Variables))
}
