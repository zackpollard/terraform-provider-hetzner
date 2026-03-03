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

func newTestBootRescueServer() *httptest.Server {
	password := "rescue-pass-123"
	active := false
	mux := http.NewServeMux()

	mux.HandleFunc("/boot/123/rescue", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			active = true
			json.NewEncoder(w).Encode(map[string]interface{}{
				"rescue": map[string]interface{}{
					"server_ip":       "1.2.3.4",
					"server_ipv6_net": "2a01:4f8::/64",
					"server_number":   123,
					"os":              "linux",
					"active":          true,
					"password":        password,
					"keyboard":        "us",
				},
			})
		case http.MethodGet:
			if active {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"rescue": map[string]interface{}{
						"server_ip":       "1.2.3.4",
						"server_ipv6_net": "2a01:4f8::/64",
						"server_number":   123,
						"os":              "linux",
						"active":          true,
						"password":        nil,
						"keyboard":        "us",
					},
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"rescue": map[string]interface{}{
						"server_ip":       "1.2.3.4",
						"server_ipv6_net": "2a01:4f8::/64",
						"server_number":   123,
						"os":              []string{"linux", "vkvm"},
						"active":          false,
						"password":        nil,
						"keyboard":        "us",
					},
				})
			}
		case http.MethodDelete:
			active = false
			json.NewEncoder(w).Encode(map[string]interface{}{
				"rescue": map[string]interface{}{
					"server_ip":       "1.2.3.4",
					"server_ipv6_net": "2a01:4f8::/64",
					"server_number":   123,
					"os":              []string{"linux", "vkvm"},
					"active":          false,
					"password":        nil,
				},
			})
		}
	})

	return httptest.NewServer(mux)
}

func TestAccBootRescueResource_Create(t *testing.T) {
	ts := newTestBootRescueServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_boot_rescue" "test" {
  server_number = 123
  os            = "linux"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_boot_rescue.test", "server_number", "123"),
					resource.TestCheckResourceAttr("hetzner_boot_rescue.test", "os", "linux"),
					resource.TestCheckResourceAttr("hetzner_boot_rescue.test", "server_ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("hetzner_boot_rescue.test", "active", "true"),
				),
			},
		},
	})
}

func TestAccBootRescueResource_Import(t *testing.T) {
	ts := newTestBootRescueServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_boot_rescue" "test" {
  server_number = 123
  os            = "linux"
}`,
			},
			{
				ResourceName:      "hetzner_boot_rescue.test",
				ImportState:       true,
				ImportStateId:     "123",
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccBootRescueDataSource(t *testing.T) {
	ts := newTestBootRescueServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_boot_rescue" "test" {
  server_number = 123
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_boot_rescue.test", "server_number", "123"),
					resource.TestCheckResourceAttr("data.hetzner_boot_rescue.test", "server_ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_boot_rescue.test", "active", "false"),
				),
			},
		},
	})
}
