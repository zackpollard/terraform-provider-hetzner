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

func TestUnitSubnetDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/subnet/10.0.0.0" {
			_ = json.NewEncoder(w).Encode(subnetDetailAPIResponse{
				Subnet: subnetDetailAPI{
					IP: "10.0.0.0", Mask: 24, Gateway: "10.0.0.1",
					ServerIP: "1.2.3.4", ServerNumber: 321,
					Failover: false, Locked: false, TrafficWarnings: true,
					TrafficHourly: 100, TrafficDaily: 1000, TrafficMonthly: 10,
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
				Config: `data "hetzner_subnet" "test" {
					ip = "10.0.0.0"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_subnet.test", "ip", "10.0.0.0"),
					resource.TestCheckResourceAttr("data.hetzner_subnet.test", "mask", "24"),
					resource.TestCheckResourceAttr("data.hetzner_subnet.test", "gateway", "10.0.0.1"),
					resource.TestCheckResourceAttr("data.hetzner_subnet.test", "server_ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_subnet.test", "server_number", "321"),
					resource.TestCheckResourceAttr("data.hetzner_subnet.test", "failover", "false"),
				),
			},
		},
	})
}

func TestUnitSubnetsDataSource_list(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/subnet" {
			_ = json.NewEncoder(w).Encode([]subnetListAPIResponse{
				{Subnet: subnetDetailAPI{
					IP: "10.0.0.0", Mask: 24, Gateway: "10.0.0.1",
					ServerIP: "1.2.3.4", ServerNumber: 321,
					Failover: false, Locked: false, TrafficWarnings: true,
					TrafficHourly: 100, TrafficDaily: 1000, TrafficMonthly: 10,
				}},
				{Subnet: subnetDetailAPI{
					IP: "172.16.0.0", Mask: 28, Gateway: "172.16.0.1",
					ServerIP: "5.6.7.8", ServerNumber: 322,
					Failover: true, Locked: false, TrafficWarnings: false,
					TrafficHourly: 0, TrafficDaily: 0, TrafficMonthly: 0,
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
				Config: `data "hetzner_subnets" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_subnets.test", "subnets.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_subnets.test", "subnets.0.ip", "10.0.0.0"),
					resource.TestCheckResourceAttr("data.hetzner_subnets.test", "subnets.0.mask", "24"),
					resource.TestCheckResourceAttr("data.hetzner_subnets.test", "subnets.1.ip", "172.16.0.0"),
					resource.TestCheckResourceAttr("data.hetzner_subnets.test", "subnets.1.failover", "true"),
				),
			},
		},
	})
}
