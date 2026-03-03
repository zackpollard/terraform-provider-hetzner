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

func TestUnitIPDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4" {
			json.NewEncoder(w).Encode(ipDetailAPIResponse{
				IP: ipDetailAPI{
					IP: "1.2.3.4", ServerIP: "10.0.0.1", ServerNumber: 321,
					Locked: false, SeparateMAC: nil, TrafficWarnings: true,
					TrafficHourly: 100, TrafficDaily: 1000, TrafficMonthly: 10,
					Gateway: "1.2.3.1", Mask: 32, Broadcast: "1.2.3.4",
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
				Config: `data "hetzner_ip" "test" {
					ip = "1.2.3.4"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "server_ip", "10.0.0.1"),
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "server_number", "321"),
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "locked", "false"),
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "traffic_warnings", "true"),
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "gateway", "1.2.3.1"),
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "mask", "32"),
					resource.TestCheckResourceAttr("data.hetzner_ip.test", "broadcast", "1.2.3.4"),
				),
			},
		},
	})
}

func TestUnitIPsDataSource_list(t *testing.T) {
	mac := "aa:bb:cc:dd:ee:ff"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/ip" {
			json.NewEncoder(w).Encode([]ipListAPIResponse{
				{IP: ipListAPI{
					IP: "1.2.3.4", ServerIP: "10.0.0.1", ServerNumber: 321,
					Locked: false, SeparateMAC: nil, TrafficWarnings: true,
					TrafficHourly: 100, TrafficDaily: 1000, TrafficMonthly: 10,
				}},
				{IP: ipListAPI{
					IP: "5.6.7.8", ServerIP: "10.0.0.2", ServerNumber: 322,
					Locked: true, SeparateMAC: &mac, TrafficWarnings: false,
					TrafficHourly: 200, TrafficDaily: 2000, TrafficMonthly: 20,
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
				Config: `data "hetzner_ips" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_ips.test", "ips.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_ips.test", "ips.0.ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_ips.test", "ips.1.ip", "5.6.7.8"),
					resource.TestCheckResourceAttr("data.hetzner_ips.test", "ips.1.separate_mac", "aa:bb:cc:dd:ee:ff"),
				),
			},
		},
	})
}
