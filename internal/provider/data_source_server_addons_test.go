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

func TestUnitServerAddonsDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/order/server_addon/321/product" {
			_ = json.NewEncoder(w).Encode([]serverAddonAPIResponse{
				{Addon: serverAddonAPI{
					ID:   "primary_ipv4",
					Name: "Primary IPv4",
					Type: "ipv4",
					Price: []serverOrderProductPriceAPI{
						{
							Location:   "FSN",
							Price:      priceAmountAPI{"1.0000", "1.1900"},
							PriceSetup: priceAmountAPI{"0.0000", "0.0000"},
						},
					},
				}},
				{Addon: serverAddonAPI{
					ID:   "failover",
					Name: "Failover IP",
					Type: "failover",
					Price: []serverOrderProductPriceAPI{
						{
							Location:   "FSN",
							Price:      priceAmountAPI{"2.0000", "2.3800"},
							PriceSetup: priceAmountAPI{"0.0000", "0.0000"},
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
				Config: `data "hetzner_server_addons" "test" {
					server_number = 321
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.0.id", "primary_ipv4"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.0.name", "Primary IPv4"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.0.type", "ipv4"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.0.price.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.0.price.0.location", "FSN"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.0.price.0.price_net", "1.0000"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.0.price.0.price_gross", "1.1900"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.1.id", "failover"),
					resource.TestCheckResourceAttr("data.hetzner_server_addons.test", "addons.1.name", "Failover IP"),
				),
			},
		},
	})
}
