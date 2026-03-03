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

func newTestBootWindowsServer() *httptest.Server {
	active := false
	mux := http.NewServeMux()

	mux.HandleFunc("/boot/123/windows", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			active = true
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"windows": map[string]interface{}{
					"server_ip":       "1.2.3.4",
					"server_ipv6_net": "2a01:4f8::/64",
					"server_number":   123,
					"dist":            "Windows Server 2022",
					"lang":            "en",
					"active":          true,
					"password":        "win-pass-123",
				},
			})
		case http.MethodGet:
			if active {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"windows": map[string]interface{}{
						"server_ip":       "1.2.3.4",
						"server_ipv6_net": "2a01:4f8::/64",
						"server_number":   123,
						"dist":            "Windows Server 2022",
						"lang":            "en",
						"active":          true,
						"password":        nil,
					},
				})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"windows": map[string]interface{}{
						"server_ip":       "1.2.3.4",
						"server_ipv6_net": "2a01:4f8::/64",
						"server_number":   123,
						"dist":            []string{"Windows Server 2022"},
						"lang":            []string{"en"},
						"active":          false,
						"password":        nil,
					},
				})
			}
		case http.MethodDelete:
			active = false
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"windows": map[string]interface{}{
					"server_ip":       "1.2.3.4",
					"server_ipv6_net": "2a01:4f8::/64",
					"server_number":   123,
					"active":          false,
					"password":        nil,
				},
			})
		}
	})

	return httptest.NewServer(mux)
}

func TestUnitBootWindowsResource_Create(t *testing.T) {
	ts := newTestBootWindowsServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_boot_windows" "test" {
  server_number = 123
  dist          = "Windows Server 2022"
  lang          = "en"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_boot_windows.test", "server_number", "123"),
					resource.TestCheckResourceAttr("hetzner_boot_windows.test", "dist", "Windows Server 2022"),
					resource.TestCheckResourceAttr("hetzner_boot_windows.test", "active", "true"),
				),
			},
		},
	})
}

func TestUnitBootWindowsDataSource(t *testing.T) {
	ts := newTestBootWindowsServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_boot_windows" "test" {
  server_number = 123
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_boot_windows.test", "server_number", "123"),
					resource.TestCheckResourceAttr("data.hetzner_boot_windows.test", "active", "false"),
				),
			},
		},
	})
}
