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

func newTestResetServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/reset/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"reset": map[string]interface{}{
				"server_ip":        "1.2.3.4",
				"server_ipv6_net":  "2a01:4f8::/64",
				"server_number":    123,
				"type":             []string{"sw", "hw", "man"},
				"operating_status": "running",
			},
		})
	})

	return httptest.NewServer(mux)
}

func TestUnitResetDataSource(t *testing.T) {
	ts := newTestResetServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_reset" "test" {
  server_number = 123
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_reset.test", "server_number", "123"),
					resource.TestCheckResourceAttr("data.hetzner_reset.test", "server_ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_reset.test", "operating_status", "running"),
					resource.TestCheckResourceAttr("data.hetzner_reset.test", "type.#", "3"),
					resource.TestCheckResourceAttr("data.hetzner_reset.test", "type.0", "sw"),
				),
			},
		},
	})
}
