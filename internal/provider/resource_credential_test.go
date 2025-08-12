package provider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccCredentialResource(t *testing.T) {
	// Generate a real ECDSA private key (P256 curve)
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Error generating key:", err)
		return
	}

	// Convert to PKCS8 ASN.1 DER format
	der, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		fmt.Println("Error marshalling private key:", err)
		return
	}

	// Encode as PEM (OpenSSH-compatible format)
	pemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	// Encode the private key in OpenSSH format
	opensshKey := strings.ReplaceAll(string(pemBytes), "\n", "\\n")

	resourceName := "test-credential-container-registry-" + acctest.RandString(5)
	resource1 := CredentialAPIModel{
		Name:        "test-credential-machine-" + acctest.RandString(5),
		Description: "test description 1",
		Inputs:      "{\"become_method\":\"sudo\",\"become_password\":\"ASK\",\"password\":\"test1234\",\"username\":\"awx\"}",
	}
	resource1inputs, ok := resource1.Inputs.(string)
	if !ok {
		return
	}
	resource2 := CredentialAPIModel{
		Name:        "test-credential-source-control-" + acctest.RandString(5),
		Description: "test description 2",
		Inputs:      "{\"ssh_key_data\":\"" + opensshKey + "\",\"username\":\"awx\"}",
	}
	resource2inputs, ok := resource2.Inputs.(string)
	if !ok {
		return
	}
	resource3 := CredentialAPIModel{
		Name:        resourceName,
		Description: "test description 3",
		Inputs:      "{\"host\":\"quay.io\",\"password\":\"test1234\",\"username\":\"test\",\"verify_ssl\":true}",
	}
	resource3inputs, ok := resource3.Inputs.(string)
	if !ok {
		return
	}
	resource4 := CredentialAPIModel{
		Name:        resourceName,
		Description: "test description 4",
		Inputs:      "{\"host\":\"quay.io\",\"password\":\"new4567\",\"username\":\"test2\",\"verify_ssl\":false}",
	}
	resource4inputs, ok := resource4.Inputs.(string)
	if !ok {
		return
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCredential1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-machine", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-machine", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-machine", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource1inputs),
					),
				},
			},
			{
				Config: testAccCredential2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-source-control", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-source-control", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-source-control", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource2inputs),
					),
				},
			},
			{
				Config: testAccCredential3Config(resource3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource3.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource3.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource3inputs),
					),
				},
			},
			{
				ResourceName:            "awx_credential.test-container-registry",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inputs"},
			},
			{
				Config: testAccCredential3Config(resource4),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource4.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource4.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource4inputs),
					),
				},
			},
		},
	})
}

func testAccCredential1Config(resource CredentialAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test-machine" {
  name        = "%s"
}
data "awx_credential_type" "test-machine" {
  name = "Machine"
  kind = "ssh"
}
resource "awx_credential" "test-machine" {
  name            = "%s"
  description	  = "%s"
  organization    = awx_organization.test-machine.id
  credential_type = data.awx_credential_type.test-machine.id
  inputs = jsonencode(%s)
}
  `, acctest.RandString(5), resource.Name, resource.Description, resource.Inputs))
}

func testAccCredential2Config(resource CredentialAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test-source-control" {
  name        = "%s"
}
  data "awx_credential_type" "test-source-control" {
	name = "Source Control"
	kind = "scm"
  }
resource "awx_credential" "test-source-control" {
  name            = "%s"
  description	  = "%s"
  organization    = awx_organization.test-source-control.id
  credential_type = data.awx_credential_type.test-source-control.id
	inputs = jsonencode(%s)
}
  `, acctest.RandString(5), resource.Name, resource.Description, resource.Inputs))
}

func testAccCredential3Config(resource CredentialAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test-container-registry" {
  name        = "%s"
}
data "awx_credential_type" "test-container-registry" {
  name = "Container Registry"
  kind = "registry"
}
resource "awx_credential" "test-container-registry" {
  name            = "%s"
  description	  = "%s"
  organization    = awx_organization.test-container-registry.id
  credential_type = data.awx_credential_type.test-container-registry.id
  inputs = jsonencode(%s)
}
  `, acctest.RandString(5), resource.Name, resource.Description, resource.Inputs))
}
