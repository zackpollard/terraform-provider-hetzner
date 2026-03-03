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

func TestUnitFailoverResource_create(t *testing.T) {
	activeIP := "10.0.0.1"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/failover/192.168.1.1":
			_ = r.ParseForm()
			activeIP = r.FormValue("active_server_ip")
			_ = json.NewEncoder(w).Encode(failoverAPIResponse{
				Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: activeIP,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/failover/192.168.1.1":
			_ = json.NewEncoder(w).Encode(failoverAPIResponse{
				Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: activeIP,
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/failover/192.168.1.1":
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
				Config: `resource "hetzner_failover" "test" {
					ip               = "192.168.1.1"
					active_server_ip = "10.0.0.1"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_failover.test", "ip", "192.168.1.1"),
					resource.TestCheckResourceAttr("hetzner_failover.test", "active_server_ip", "10.0.0.1"),
					resource.TestCheckResourceAttr("hetzner_failover.test", "netmask", "255.255.255.255"),
					resource.TestCheckResourceAttr("hetzner_failover.test", "server_number", "321"),
				),
			},
		},
	})
}

func TestUnitFailoverResource_update(t *testing.T) {
	activeIP := "10.0.0.1"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/failover/192.168.1.1":
			_ = r.ParseForm()
			activeIP = r.FormValue("active_server_ip")
			_ = json.NewEncoder(w).Encode(failoverAPIResponse{
				Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: activeIP,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/failover/192.168.1.1":
			_ = json.NewEncoder(w).Encode(failoverAPIResponse{
				Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: activeIP,
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/failover/192.168.1.1":
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
				Config: `resource "hetzner_failover" "test" {
					ip               = "192.168.1.1"
					active_server_ip = "10.0.0.1"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_failover.test", "active_server_ip", "10.0.0.1"),
			},
			{
				Config: `resource "hetzner_failover" "test" {
					ip               = "192.168.1.1"
					active_server_ip = "10.0.0.2"
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_failover.test", "active_server_ip", "10.0.0.2"),
			},
		},
	})
}

func TestUnitFailoverResource_import(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/failover/192.168.1.1":
			_ = json.NewEncoder(w).Encode(failoverAPIResponse{
				Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: "10.0.0.1",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/failover/192.168.1.1":
			_ = json.NewEncoder(w).Encode(failoverAPIResponse{
				Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: "10.0.0.1",
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/failover/192.168.1.1":
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
				Config: `resource "hetzner_failover" "test" {
					ip               = "192.168.1.1"
					active_server_ip = "10.0.0.1"
				}`,
			},
			{
				ResourceName:                         "hetzner_failover.test",
				ImportState:                          true,
				ImportStateId:                        "192.168.1.1",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}

func TestUnitFailoverDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/failover/192.168.1.1" {
			_ = json.NewEncoder(w).Encode(failoverAPIResponse{
				Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: "10.0.0.1",
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
				Config: `data "hetzner_failover" "test" {
					ip = "192.168.1.1"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_failover.test", "active_server_ip", "10.0.0.1"),
					resource.TestCheckResourceAttr("data.hetzner_failover.test", "netmask", "255.255.255.255"),
					resource.TestCheckResourceAttr("data.hetzner_failover.test", "server_number", "321"),
				),
			},
		},
	})
}

func TestUnitFailoversDataSource_list(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/failover" {
			_ = json.NewEncoder(w).Encode([]failoverAPIResponse{
				{Failover: failoverAPI{
					IP: "192.168.1.1", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.100", ServerIPv6: "2001:db8::/64",
					ServerNumber: 321, ActiveServerIP: "10.0.0.1",
				}},
				{Failover: failoverAPI{
					IP: "192.168.1.2", Netmask: "255.255.255.255",
					ServerIP: "10.0.0.101", ServerIPv6: "2001:db8:1::/64",
					ServerNumber: 322, ActiveServerIP: "10.0.0.2",
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
				Config: `data "hetzner_failovers" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_failovers.test", "failovers.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_failovers.test", "failovers.0.ip", "192.168.1.1"),
					resource.TestCheckResourceAttr("data.hetzner_failovers.test", "failovers.1.ip", "192.168.1.2"),
				),
			},
		},
	})
}
