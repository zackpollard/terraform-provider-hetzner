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

func newTestWOLServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/wol/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"wol": map[string]interface{}{
				"server_ip":       "1.2.3.4",
				"server_ipv6_net": "2a01:4f8::/64",
				"server_number":   123,
			},
		})
	})

	return httptest.NewServer(mux)
}

func TestUnitWOLDataSource(t *testing.T) {
	ts := newTestWOLServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_wol" "test" {
  server_number = 123
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_wol.test", "server_number", "123"),
					resource.TestCheckResourceAttr("data.hetzner_wol.test", "server_ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_wol.test", "server_ipv6_net", "2a01:4f8::/64"),
				),
			},
		},
	})
}
