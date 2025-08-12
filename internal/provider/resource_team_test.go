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

func TestAccTeamResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	teamName := "test-team-" + acctest.RandString(5)
	teamDesc := "Test team description"
	teamDesc2 := "Test team description"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamResourceConfig(teamName, teamDesc),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(teamName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(teamDesc),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("aap25_gateway_id"),
						fmt.Sprintf("%s_team.test", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      "awx_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTeamResourceConfig(teamName, teamDesc2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(teamName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(teamDesc2),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("aap25_gateway_id"),
						fmt.Sprintf("%s_team.test", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccTeamResourceConfig(teamName, teamDesc string) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test" {
  name = "%s"
}

resource "awx_team" "test" {
  name         = "%s"
  organization = awx_organization.test.aap25_gateway_id
  description  = "%s"
}
`, acctest.RandString(5), teamName, teamDesc))
}
