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
)

func TestUnitIPResource_create(t *testing.T) {
	trafficWarnings := false
	trafficHourly := 0
	trafficDaily := 0
	trafficMonthly := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/ip/1.2.3.4":
			r.ParseForm()
			if v := r.FormValue("traffic_warnings"); v == "true" {
				trafficWarnings = true
			} else {
				trafficWarnings = false
			}
			if v := r.FormValue("traffic_hourly"); v != "" {
				fmt.Sscanf(v, "%d", &trafficHourly)
			}
			if v := r.FormValue("traffic_daily"); v != "" {
				fmt.Sscanf(v, "%d", &trafficDaily)
			}
			if v := r.FormValue("traffic_monthly"); v != "" {
				fmt.Sscanf(v, "%d", &trafficMonthly)
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4":
			json.NewEncoder(w).Encode(ipDetailAPIResponse{
				IP: ipDetailAPI{
					IP: "1.2.3.4", ServerIP: "10.0.0.1", ServerNumber: 321,
					Locked: false, SeparateMAC: nil,
					TrafficWarnings: trafficWarnings, TrafficHourly: trafficHourly,
					TrafficDaily: trafficDaily, TrafficMonthly: trafficMonthly,
					Gateway: "1.2.3.1", Mask: 32, Broadcast: "1.2.3.4",
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
				Config: `resource "hetzner_ip" "test" {
					ip               = "1.2.3.4"
					traffic_warnings = true
					traffic_hourly   = 100
					traffic_daily    = 1000
					traffic_monthly  = 10
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ip.test", "ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_warnings", "true"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_hourly", "100"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_daily", "1000"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_monthly", "10"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "server_ip", "10.0.0.1"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "gateway", "1.2.3.1"),
				),
			},
		},
	})
}

func TestUnitIPResource_update(t *testing.T) {
	trafficWarnings := true
	trafficHourly := 100
	trafficDaily := 1000
	trafficMonthly := 10

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/ip/1.2.3.4":
			r.ParseForm()
			if v := r.FormValue("traffic_warnings"); v == "true" {
				trafficWarnings = true
			} else if v == "false" {
				trafficWarnings = false
			}
			if v := r.FormValue("traffic_hourly"); v != "" {
				fmt.Sscanf(v, "%d", &trafficHourly)
			}
			if v := r.FormValue("traffic_daily"); v != "" {
				fmt.Sscanf(v, "%d", &trafficDaily)
			}
			if v := r.FormValue("traffic_monthly"); v != "" {
				fmt.Sscanf(v, "%d", &trafficMonthly)
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4":
			json.NewEncoder(w).Encode(ipDetailAPIResponse{
				IP: ipDetailAPI{
					IP: "1.2.3.4", ServerIP: "10.0.0.1", ServerNumber: 321,
					Locked: false, SeparateMAC: nil,
					TrafficWarnings: trafficWarnings, TrafficHourly: trafficHourly,
					TrafficDaily: trafficDaily, TrafficMonthly: trafficMonthly,
					Gateway: "1.2.3.1", Mask: 32, Broadcast: "1.2.3.4",
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
				Config: `resource "hetzner_ip" "test" {
					ip               = "1.2.3.4"
					traffic_warnings = true
					traffic_hourly   = 100
					traffic_daily    = 1000
					traffic_monthly  = 10
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_hourly", "100"),
			},
			{
				Config: `resource "hetzner_ip" "test" {
					ip               = "1.2.3.4"
					traffic_warnings = false
					traffic_hourly   = 200
					traffic_daily    = 2000
					traffic_monthly  = 20
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_warnings", "false"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_hourly", "200"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_daily", "2000"),
					resource.TestCheckResourceAttr("hetzner_ip.test", "traffic_monthly", "20"),
				),
			},
		},
	})
}

func TestUnitIPResource_import(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/ip/1.2.3.4":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4":
			json.NewEncoder(w).Encode(ipDetailAPIResponse{
				IP: ipDetailAPI{
					IP: "1.2.3.4", ServerIP: "10.0.0.1", ServerNumber: 321,
					Locked: false, SeparateMAC: nil,
					TrafficWarnings: true, TrafficHourly: 100,
					TrafficDaily: 1000, TrafficMonthly: 10,
					Gateway: "1.2.3.1", Mask: 32, Broadcast: "1.2.3.4",
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
				Config: `resource "hetzner_ip" "test" {
					ip               = "1.2.3.4"
					traffic_warnings = true
					traffic_hourly   = 100
					traffic_daily    = 1000
					traffic_monthly  = 10
				}`,
			},
			{
				ResourceName:                         "hetzner_ip.test",
				ImportState:                          true,
				ImportStateId:                        "1.2.3.4",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}
