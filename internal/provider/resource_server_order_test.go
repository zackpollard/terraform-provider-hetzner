// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// newServerOrderMock creates a mock HTTP handler that tracks server name and cancellation state.
func newServerOrderMock() (http.Handler, *serverOrderMockState) {
	state := &serverOrderMockState{
		name:                     "ordered-server",
		earliestCancellationDate: "2026-06-30",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		case r.Method == http.MethodPost && r.URL.Path == "/server/999/cancellation":
			_ = r.ParseForm()
			date := r.FormValue("cancellation_date")
			state.cancellationDate = &date
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "10.0.0.1", ServerNumber: 999,
					EarliestCancellationDate: state.earliestCancellationDate,
					Cancelled: true, CancellationDate: state.cancellationDate,
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/server/999/cancellation":
			state.cancellationDate = nil
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "10.0.0.1", ServerNumber: 999,
					EarliestCancellationDate: state.earliestCancellationDate,
					Cancelled: false, CancellationDate: nil,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/999/cancellation":
			cancelled := state.cancellationDate != nil
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "10.0.0.1", ServerNumber: 999,
					EarliestCancellationDate: state.earliestCancellationDate,
					Cancelled: cancelled, CancellationDate: state.cancellationDate,
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/999":
			_ = r.ParseForm()
			state.name = r.FormValue("server_name")
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "10.0.0.1", ServerIPv6: "2001:db8::/64",
					ServerNumber: 999, ServerName: state.name,
					Product: "AX42", DC: "FSN1-DC14",
					Traffic: "unlimited", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/999":
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "10.0.0.1", ServerIPv6: "2001:db8::/64",
					ServerNumber: 999, ServerName: state.name,
					Product: "AX42", DC: "FSN1-DC14",
					Traffic: "unlimited", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}), state
}

type serverOrderMockState struct {
	name                     string
	cancellationDate         *string
	earliestCancellationDate string
}

func TestUnitServerOrderResource_create(t *testing.T) {
	handler, _ := newServerOrderMock()
	ts := httptest.NewServer(handler)
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
					resource.TestCheckResourceAttr("hetzner_server_order.test", "earliest_cancellation_date", "2026-06-30"),
				),
			},
		},
	})
}

func TestUnitServerOrderResource_create_with_name(t *testing.T) {
	handler, mock := newServerOrderMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id  = "12345"
					server_name = "my-named-server"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_order.test", "server_name", "my-named-server"),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			if mock.name != "my-named-server" {
				return fmt.Errorf("expected server name to be 'my-named-server', got '%s'", mock.name)
			}
			return nil
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
		case r.Method == http.MethodGet && r.URL.Path == "/server/888/cancellation":
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "10.0.0.2", ServerNumber: 888,
					EarliestCancellationDate: "2026-06-30",
					Cancelled: false, CancellationDate: nil,
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

func TestUnitServerOrderResource_rename(t *testing.T) {
	handler, _ := newServerOrderMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id  = "12345"
					server_name = "original-name"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_server_order.test", "server_name", "original-name"),
			},
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id  = "12345"
					server_name = "renamed-server"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_server_order.test", "server_name", "renamed-server"),
			},
		},
	})
}

func TestUnitServerOrderResource_cancellation(t *testing.T) {
	handler, _ := newServerOrderMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			// Step 1: Create without cancellation.
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id  = "12345"
					server_name = "my-server"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_order.test", "server_name", "my-server"),
					resource.TestCheckNoResourceAttr("hetzner_server_order.test", "cancellation_date"),
					resource.TestCheckResourceAttr("hetzner_server_order.test", "earliest_cancellation_date", "2026-06-30"),
				),
			},
			// Step 2: Schedule cancellation.
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id        = "12345"
					server_name       = "my-server"
					cancellation_date = "2026-12-31"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server_order.test", "cancellation_date", "2026-12-31"),
				),
			},
			// Step 3: Revoke cancellation.
			{
				Config: `resource "hetzner_server_order" "test" {
					product_id  = "12345"
					server_name = "my-server"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("hetzner_server_order.test", "cancellation_date"),
				),
			},
		},
	})
}

func TestUnitServerOrderResource_import(t *testing.T) {
	handler, _ := newServerOrderMock()
	ts := httptest.NewServer(handler)
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
				ImportStateVerifyIgnore: []string{"product_id", "source", "test", "transaction_id", "reserve_location"},
			},
		},
	})
}

func TestUnitServerOrderResource_delete(t *testing.T) {
	cancelDate := ""
	handler, _ := newServerOrderMock()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/server/999/cancellation" {
			_ = r.ParseForm()
			cancelDate = r.FormValue("cancellation_date")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		handler.ServeHTTP(w, r)
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
		},
		CheckDestroy: func(s *terraform.State) error {
			// Delete should schedule cancellation at the earliest date, not "now".
			if cancelDate != "2026-06-30" {
				return fmt.Errorf("expected cancellation_date=2026-06-30, got %q", cancelDate)
			}
			return nil
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
		case r.Method == http.MethodGet && r.URL.Path == "/server/999/cancellation":
			cancelDate := "2026-12-31"
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "10.0.0.1", ServerNumber: 999,
					EarliestCancellationDate: "2026-06-30",
					Cancelled: true, CancellationDate: &cancelDate,
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/999/cancellation":
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
