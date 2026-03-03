// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories creates provider factories for acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hetzner": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables are set
// before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if v := os.Getenv("HETZNER_ROBOT_USERNAME"); v == "" {
		t.Fatal("HETZNER_ROBOT_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("HETZNER_ROBOT_PASSWORD"); v == "" {
		t.Fatal("HETZNER_ROBOT_PASSWORD must be set for acceptance tests")
	}
}
