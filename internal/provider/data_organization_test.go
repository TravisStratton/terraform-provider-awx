package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccOrganizationDataSource(t *testing.T) {
	if os.Getenv("TOWER_PLATFORM") == "awx" || os.Getenv("TOWER_PLATFORM") == "aap2.4" {
		resource1 := OrganizationAPIModel{
			Name:        "test-organization-" + acctest.RandString(5),
			Description: "test description 1",
			DefaultEnv:  1,
			MaxHosts:    100,
		}
		resource2 := OrganizationAPIModel{
			Name:        "test-organization-" + acctest.RandString(5),
			Description: "test description 1",
			DefaultEnv:  1,
			MaxHosts:    100,
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
					Config: testAccOrganizationDataSourceIdConfig1(resource1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-id",
							tfjsonpath.New("name"),
							knownvalue.StringExact(resource1.Name),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-id",
							tfjsonpath.New("description"),
							knownvalue.StringExact(resource1.Description),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-id",
							tfjsonpath.New("default_environment"),
							knownvalue.Int32Exact(int32(resource1.DefaultEnv)),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-id",
							tfjsonpath.New("max_hosts"),
							knownvalue.Int32Exact(int32(resource1.MaxHosts)),
						),
					},
				},
				// Read by name testing
				{
					Config: testAccOrganizationDataSourceNameConfig1(resource2),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-name",
							tfjsonpath.New("name"),
							knownvalue.StringExact(resource2.Name),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-name",
							tfjsonpath.New("description"),
							knownvalue.StringExact(resource2.Description),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-name",
							tfjsonpath.New("default_environment"),
							knownvalue.Int32Exact(int32(resource2.DefaultEnv)),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-name",
							tfjsonpath.New("max_hosts"),
							knownvalue.Int32Exact(int32(resource2.MaxHosts)),
						),
					},
				},
			},
		})
	} else {
		resource1 := OrganizationAPIModel{
			Name:        "test-organization-" + acctest.RandString(5),
			Description: "test description 1",
		}
		resource2 := OrganizationAPIModel{
			Name:        "test-organization-" + acctest.RandString(5),
			Description: "test description 1",
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
					Config: testAccOrganizationDataSourceIdConfig2(resource1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-id",
							tfjsonpath.New("name"),
							knownvalue.StringExact(resource1.Name),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-id",
							tfjsonpath.New("description"),
							knownvalue.StringExact(resource1.Description),
						),
					},
				},
				// Read by name testing
				{
					Config: testAccOrganizationDataSourceNameConfig2(resource2),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-name",
							tfjsonpath.New("name"),
							knownvalue.StringExact(resource2.Name),
						),
						statecheck.ExpectKnownValue(
							"data.awx_organization.test-name",
							tfjsonpath.New("description"),
							knownvalue.StringExact(resource2.Description),
						),
					},
				},
			},
		})
	}
}

func testAccOrganizationDataSourceIdConfig1(resource OrganizationAPIModel) string {
	return fmt.Sprintf(`
resource "awx_organization" "test-id" {
  name        			= "%s"
  description 			= "%s"
  default_environment 	= %d
  max_hosts				= %d
}
data "awx_organization" "test-id" {
  id = awx_organization.test-id.id
}
`, resource.Name, resource.Description, resource.DefaultEnv, resource.MaxHosts)
}

func testAccOrganizationDataSourceIdConfig2(resource OrganizationAPIModel) string {
	return fmt.Sprintf(`
resource "awx_organization" "test-id" {
  name        			= "%s"
  description 			= "%s"
}
data "awx_organization" "test-id" {
  id = awx_organization.test-id.id
}
`, resource.Name, resource.Description)
}

func testAccOrganizationDataSourceNameConfig1(resource OrganizationAPIModel) string {
	return fmt.Sprintf(`
resource "awx_organization" "test-name" {
  name        			= "%s"
  description 			= "%s"
  default_environment 	= %d
  max_hosts				= %d
}
data "awx_organization" "test-name" {
  name = awx_organization.test-name.name
}
`, resource.Name, resource.Description, resource.DefaultEnv, resource.MaxHosts)
}

func testAccOrganizationDataSourceNameConfig2(resource OrganizationAPIModel) string {
	return fmt.Sprintf(`
resource "awx_organization" "test-name" {
  name        			= "%s"
  description 			= "%s"
}
data "awx_organization" "test-name" {
  name = awx_organization.test-name.name
}
`, resource.Name, resource.Description)
}
