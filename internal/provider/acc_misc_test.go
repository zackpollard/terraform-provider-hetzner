// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// --- Traffic data source ---

func TestAccTrafficDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	serverNumber := testAccGetOrCreateServer(t)
	serverIP := testAccServerIP(t, serverNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "hetzner_traffic" "test" {
					ips  = [%q]
					from = "2025-01"
					to   = "2025-02"
					type = "year"
				}`, serverIP),
				Check: resource.TestCheckResourceAttrSet("data.hetzner_traffic.test", "data.#"),
			},
		},
	})
}

// --- rDNS list data source ---

func TestAccRDNSListDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_rdns_list" "test" {
				}`,
				Check: resource.TestCheckResourceAttrSet("data.hetzner_rdns_list.test", "entries.#"),
			},
		},
	})
}

func TestAccRDNSListDataSource_filterByServerIP(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	serverNumber := testAccGetOrCreateServer(t)
	serverIP := testAccServerIP(t, serverNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "hetzner_rdns_list" "test" {
					server_ip = %q
				}`, serverIP),
				Check: resource.TestCheckResourceAttrSet("data.hetzner_rdns_list.test", "entries.#"),
			},
		},
	})
}

// --- Server market products data source ---

func TestAccServerMarketProductsDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_server_market_products" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.hetzner_server_market_products.test",
						tfjsonpath.New("products"),
						knownvalue.NotNull()),
				},
			},
		},
	})
}

// --- Server order products data source ---

func TestAccServerOrderProductsDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_server_order_products" "test" {
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.hetzner_server_order_products.test",
						tfjsonpath.New("products"),
						knownvalue.NotNull()),
				},
			},
		},
	})
}

// --- Server order (market/auction) ---

// TestAccServerOrder_Market orders the cheapest hourly auction server via hetzner_server_order.
// Gated behind HETZNER_TEST_SERVER_ORDER=1 since it costs money.
func TestAccServerOrder_Market(t *testing.T) {
	if os.Getenv("HETZNER_TEST_SERVER_ORDER") != "1" {
		t.Skip("HETZNER_TEST_SERVER_ORDER not set to 1; skipping (this test orders and cancels a server)")
	}

	productID := fmt.Sprintf("%d", testAccFindCheapestServer(t))
	keyFingerprint := testAccGetSSHKeyFingerprint(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerOrderMarketConfig(productID, keyFingerprint),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_order.test", "source", "market"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "server_number"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "server_ip"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "product"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "dc"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "status", "ready"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "cancelled", "false"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "transaction_id"),
				),
			},
			// Destroy cancels the server.
		},
	})
}

// --- Server order (standard) ---

// TestAccServerOrder_Standard orders the cheapest standard server with no setup fee
// via hetzner_server_order. Gated behind HETZNER_TEST_SERVER_ORDER=1 since it costs money.
func TestAccServerOrder_Standard(t *testing.T) {
	if os.Getenv("HETZNER_TEST_SERVER_ORDER") != "1" {
		t.Skip("HETZNER_TEST_SERVER_ORDER not set to 1; skipping (this test orders and cancels a server)")
	}

	productID := testAccFindCheapestStandardServer(t)
	keyFingerprint := testAccGetSSHKeyFingerprint(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerOrderStandardConfig(productID, keyFingerprint),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_order.test", "source", "standard"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "server_number"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "server_ip"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "product"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "dc"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "status", "ready"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "cancelled", "false"),
					resource.TestCheckResourceAttrSet("hetzner_server_order.test", "transaction_id"),
				),
			},
			// Destroy cancels the server.
		},
	})
}

// --- Server addons data source ---

func TestAccServerAddonsDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	serverNumber := testAccGetOrCreateServer(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "hetzner_server_addons" "test" {
					server_number = %s
				}`, serverNumber),
				Check: resource.TestCheckResourceAttrSet("data.hetzner_server_addons.test", "addons.#"),
			},
		},
	})
}
