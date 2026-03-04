// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFailover_DataSource reads failover IP details via data source.
func TestAccFailover_DataSource(t *testing.T) {
	failoverIP := testAccFailoverIP(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFailoverDataSourceConfig(failoverIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_failover.test", "ip", failoverIP),
					resource.TestCheckResourceAttrSet("data.hetzner_failover.test", "server_ip"),
					resource.TestCheckResourceAttrSet("data.hetzner_failover.test", "server_number"),
				),
			},
		},
	})
}

// TestAccFailovers_DataSource lists all failover IPs.
func TestAccFailovers_DataSource(t *testing.T) {
	_ = testAccFailoverIP(t) // Skip if no failover IP configured.

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFailoversDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_failovers.test", "failovers.#"),
				),
			},
		},
	})
}

// TestAccFailover_CRUD tests managing a failover IP resource.
// Requires HETZNER_TEST_FAILOVER_IP and a server with IPv4.
func TestAccFailover_CRUD(t *testing.T) {
	failoverIP := testAccFailoverIP(t)
	serverNumber := testAccGetOrCreateServer(t)
	serverIP := testAccServerIP(t, serverNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create: route failover IP to server.
			{
				Config: testAccFailoverConfig(failoverIP, serverIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_failover.test", "ip", failoverIP),
					resource.TestCheckResourceAttr("hetzner_failover.test", "active_server_ip", serverIP),
					resource.TestCheckResourceAttrSet("hetzner_failover.test", "netmask"),
					resource.TestCheckResourceAttrSet("hetzner_failover.test", "server_number"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_failover.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}
