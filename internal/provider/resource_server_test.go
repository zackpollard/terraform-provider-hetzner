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

// newServerMock creates a mock HTTP handler that tracks server name and cancellation state.
func newServerMock() (http.Handler, *serverMockState) {
	state := &serverMockState{
		name:                     "my-server",
		earliestCancellationDate: "2026-06-30",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/server/321/cancellation":
			_ = r.ParseForm()
			date := r.FormValue("cancellation_date")
			state.cancellationDate = &date
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "1.2.3.4", ServerNumber: 321,
					EarliestCancellationDate: state.earliestCancellationDate,
					Cancelled: true, CancellationDate: state.cancellationDate,
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/server/321/cancellation":
			state.cancellationDate = nil
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "1.2.3.4", ServerNumber: 321,
					EarliestCancellationDate: state.earliestCancellationDate,
					Cancelled: false, CancellationDate: nil,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/321/cancellation":
			cancelled := state.cancellationDate != nil
			_ = json.NewEncoder(w).Encode(serverCancellationAPIResponse{
				Cancellation: serverCancellationAPI{
					ServerIP: "1.2.3.4", ServerNumber: 321,
					EarliestCancellationDate: state.earliestCancellationDate,
					Cancelled: cancelled, CancellationDate: state.cancellationDate,
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/server/321":
			_ = r.ParseForm()
			state.name = r.FormValue("server_name")
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "1.2.3.4", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ServerName: state.name,
					Product: "DS 3000", DC: "FSN1-DC14",
					Traffic: "5 TB", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/server/321":
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "1.2.3.4", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ServerName: state.name,
					Product: "DS 3000", DC: "FSN1-DC14",
					Traffic: "5 TB", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}), state
}

type serverMockState struct {
	name                     string
	cancellationDate         *string
	earliestCancellationDate string
}

func TestUnitServerResource_create(t *testing.T) {
	handler, _ := newServerMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server" "test" {
					server_number = 321
					server_name   = "my-server"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server.test", "server_number", "321"),
					resource.TestCheckResourceAttr("hetzner_server.test", "server_name", "my-server"),
					resource.TestCheckResourceAttr("hetzner_server.test", "server_ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("hetzner_server.test", "product", "DS 3000"),
					resource.TestCheckResourceAttr("hetzner_server.test", "dc", "FSN1-DC14"),
					resource.TestCheckResourceAttr("hetzner_server.test", "status", "ready"),
					resource.TestCheckResourceAttr("hetzner_server.test", "earliest_cancellation_date", "2026-06-30"),
				),
			},
		},
	})
}

func TestUnitServerResource_update(t *testing.T) {
	handler, _ := newServerMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server" "test" {
					server_number = 321
					server_name   = "my-server"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_server.test", "server_name", "my-server"),
			},
			{
				Config: `resource "hetzner_server" "test" {
					server_number = 321
					server_name   = "renamed-server"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_server.test", "server_name", "renamed-server"),
			},
		},
	})
}

func TestUnitServerResource_import(t *testing.T) {
	handler, _ := newServerMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server" "test" {
					server_number = 321
					server_name   = "my-server"
				}`,
			},
			{
				ResourceName:                         "hetzner_server.test",
				ImportState:                          true,
				ImportStateId:                        "321",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "server_number",
				ImportStateVerifyIgnore:              []string{"reserve_location"},
			},
		},
	})
}

func TestUnitServerResource_cancellation(t *testing.T) {
	handler, _ := newServerMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			// Step 1: Create server without cancellation.
			{
				Config: `resource "hetzner_server" "test" {
					server_number = 321
					server_name   = "my-server"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server.test", "server_name", "my-server"),
					resource.TestCheckNoResourceAttr("hetzner_server.test", "cancellation_date"),
					resource.TestCheckResourceAttr("hetzner_server.test", "earliest_cancellation_date", "2026-06-30"),
				),
			},
			// Step 2: Schedule cancellation.
			{
				Config: `resource "hetzner_server" "test" {
					server_number     = 321
					server_name       = "my-server"
					cancellation_date = "2026-12-31"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_server.test", "cancellation_date", "2026-12-31"),
					resource.TestCheckResourceAttr("hetzner_server.test", "earliest_cancellation_date", "2026-06-30"),
				),
			},
			// Step 3: Revoke cancellation by removing the attribute.
			{
				Config: `resource "hetzner_server" "test" {
					server_number = 321
					server_name   = "my-server"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("hetzner_server.test", "cancellation_date"),
				),
			},
		},
	})
}

func TestUnitServerResource_delete_cancels(t *testing.T) {
	handler, mock := newServerMock()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_server" "test" {
					server_number = 321
					server_name   = "my-server"
				}`,
			},
		},
		// After destroy, the mock should have received a cancellation POST.
		CheckDestroy: func(s *terraform.State) error {
			if mock.cancellationDate == nil {
				return fmt.Errorf("expected cancellation to be set after destroy")
			}
			if *mock.cancellationDate != "now" {
				return fmt.Errorf("expected cancellation_date=now, got %s", *mock.cancellationDate)
			}
			return nil
		},
	})
}

func TestUnitServerDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/server/321" {
			_ = json.NewEncoder(w).Encode(serverDetailAPIResponse{
				Server: serverDetailAPI{
					ServerIP: "1.2.3.4", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ServerName: "my-server",
					Product: "DS 3000", DC: "FSN1-DC14",
					Traffic: "5 TB", Status: "ready",
					Cancelled: false, PaidUntil: "2026-12-31",
				},
			})
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_server" "test" {
					server_number = 321
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_server.test", "server_name", "my-server"),
					resource.TestCheckResourceAttr("data.hetzner_server.test", "server_ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_server.test", "product", "DS 3000"),
				),
			},
		},
	})
}

func TestUnitServersDataSource_list(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/server" {
			_ = json.NewEncoder(w).Encode([]serverListAPIResponse{
				{Server: serverListAPI{
					ServerIP: "1.2.3.4", ServerIPv6: "2001:db8::/64", ServerNumber: 321,
					ServerName: "server-1", Product: "DS 3000", DC: "FSN1-DC14",
					Traffic: "5 TB", Status: "ready", Cancelled: false, PaidUntil: "2026-12-31",
				}},
				{Server: serverListAPI{
					ServerIP: "5.6.7.8", ServerIPv6: "2001:db8:1::/64", ServerNumber: 322,
					ServerName: "server-2", Product: "EX 42", DC: "NBG1-DC3",
					Traffic: "20 TB", Status: "ready", Cancelled: false, PaidUntil: "2026-06-30",
				}},
			})
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_servers" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_servers.test", "servers.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_servers.test", "servers.0.server_name", "server-1"),
					resource.TestCheckResourceAttr("data.hetzner_servers.test", "servers.1.server_name", "server-2"),
				),
			},
		},
	})
}
