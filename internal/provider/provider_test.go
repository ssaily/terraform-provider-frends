package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories instantiates the provider during acceptance tests.
// The factory is called for each Terraform CLI command executed by the test framework.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"frends": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables are set before any
// acceptance test runs.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("FRENDS_HOST_URL"); v == "" {
		t.Fatal("FRENDS_HOST_URL must be set for acceptance tests")
	}
	if v := os.Getenv("FRENDS_TOKEN"); v == "" {
		t.Fatal("FRENDS_TOKEN must be set for acceptance tests")
	}
}
