package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccUserDataSource(t *testing.T) {
	if os.Getenv("TOWER_PLATFORM") == "awx" || os.Getenv("TOWER_PLATFORM") == "aap2.4" {
		resource1 := UserAPIModel{
			Username:    "test-user-" + acctest.RandString(5),
			FirstName:   "test-firstname",
			LastName:    "test-lastname",
			Email:       "test@example.com",
			Password:    acctest.RandString(20),
			IsSuperuser: true,
		}
		resource2 := UserAPIModel{
			Username:        "test-user-" + acctest.RandString(5),
			FirstName:       "test-firstname",
			LastName:        "test-lastname",
			Email:           "test@example.com",
			Password:        acctest.RandString(20),
			IsSystemAuditor: true,
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
					Config: testAccUserDataSourceIdConfig(resource1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource1.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource1.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource1.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource1.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource1.IsSuperuser),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("is_system_auditor"),
							knownvalue.Bool(false),
						),
					},
				},
				// Read by name testing
				{
					Config: testAccUserDataSourceNameConfig(resource2),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource2.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource2.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource2.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource2.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(false),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("is_system_auditor"),
							knownvalue.Bool(resource2.IsSystemAuditor),
						),
					},
				},
			},
		})
	} else {
		resource1 := UserAPIModel{
			Username:    "test-user-" + acctest.RandString(5),
			FirstName:   "test-firstname",
			LastName:    "test-lastname",
			Email:       "test@example.com",
			Password:    acctest.RandString(20),
			IsSuperuser: true,
		}
		resource2 := UserAPIModel{
			Username:  "test-user-" + acctest.RandString(5),
			FirstName: "test-firstname",
			LastName:  "test-lastname",
			Email:     "test@example.com",
			Password:  acctest.RandString(20),
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
					Config: testAccUserDataSourceIdConfig(resource1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource1.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource1.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource1.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource1.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-id", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource1.IsSuperuser),
						),
					},
				},
				// Read by name testing
				{
					Config: testAccUserDataSourceNameConfig(resource2),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource2.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource2.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource2.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource2.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("data.%s_user.test-name", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(false),
						),
					},
				},
			},
		})
	}
}

func testAccUserDataSourceIdConfig(resource UserAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_user" "test-id" {
  username      = "%s"
  first_name 	= "%s"
  last_name 	= "%s"
  email			= "%s"
  password 		= "%s"
  is_superuser  = %v
}
data "awx_user" "test-id" {
  id = awx_user.test-id.id
}
`, resource.Username, resource.FirstName, resource.LastName, resource.Email, resource.Password, resource.IsSuperuser))
}

func testAccUserDataSourceNameConfig(resource UserAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_user" "test-name" {
  username      	= "%s"
  first_name 		= "%s"
  last_name 		= "%s"
  email				= "%s"
  password 			= "%s"
  is_system_auditor = %v
}
data "awx_user" "test-name" {
  username = awx_user.test-name.username
}
`, resource.Username, resource.FirstName, resource.LastName, resource.Email, resource.Password, resource.IsSystemAuditor))
}
