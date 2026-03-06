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

func TestUnitTrafficDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/traffic":
			// Return array-style data like the real API does.
			resp := map[string]interface{}{
				"traffic": map[string]interface{}{
					"type": "month",
					"from": "2025-01-01",
					"to":   "2025-01-31",
					"data": map[string]interface{}{
						"1.2.3.4": []map[string]float64{
							{"in": 10.5, "out": 20.3, "sum": 30.8},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_traffic" "test" {
					ips  = ["1.2.3.4"]
					from = "2025-01-01"
					to   = "2025-01-31"
					type = "month"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.0.ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.0.in", "10.5"),
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.0.out", "20.3"),
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.0.sum", "30.8"),
				),
			},
		},
	})
}

func TestUnitTrafficDataSource_multipleEntries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/traffic":
			// Multiple time buckets get summed.
			resp := map[string]interface{}{
				"traffic": map[string]interface{}{
					"type": "month",
					"from": "2025-01-01",
					"to":   "2025-03-01",
					"data": map[string]interface{}{
						"1.2.3.4": []map[string]float64{
							{"in": 10.0, "out": 20.0, "sum": 30.0},
							{"in": 5.0, "out": 10.0, "sum": 15.0},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_traffic" "test" {
					ips  = ["1.2.3.4"]
					from = "2025-01-01"
					to   = "2025-03-01"
					type = "month"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.0.in", "15"),
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.0.out", "30"),
					resource.TestCheckResourceAttr("data.hetzner_traffic.test", "data.0.sum", "45"),
				),
			},
		},
	})
}
