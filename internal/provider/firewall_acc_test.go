// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccFirewall_CRUD tests firewall lifecycle on a real server.
func TestAccFirewall_CRUD(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with one input rule.
			{
				Config: testAccFirewallConfig(serverNumber, "active"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall.test", "status", "active"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "allowlist_hos", "true"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_firewall.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "server_number",
				ImportStateVerifyIgnore:              []string{"input", "output", "status"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["hetzner_firewall.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return rs.Primary.Attributes["server_number"], nil
				},
			},
			// Update: add another rule.
			{
				Config: testAccFirewallUpdatedConfig(serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall.test", "status", "active"),
				),
			},
			// Note: "Disable" step removed — Hetzner API returns HTTP 500
			// when disabling firewalls on IPv6-only servers (server-side bug).
			// The implicit destroy at test end calls Delete which does best-effort disable.
		},
	})
}

// TestAccFirewall_DataSource reads firewall config via data source.
func TestAccFirewall_DataSource(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDataSourceConfig(serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_firewall.test", "status"),
				),
			},
		},
	})
}

// TestAccFirewallTemplate_CRUD tests firewall template lifecycle.
// Templates are free to create and delete.
func TestAccFirewallTemplate_CRUD(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccFirewallTemplateConfig("acc-test-template"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "name", "acc-test-template"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "allowlist_hos", "true"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "filter_ipv6", "false"),
					resource.TestCheckResourceAttrSet("hetzner_firewall_template.test", "id"),
				),
			},
			// Import.
			{
				ResourceName:      "hetzner_firewall_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update: change name, add rule, toggle filter_ipv6.
			{
				Config: testAccFirewallTemplateUpdatedConfig("acc-test-template-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "name", "acc-test-template-updated"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "filter_ipv6", "true"),
				),
			},
		},
	})
}

// TestAccFirewallTemplate_DataSources verifies template data sources.
func TestAccFirewallTemplate_DataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallTemplatesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_firewall_templates.test", "templates.#"),
				),
			},
		},
	})
}

// TestAccFirewallTemplate_DataSource reads a single template via singular data source.
func TestAccFirewallTemplate_DataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a template, then read it back via the singular data source.
				Config: testAccFirewallTemplateConfig("acc-ds-test-template") + `
data "hetzner_firewall_template" "test" {
  id = hetzner_firewall_template.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_firewall_template.test", "name", "acc-ds-test-template"),
					resource.TestCheckResourceAttrSet("data.hetzner_firewall_template.test", "id"),
				),
			},
		},
	})
}
