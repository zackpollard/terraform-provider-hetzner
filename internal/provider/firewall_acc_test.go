// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttrSet("hetzner_firewall.test", "server_ip"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_firewall.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "server_number",
			},
			// Update: add another rule.
			{
				Config: testAccFirewallUpdatedConfig(serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall.test", "status", "active"),
				),
			},
			// Disable the firewall.
			{
				Config: testAccFirewallConfig(serverNumber, "disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall.test", "status", "disabled"),
				),
			},
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
