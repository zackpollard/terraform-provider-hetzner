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

func TestUnitServerOrderProductsDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/order/server/product" {
			_ = json.NewEncoder(w).Encode([]serverOrderProductAPIResponse{
				{Product: serverOrderProductAPI{
					ID:          "DS 3000",
					Name:        "Dedicated Server DS 3000",
					Description: []string{"Intel Core i7-6700", "64 GB RAM"},
					Traffic:     "20 TB",
					Dist:        []string{"Rescue System", "linux"},
					Arch:        []int64{3},
					Lang:        []string{"en"},
					Location:    []string{"FSN", "NBG"},
					Prices: []serverOrderProductPriceAPI{
						{
							Location:   "FSN",
							Price:      priceAmountAPI{"25.2100", "30.0000"},
							PriceSetup: priceAmountAPI{"25.2100", "30.0000"},
						},
					},
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
				Config: `data "hetzner_server_order_products" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.id", "DS 3000"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.name", "Dedicated Server DS 3000"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.traffic", "20 TB"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.description.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.description.0", "Intel Core i7-6700"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.location.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.location.0", "FSN"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.prices.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.prices.0.location", "FSN"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.prices.0.price_net", "25.2100"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.prices.0.price_gross", "30.0000"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.orderable_addons.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.orderable_addons.0.id", "primary_ipv4"),
					resource.TestCheckResourceAttr("data.hetzner_server_order_products.test", "products.0.orderable_addons.0.name", "Primary IPv4"),
				),
			},
		},
	})
}
