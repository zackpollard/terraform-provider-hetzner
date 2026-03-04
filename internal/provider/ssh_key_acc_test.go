// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"golang.org/x/crypto/ssh"
)

// testAccGenerateSSHKey generates a real ed25519 SSH public key for testing.
func testAccGenerateSSHKey(t *testing.T, comment string) string {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key: %s", err)
	}
	_ = pem.Block{} // satisfy import if needed
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("Failed to create SSH signer: %s", err)
	}
	pubKey := ssh.MarshalAuthorizedKey(signer.PublicKey())
	// ssh.MarshalAuthorizedKey adds a trailing newline; trim it and add comment.
	return fmt.Sprintf("%s %s", pubKey[:len(pubKey)-1], comment)
}

// TestAccSSHKey_CRUD tests full lifecycle: create, read, update name, import, and destroy.
// SSH keys are free to create/delete, so this always runs with TF_ACC.
func TestAccSSHKey_CRUD(t *testing.T) {
	pubKey := testAccGenerateSSHKey(t, "acc-test")

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
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["hetzner_ssh_key.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return rs.Primary.Attributes["fingerprint"], nil
				},
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
	pubKey := testAccGenerateSSHKey(t, "acc-ds-test")

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
					resource.TestCheckResourceAttrSet("data.hetzner_ssh_keys.all", "ssh_keys.#"),
				),
			},
		},
	})
}

// TestAccSSHKey_DataSource reads a single SSH key via the singular data source.
func TestAccSSHKey_DataSource(t *testing.T) {
	pubKey := testAccGenerateSSHKey(t, "acc-singular-ds-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a key, then read it back via the singular data source.
				Config: testAccSSHKeyConfig("acc-singular-ds-test-key", pubKey) + `
data "hetzner_ssh_key" "test" {
  fingerprint = hetzner_ssh_key.test.fingerprint
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_ssh_key.test", "name", "acc-singular-ds-test-key"),
					resource.TestCheckResourceAttrSet("data.hetzner_ssh_key.test", "fingerprint"),
					resource.TestCheckResourceAttrSet("data.hetzner_ssh_key.test", "type"),
					resource.TestCheckResourceAttrSet("data.hetzner_ssh_key.test", "size"),
				),
			},
		},
	})
}

// TestAccSSHKey_ReplaceOnKeyChange verifies that changing the key data forces replacement.
func TestAccSSHKey_ReplaceOnKeyChange(t *testing.T) {
	pubKey1 := testAccGenerateSSHKey(t, "key1")
	pubKey2 := testAccGenerateSSHKey(t, "key2")

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
