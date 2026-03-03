// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// --- Server resource tests ---

// TestAccServer_Rename tests renaming a server and importing it.
func TestAccServer_Rename(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set server name.
			{
				Config: testAccServerConfig(serverNumber, "acc-test-server"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server.test", "server_name", "acc-test-server"),
					resource.TestCheckResourceAttrSet("hetzner_server.test", "server_ip"),
					resource.TestCheckResourceAttrSet("hetzner_server.test", "product"),
					resource.TestCheckResourceAttrSet("hetzner_server.test", "dc"),
					resource.TestCheckResourceAttr("hetzner_server.test", "status", "ready"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_server.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "server_number",
			},
			// Update name.
			{
				Config: testAccServerConfig(serverNumber, "acc-test-server-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server.test", "server_name", "acc-test-server-renamed"),
				),
			},
		},
	})
}

// --- Server data sources ---

// TestAccServer_DataSource reads a single server via data source.
func TestAccServer_DataSource(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig(serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_server.test", "server_ip"),
					resource.TestCheckResourceAttrSet("data.hetzner_server.test", "product"),
					resource.TestCheckResourceAttrSet("data.hetzner_server.test", "dc"),
					resource.TestCheckResourceAttrSet("data.hetzner_server.test", "status"),
				),
			},
		},
	})
}

// TestAccServers_DataSource lists all servers.
func TestAccServers_DataSource(t *testing.T) {
	_ = testAccGetOrCreateServer(t) // Ensure at least one server exists.

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_servers.test", "servers.#"),
				),
			},
		},
	})
}

// --- rDNS tests ---

// TestAccRDNS_CRUD tests creating, updating, and importing an rDNS entry for a server IP.
func TestAccRDNS_CRUD(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)
	serverIP := testAccServerIP(t, serverNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create rDNS.
			{
				Config: testAccRDNSConfig(serverIP, "acc-test.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_rdns.test", "ip", serverIP),
					resource.TestCheckResourceAttr("hetzner_rdns.test", "ptr", "acc-test.example.com"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_rdns.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
			// Update PTR.
			{
				Config: testAccRDNSConfig(serverIP, "acc-test-updated.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_rdns.test", "ptr", "acc-test-updated.example.com"),
				),
			},
		},
	})
}

// TestAccRDNS_DataSource reads an rDNS entry via data source.
func TestAccRDNS_DataSource(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)
	serverIP := testAccServerIP(t, serverNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create rDNS then read via data source.
				Config: testAccRDNSConfig(serverIP, "acc-ds-test.example.com") + "\n" + testAccRDNSDataSourceConfig(serverIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_rdns.test", "ptr"),
				),
			},
		},
	})
}

// --- Boot rescue tests ---

// TestAccBootRescue_ActivateDeactivate tests activating rescue mode and verifying password is returned.
func TestAccBootRescue_ActivateDeactivate(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Activate rescue.
			{
				Config: testAccBootRescueConfig(serverNumber, "linux"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_boot_rescue.test", "os", "linux"),
					resource.TestCheckResourceAttr("hetzner_boot_rescue.test", "active", "true"),
					resource.TestCheckResourceAttrSet("hetzner_boot_rescue.test", "password"),
					resource.TestCheckResourceAttrSet("hetzner_boot_rescue.test", "server_ip"),
				),
			},
			// Destroy will deactivate rescue.
		},
	})
}

// TestAccBootRescue_DataSource reads boot rescue options via data source.
func TestAccBootRescue_DataSource(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBootRescueDataSourceConfig(serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_boot_rescue.test", "server_ip"),
				),
			},
		},
	})
}

// --- IP data source tests ---

// TestAccIP_DataSource reads IP details for a server's main IP.
func TestAccIP_DataSource(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)
	serverIP := testAccServerIP(t, serverNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPDataSourceConfig(serverIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "ip", serverIP),
					resource.TestCheckResourceAttrSet("data.hetzner_ip.test", "server_number"),
				),
			},
		},
	})
}

// TestAccIPs_DataSource lists all IPs.
func TestAccIPs_DataSource(t *testing.T) {
	_ = testAccGetOrCreateServer(t) // Ensure at least one IP exists.

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_ips.test", "ips.#"),
				),
			},
		},
	})
}

// --- Reset and WoL data source tests ---

// TestAccReset_DataSource reads reset options for a server.
func TestAccReset_DataSource(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResetDataSourceConfig(serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_reset.test", "server_number"),
				),
			},
		},
	})
}

// TestAccWOL_DataSource reads WoL availability for a server.
func TestAccWOL_DataSource(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWOLDataSourceConfig(serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_wol.test", "server_number"),
				),
			},
		},
	})
}

// --- Subnets data source ---

// TestAccSubnets_DataSource lists all subnets.
func TestAccSubnets_DataSource(t *testing.T) {
	_ = testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetsDataSourceConfig(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// May have 0 subnets - just verify it doesn't error.
				),
			},
		},
	})
}
