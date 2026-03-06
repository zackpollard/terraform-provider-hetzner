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

func TestUnitRDNSListDataSource_readAll(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rdns":
			_ = json.NewEncoder(w).Encode([]rdnsListAPIEntry{
				{Rdns: rdnsAPIModel{IP: "1.2.3.4", PTR: "host1.example.com"}},
				{Rdns: rdnsAPIModel{IP: "5.6.7.8", PTR: "host2.example.com"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_rdns_list" "test" {
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_rdns_list.test", "entries.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_rdns_list.test", "entries.0.ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_rdns_list.test", "entries.0.ptr", "host1.example.com"),
					resource.TestCheckResourceAttr("data.hetzner_rdns_list.test", "entries.1.ip", "5.6.7.8"),
					resource.TestCheckResourceAttr("data.hetzner_rdns_list.test", "entries.1.ptr", "host2.example.com"),
				),
			},
		},
	})
}

func TestUnitRDNSListDataSource_filterByServerIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rdns":
			if r.URL.Query().Get("server_ip") == "10.0.0.1" {
				_ = json.NewEncoder(w).Encode([]rdnsListAPIEntry{
					{Rdns: rdnsAPIModel{IP: "1.2.3.4", PTR: "host1.example.com"}},
				})
			} else {
				_ = json.NewEncoder(w).Encode([]rdnsListAPIEntry{
					{Rdns: rdnsAPIModel{IP: "1.2.3.4", PTR: "host1.example.com"}},
					{Rdns: rdnsAPIModel{IP: "5.6.7.8", PTR: "host2.example.com"}},
				})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_rdns_list" "test" {
					server_ip = "10.0.0.1"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_rdns_list.test", "entries.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_rdns_list.test", "entries.0.ip", "1.2.3.4"),
				),
			},
		},
	})
}
