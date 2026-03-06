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
