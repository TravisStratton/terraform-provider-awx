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

func TestAccCredentialInputSourcesResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	orgName := acctest.RandString(8)
	srcCredName := acctest.RandString(8)
	tgtCredName := acctest.RandString(8)
	secKey := "secret/tower/username"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1 - basic test with multiple input sources
			{
				Config: testAccCredInputSrcStep1Config(orgName, srcCredName, tgtCredName, secKey),
				ConfigStateChecks: []statecheck.StateCheck{
					// first input source
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact("Testing create"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("input_field_name"),
						knownvalue.StringExact("ssh_key_data"),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_target_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("target_credential"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_source_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("source_credential"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("metadata"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_path":      knownvalue.StringExact(""),
							"secret_backend": knownvalue.StringExact(""),
							"secret_key":     knownvalue.StringExact("acce-ansible"),
							"secret_path":    knownvalue.StringExact("secret/tower/deploy-keys"),
							"secret_version": knownvalue.StringExact(""),
						}),
					),
					// second input source
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact("Testing create2"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("input_field_name"),
						knownvalue.StringExact("username"),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_target_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("target_credential"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_source_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("source_credential"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("metadata"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_path":      knownvalue.StringExact(""),
							"secret_backend": knownvalue.StringExact(""),
							"secret_key":     knownvalue.StringExact("acce-ansible-2"),
							"secret_path":    knownvalue.StringExact(secKey),
							"secret_version": knownvalue.StringExact(""),
						}),
					),
				},
			}, // end Step 1
			// Step 2
			{
				ResourceName:      "awx_credential_input_sources.example_hashi_cred_input_src",
				ImportState:       true,
				ImportStateVerify: true,
			}, // end Step 2
			// Step 3 - test change to input source metadata
			{
				Config: testAccCredInputSrcStep1Config(orgName, srcCredName, tgtCredName, secKey+"_2"),
				ConfigStateChecks: []statecheck.StateCheck{
					// first input source
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact("Testing create"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("input_field_name"),
						knownvalue.StringExact("ssh_key_data"),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_target_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("target_credential"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_source_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("source_credential"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src", configprefix.Prefix),
						tfjsonpath.New("metadata"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_path":      knownvalue.StringExact(""),
							"secret_backend": knownvalue.StringExact(""),
							"secret_key":     knownvalue.StringExact("acce-ansible"),
							"secret_path":    knownvalue.StringExact("secret/tower/deploy-keys"),
							"secret_version": knownvalue.StringExact(""),
						}),
					),
					// second input source
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact("Testing create2"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("input_field_name"),
						knownvalue.StringExact("username"),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_target_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("target_credential"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.example_hashi_source_cred", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("source_credential"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_input_sources.example_hashi_cred_input_src_2", configprefix.Prefix),
						tfjsonpath.New("metadata"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"auth_path":      knownvalue.StringExact(""),
							"secret_backend": knownvalue.StringExact(""),
							"secret_key":     knownvalue.StringExact("acce-ansible-2"),
							"secret_path":    knownvalue.StringExact(secKey + "_2"),
							"secret_version": knownvalue.StringExact(""),
						}),
					),
				},
			}, // end Step 3
		}, // end all Steps
	}) // close TestCase structure & resource.Test() function all
} // end func TestAccCredentialInputSourcesResource

func testAccCredInputSrcStep1Config(orgName, srcCredName, tgtCredName, secretKey string) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "example" {
  name        = "%s"
  description = "example"
}

data "awx_credential_type" "hashi_vault_secret_lookup" {
    name = "HashiCorp Vault Secret Lookup"
    kind = "external"
}

data "awx_credential_type" "source_control" {
    name = "Source Control"
    kind = "scm"
}

resource "awx_credential" "example_hashi_source_cred" {
  credential_type = data.awx_credential_type.hashi_vault_secret_lookup.id
  name            = "%s"
  organization    = awx_organization.example.id
  inputs = jsonencode({
        "api_version": "v2",
        "cacert": "",
        "default_auth_path": "approle",
        "kubernetes_role": "",
        "namespace": "fake",
        "role_id": "fakse",
        "secret_id": "fakevalue",
        "url": "https://hcp-vault-private-vault-fake.z1.hashicorp.cloud:8200",
        "username": ""})
}

resource "awx_credential" "example_hashi_target_cred" {
  credential_type = data.awx_credential_type.source_control.id
  name            = "%s"
  organization    = awx_organization.example.id
}

resource "awx_credential_input_sources" "example_hashi_cred_input_src" {
    description = "Testing create"
    input_field_name = "ssh_key_data"
    metadata = {
        "auth_path": ""
        "secret_key": "acce-ansible"
        "secret_path": "secret/tower/deploy-keys"
        "secret_backend": ""
        "secret_version": ""
    }
    target_credential = awx_credential.example_hashi_target_cred.id
    source_credential = awx_credential.example_hashi_source_cred.id
}

resource "awx_credential_input_sources" "example_hashi_cred_input_src_2" {
    description = "Testing create2"
    input_field_name = "username"
    metadata = {
        "auth_path": ""
        "secret_key": "acce-ansible-2"
        "secret_path": "%s"
        "secret_backend": ""
        "secret_version": ""
    }
    target_credential = awx_credential.example_hashi_target_cred.id
    source_credential = awx_credential.example_hashi_source_cred.id
}`, orgName, srcCredName, tgtCredName, secretKey))
}
