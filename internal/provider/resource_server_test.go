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
