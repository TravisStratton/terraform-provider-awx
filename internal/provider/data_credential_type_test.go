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

func TestAccCredentialTypeDataSource(t *testing.T) {
	resource1 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description",
		Kind:        "cloud",
	}
	resource2 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description",
		Kind:        "cloud",
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
				Config: testAccCredentialTypeDataSourceIDConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-id", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-id", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource1.Kind),
					),
				},
			},
			// Read by name testing
			{
				Config: testAccCredentialTypeDataSourceNameConfig(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-name", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource2.Kind),
					),
				},
			},
		},
	})
}

func testAccCredentialTypeDataSourceIDConfig(resource CredentialTypeAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_credential_type" "test-id" {
  name         = "%s"
  description  = "%s"
}
data "awx_credential_type" "test-id" {
  id = awx_credential_type.test-id.id
}
  `, resource.Name, resource.Description))
}

func testAccCredentialTypeDataSourceNameConfig(resource CredentialTypeAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_credential_type" "test-name" {
  name         = "%s"
  description  = "%s"
}
data "awx_credential_type" "test-name" {
  name = awx_credential_type.test-name.name
  kind = awx_credential_type.test-name.kind
}
  `, resource.Name, resource.Description))
}
