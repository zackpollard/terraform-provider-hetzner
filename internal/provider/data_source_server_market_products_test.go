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

func TestUnitServerMarketProductsDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/order/server_market/product" {
			_ = json.NewEncoder(w).Encode([]serverMarketProductAPIResponse{
				{Product: serverMarketProductAPI{
					ID:             1234567,
					Name:           "SB-A",
					Description:    []string{"Intel Core i7-6700", "64 GB RAM"},
					Traffic:        "20 TB",
					Dist:           []string{"Rescue System", "linux"},
					CPU:            "Intel Core i7-6700",
					CPUBenchmark:   9876,
					MemorySize:     64,
					HDDSize:        512,
					HDDText:        "2x SSD 512 GB",
					HDDCount:       2,
					Datacenter:     "FSN1-DC14",
					NetworkSpeed:   "1 Gbit",
					Price:          "25.2100",
					PriceHourly:    "0.0370",
					FixedPrice:     false,
					NextReduceDate: "2026-04-01",
					OrderableAddons: []serverOrderProductAddonAPI{
						{
							ID:   "primary_ipv4",
							Name: "Primary IPv4",
							Min:  0,
							Max:  1,
							Prices: []serverOrderProductPriceAPI{
								{
									Location:   "FSN",
									Price:      priceAmountAPI{"1.0000", "1.1900"},
									PriceSetup: priceAmountAPI{"0.0000", "0.0000"},
								},
							},
						},
					},
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
				Config: `data "hetzner_server_market_products" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.id", "1234567"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.name", "SB-A"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.cpu", "Intel Core i7-6700"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.cpu_benchmark", "9876"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.memory_size", "64"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.hdd_size", "512"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.hdd_text", "2x SSD 512 GB"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.hdd_count", "2"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.datacenter", "FSN1-DC14"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.price", "25.2100"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.fixed_price", "false"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.orderable_addons.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_server_market_products.test", "products.0.orderable_addons.0.id", "primary_ipv4"),
				),
			},
		},
	})
}
