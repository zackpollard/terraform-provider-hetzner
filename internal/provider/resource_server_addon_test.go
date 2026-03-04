// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitServerAddonResource_create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server_addon/transaction":
			_ = r.ParseForm()
			sn := 321
			_ = json.NewEncoder(w).Encode(addonTransactionAPIResponse{
				Transaction: addonTransactionAPI{
					ID:           "txn-addon-001",
					Status:       "ready",
					ServerNumber: &sn,
					ProductID:    r.FormValue("product_id"),
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/order/server_addon/transaction/txn-addon-001":
			sn := 321
			_ = json.NewEncoder(w).Encode(addonTransactionAPIResponse{
				Transaction: addonTransactionAPI{
					ID:           "txn-addon-001",
					Status:       "ready",
					ServerNumber: &sn,
					ProductID:    "primary_ipv4",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_addon" "test" {
					server_number = 321
					product_id    = "primary_ipv4"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_addon.test", "transaction_id", "txn-addon-001"),
					resource.TestCheckResourceAttr("hetzner_server_addon.test", "status", "ready"),
					resource.TestCheckResourceAttr("hetzner_server_addon.test", "server_number", "321"),
					resource.TestCheckResourceAttr("hetzner_server_addon.test", "product_id", "primary_ipv4"),
				),
			},
		},
	})
}

func TestUnitServerAddonResource_import(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server_addon/transaction":
			sn := 321
			_ = json.NewEncoder(w).Encode(addonTransactionAPIResponse{
				Transaction: addonTransactionAPI{
					ID:           "txn-addon-002",
					Status:       "ready",
					ServerNumber: &sn,
					ProductID:    "primary_ipv4",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/order/server_addon/transaction/txn-addon-002":
			sn := 321
			_ = json.NewEncoder(w).Encode(addonTransactionAPIResponse{
				Transaction: addonTransactionAPI{
					ID:           "txn-addon-002",
					Status:       "ready",
					ServerNumber: &sn,
					ProductID:    "primary_ipv4",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_addon" "test" {
					server_number = 321
					product_id    = "primary_ipv4"
				}`,
			},
			{
				ResourceName:                         "hetzner_server_addon.test",
				ImportState:                          true,
				ImportStateId:                        "txn-addon-002",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "transaction_id",
			},
		},
	})
}

func TestUnitServerAddonResource_delete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server_addon/transaction":
			sn := 321
			_ = json.NewEncoder(w).Encode(addonTransactionAPIResponse{
				Transaction: addonTransactionAPI{
					ID:           "txn-addon-003",
					Status:       "ready",
					ServerNumber: &sn,
					ProductID:    "primary_ipv4",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/order/server_addon/transaction/txn-addon-003":
			sn := 321
			_ = json.NewEncoder(w).Encode(addonTransactionAPIResponse{
				Transaction: addonTransactionAPI{
					ID:           "txn-addon-003",
					Status:       "ready",
					ServerNumber: &sn,
					ProductID:    "primary_ipv4",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_addon" "test" {
					server_number = 321
					product_id    = "primary_ipv4"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_server_addon.test", "transaction_id", "txn-addon-003"),
			},
			{
				Config:  `# empty - trigger destroy`,
				Destroy: true,
			},
		},
	})
}
