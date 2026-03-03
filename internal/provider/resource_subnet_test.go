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

func TestUnitSubnetResource_create(t *testing.T) {
	trafficWarnings := false
	trafficHourly := 0
	trafficDaily := 0
	trafficMonthly := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/subnet/10.0.0.0":
			_ = r.ParseForm()
			if v := r.FormValue("traffic_warnings"); v == "true" {
				trafficWarnings = true
			} else {
				trafficWarnings = false
			}
			if v := r.FormValue("traffic_hourly"); v != "" {
				_, _ = fmt.Sscanf(v, "%d", &trafficHourly)
			}
			if v := r.FormValue("traffic_daily"); v != "" {
				_, _ = fmt.Sscanf(v, "%d", &trafficDaily)
			}
			if v := r.FormValue("traffic_monthly"); v != "" {
				_, _ = fmt.Sscanf(v, "%d", &trafficMonthly)
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/subnet/10.0.0.0":
			_ = json.NewEncoder(w).Encode(subnetDetailAPIResponse{
				Subnet: subnetDetailAPI{
					IP: "10.0.0.0", Mask: 24, Gateway: "10.0.0.1",
					ServerIP: "1.2.3.4", ServerNumber: 321,
					Failover: false, Locked: false,
					TrafficWarnings: trafficWarnings, TrafficHourly: trafficHourly,
					TrafficDaily: trafficDaily, TrafficMonthly: trafficMonthly,
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
				Config: `resource "hetzner_subnet" "test" {
					ip               = "10.0.0.0"
					traffic_warnings = true
					traffic_hourly   = 100
					traffic_daily    = 1000
					traffic_monthly  = 10
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_subnet.test", "ip", "10.0.0.0"),
					resource.TestCheckResourceAttr("hetzner_subnet.test", "traffic_warnings", "true"),
					resource.TestCheckResourceAttr("hetzner_subnet.test", "traffic_hourly", "100"),
					resource.TestCheckResourceAttr("hetzner_subnet.test", "mask", "24"),
					resource.TestCheckResourceAttr("hetzner_subnet.test", "gateway", "10.0.0.1"),
				),
			},
		},
	})
}

func TestUnitSubnetResource_update(t *testing.T) {
	trafficWarnings := true
	trafficHourly := 100
	trafficDaily := 1000
	trafficMonthly := 10

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/subnet/10.0.0.0":
			_ = r.ParseForm()
			if v := r.FormValue("traffic_warnings"); v == "true" {
				trafficWarnings = true
			} else if v == "false" {
				trafficWarnings = false
			}
			if v := r.FormValue("traffic_hourly"); v != "" {
				_, _ = fmt.Sscanf(v, "%d", &trafficHourly)
			}
			if v := r.FormValue("traffic_daily"); v != "" {
				_, _ = fmt.Sscanf(v, "%d", &trafficDaily)
			}
			if v := r.FormValue("traffic_monthly"); v != "" {
				_, _ = fmt.Sscanf(v, "%d", &trafficMonthly)
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/subnet/10.0.0.0":
			_ = json.NewEncoder(w).Encode(subnetDetailAPIResponse{
				Subnet: subnetDetailAPI{
					IP: "10.0.0.0", Mask: 24, Gateway: "10.0.0.1",
					ServerIP: "1.2.3.4", ServerNumber: 321,
					Failover: false, Locked: false,
					TrafficWarnings: trafficWarnings, TrafficHourly: trafficHourly,
					TrafficDaily: trafficDaily, TrafficMonthly: trafficMonthly,
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
				Config: `resource "hetzner_subnet" "test" {
					ip               = "10.0.0.0"
					traffic_warnings = true
					traffic_hourly   = 100
					traffic_daily    = 1000
					traffic_monthly  = 10
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_subnet.test", "traffic_hourly", "100"),
			},
			{
				Config: `resource "hetzner_subnet" "test" {
					ip               = "10.0.0.0"
					traffic_warnings = false
					traffic_hourly   = 200
					traffic_daily    = 2000
					traffic_monthly  = 20
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_subnet.test", "traffic_warnings", "false"),
					resource.TestCheckResourceAttr("hetzner_subnet.test", "traffic_hourly", "200"),
				),
			},
		},
	})
}

func TestUnitSubnetResource_import(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/subnet/10.0.0.0":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/subnet/10.0.0.0":
			_ = json.NewEncoder(w).Encode(subnetDetailAPIResponse{
				Subnet: subnetDetailAPI{
					IP: "10.0.0.0", Mask: 24, Gateway: "10.0.0.1",
					ServerIP: "1.2.3.4", ServerNumber: 321,
					Failover: false, Locked: false,
					TrafficWarnings: true, TrafficHourly: 100,
					TrafficDaily: 1000, TrafficMonthly: 10,
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
				Config: `resource "hetzner_subnet" "test" {
					ip               = "10.0.0.0"
					traffic_warnings = true
					traffic_hourly   = 100
					traffic_daily    = 1000
					traffic_monthly  = 10
				}`,
			},
			{
				ResourceName:                         "hetzner_subnet.test",
				ImportState:                          true,
				ImportStateId:                        "10.0.0.0",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}
