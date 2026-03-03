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

func TestUnitVSwitchServerResource_create(t *testing.T) {
	servers := []vSwitchServerEntry{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vswitch/123/server":
			_ = r.ParseForm()
			servers = append(servers, vSwitchServerEntry{
				ServerNumber: 321,
				ServerIP:     "1.2.3.4",
				ServerIPv6:   "2001:db8::/64",
				Status:       "ready",
			})
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/vswitch/123":
			_ = json.NewEncoder(w).Encode(vSwitchDetailAPIResponse{
				ID: 123, Name: "test-vs", Vlan: 4000, Cancelled: false,
				Server: servers,
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/vswitch/123/server":
			servers = []vSwitchServerEntry{}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_vswitch_server" "test" {
					vswitch_id    = 123
					server_number = 321
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_vswitch_server.test", "vswitch_id", "123"),
					resource.TestCheckResourceAttr("hetzner_vswitch_server.test", "server_number", "321"),
					resource.TestCheckResourceAttr("hetzner_vswitch_server.test", "status", "ready"),
				),
			},
		},
	})
}

func TestUnitVSwitchServerResource_import(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vswitch/123/server":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/vswitch/123":
			_ = json.NewEncoder(w).Encode(vSwitchDetailAPIResponse{
				ID: 123, Name: "test-vs", Vlan: 4000, Cancelled: false,
				Server: []vSwitchServerEntry{
					{ServerNumber: 321, ServerIP: "1.2.3.4", ServerIPv6: "2001:db8::/64", Status: "ready"},
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/vswitch/123/server":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_vswitch_server" "test" {
					vswitch_id    = 123
					server_number = 321
				}`,
			},
			{
				ResourceName:                         "hetzner_vswitch_server.test",
				ImportState:                          true,
				ImportStateId:                        "123/321",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vswitch_id",
			},
		},
	})
}
