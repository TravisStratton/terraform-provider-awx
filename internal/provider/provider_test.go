package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"awx": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TOWER_HOST"); v == "" {
		t.Fatal("TOWER_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("TOWER_OAUTH_TOKEN"); v == "" {
		t.Fatal("TOWER_OAUTH_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("TOWER_PLATFORM"); v == "" {
		t.Fatal("TOWER_PLATFORM must be set for acceptance tests")
	}
}
