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

func TestUnitServerOrderResource_create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server_market/transaction":
			// Return transaction with server_number immediately (skip polling).
			serverNum := 999
			_ = json.NewEncoder(w).Encode(orderTransactionAPIResponse{
				Transaction: orderTransactionAPI{
					ID:           "txn-abc-123",
					ServerNumber: &serverNum,
					Status:       "ready",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/999":
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "10.0.0.1", ServerIPv6: "2001:db8::/64",
					ServerNumber: 999, ServerName: "ordered-server",
					Product: "AX42", DC: "FSN1-DC14",
					Traffic: "unlimited", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/999/cancellation":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id = "12345"
					addons     = ["primary_ipv4"]
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_order.test", "server_number", "999"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "server_ip", "10.0.0.1"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "server_name", "ordered-server"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "product", "AX42"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "dc", "FSN1-DC14"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "status", "ready"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "cancelled", "false"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "transaction_id", "txn-abc-123"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "source", "market"),
				),
			},
		},
	})
}

func TestUnitServerOrderResource_standard(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server/transaction":
			serverNum := 888
			_ = json.NewEncoder(w).Encode(orderTransactionAPIResponse{
				Transaction: orderTransactionAPI{
					ID:           "txn-std-456",
					ServerNumber: &serverNum,
					Status:       "ready",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/888":
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "10.0.0.2", ServerIPv6: "2001:db8:1::/64",
					ServerNumber: 888, ServerName: "standard-server",
					Product: "EX44", DC: "NBG1-DC3",
					Traffic: "20 TB", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/888/cancellation":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id = "EX44"
					source     = "standard"
					location   = "NBG1"
					dist       = "Rescue system"
					lang       = "en"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_order.test", "server_number", "888"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "source", "standard"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "product", "EX44"),
				),
			},
		},
	})
}

func TestUnitServerOrderResource_import(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server_market/transaction":
			serverNum := 999
			_ = json.NewEncoder(w).Encode(orderTransactionAPIResponse{
				Transaction: orderTransactionAPI{
					ID:           "txn-abc-123",
					ServerNumber: &serverNum,
					Status:       "ready",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/999":
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "10.0.0.1", ServerIPv6: "2001:db8::/64",
					ServerNumber: 999, ServerName: "ordered-server",
					Product: "AX42", DC: "FSN1-DC14",
					Traffic: "unlimited", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/999/cancellation":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id = "12345"
				}`,
			},
			{
				ResourceName:                         "hetzner_server_order.test",
				ImportState:                          true,
				ImportStateId:                        "999",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "server_number",
				// Input-only fields and transaction_id won't survive import.
				ImportStateVerifyIgnore: []string{"product_id", "source", "test", "transaction_id"},
			},
		},
	})
}

func TestUnitServerOrderResource_delete(t *testing.T) {
	cancelled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server_market/transaction":
			serverNum := 999
			_ = json.NewEncoder(w).Encode(orderTransactionAPIResponse{
				Transaction: orderTransactionAPI{
					ID:           "txn-del-789",
					ServerNumber: &serverNum,
					Status:       "ready",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/999":
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "10.0.0.1", ServerIPv6: "2001:db8::/64",
					ServerNumber: 999, ServerName: "to-delete",
					Product: "AX42", DC: "FSN1-DC14",
					Traffic: "unlimited", Status: "ready",
					Cancelled: cancelled, PaidUntil: "2026-12-31",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/999/cancellation":
			cancelled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id = "12345"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_server_order.test", "server_number", "999"),
			},
			{
				Config:  `# empty - trigger destroy`,
				Destroy: true,
			},
		},
	})
}

func TestUnitServerOrderResource_delete_already_cancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/order/server_market/transaction":
			serverNum := 999
			_ = json.NewEncoder(w).Encode(orderTransactionAPIResponse{
				Transaction: orderTransactionAPI{
					ID:           "txn-del-409",
					ServerNumber: &serverNum,
					Status:       "ready",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/999":
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "10.0.0.1", ServerIPv6: "2001:db8::/64",
					ServerNumber: 999, ServerName: "cancelled-server",
					Product: "AX42", DC: "FSN1-DC14",
					Traffic: "unlimited", Status: "ready",
					Cancelled: true, PaidUntil: "2026-12-31",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/999/cancellation":
			// Already cancelled - return 409.
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"status":  409,
					"code":    "SERVER_ALREADY_CANCELLED",
					"message": "Server is already cancelled",
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
				Config: `resource "hetzner_server_order" "test" {
					product_id = "12345"
				}`,
			},
			{
				Config:  `# empty - trigger destroy`,
				Destroy: true,
			},
		},
	})
}
