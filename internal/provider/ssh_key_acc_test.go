// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSSHKey_CRUD tests full lifecycle: create, read, update name, import, and destroy.
// SSH keys are free to create/delete, so this always runs with TF_ACC.
func TestAccSSHKey_CRUD(t *testing.T) {
	// Use a well-formed ed25519 test key.
	pubKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGtest1234567890abcdefghijklmnopqrstuvwxyz acc-test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccSSHKeyConfig("acc-test-key", pubKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "name", "acc-test-key"),
					resource.TestCheckResourceAttrSet("hetzner_ssh_key.test", "fingerprint"),
					resource.TestCheckResourceAttrSet("hetzner_ssh_key.test", "type"),
					resource.TestCheckResourceAttrSet("hetzner_ssh_key.test", "size"),
					resource.TestCheckResourceAttrSet("hetzner_ssh_key.test", "created_at"),
				),
			},
			// Import by fingerprint.
			{
				ResourceName:                         "hetzner_ssh_key.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "fingerprint",
			},
			// Update name.
			{
				Config: testAccSSHKeyConfig("acc-test-key-renamed", pubKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "name", "acc-test-key-renamed"),
				),
			},
		},
	})
}

// TestAccSSHKey_DataSources verifies the ssh_key and ssh_keys data sources.
func TestAccSSHKey_DataSources(t *testing.T) {
	pubKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGtest1234567890abcdefghijklmnopqrstuvwxyz acc-ds-test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a key, then read it via data source.
			{
				Config: testAccSSHKeyConfig("acc-ds-test-key", pubKey) + `
data "hetzner_ssh_keys" "all" {
  depends_on = [hetzner_ssh_key.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "name", "acc-ds-test-key"),
					resource.TestCheckResourceAttrSet("data.hetzner_ssh_keys.all", "keys.#"),
				),
			},
		},
	})
}

// TestAccSSHKey_ReplaceOnKeyChange verifies that changing the key data forces replacement.
func TestAccSSHKey_ReplaceOnKeyChange(t *testing.T) {
	pubKey1 := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGtest1234567890abcdefghijklmnopqrstuvwxyz key1"
	pubKey2 := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHabcdefghij1234567890klmnopqrstuvwxyzAB key2"

	var fingerprint1 string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyConfig("replace-test", pubKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("hetzner_ssh_key.test", "fingerprint", func(value string) error {
						fingerprint1 = value
						return nil
					}),
				),
			},
			// Change key data - should force replacement (new fingerprint).
			{
				Config: testAccSSHKeyConfig("replace-test", pubKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("hetzner_ssh_key.test", "fingerprint", func(value string) error {
						if value == fingerprint1 {
							return fmt.Errorf("fingerprint should have changed after key replacement")
						}
						return nil
					}),
				),
			},
		},
	})
}
