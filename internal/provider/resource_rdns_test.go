// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

func newTestRDNSServer() *httptest.Server {
	mux := http.NewServeMux()

	rdns := rdnsAPIModel{
		IP:  "1.2.3.4",
		PTR: "server.example.com",
	}

	mux.HandleFunc("/rdns/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rdnsAPIResponse{Rdns: rdns})
		case http.MethodPut:
			r.ParseForm()
			rdns.PTR = r.FormValue("ptr")
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rdnsAPIResponse{Rdns: rdns})
		case http.MethodPost:
			r.ParseForm()
			rdns.PTR = r.FormValue("ptr")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rdnsAPIResponse{Rdns: rdns})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	})

	return httptest.NewServer(mux)
}

func TestAccRDNSResource_Create(t *testing.T) {
	server := newTestRDNSServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_rdns" "test" {
  ip  = "1.2.3.4"
  ptr = "server.example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_rdns.test", "ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("hetzner_rdns.test", "ptr", "server.example.com"),
				),
			},
		},
	})
}

func TestAccRDNSResource_Update(t *testing.T) {
	server := newTestRDNSServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_rdns" "test" {
  ip  = "1.2.3.4"
  ptr = "server.example.com"
}`,
			},
			{
				Config: `
resource "hetzner_rdns" "test" {
  ip  = "1.2.3.4"
  ptr = "updated.example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_rdns.test", "ptr", "updated.example.com"),
				),
			},
		},
	})
}

func TestAccRDNSResource_Import(t *testing.T) {
	server := newTestRDNSServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_rdns" "test" {
  ip  = "1.2.3.4"
  ptr = "server.example.com"
}`,
			},
			{
				ResourceName:                         "hetzner_rdns.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "1.2.3.4",
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}

func TestAccRDNSDataSource(t *testing.T) {
	server := newTestRDNSServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_rdns" "test" {
  ip = "1.2.3.4"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_rdns.test", "ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.hetzner_rdns.test", "ptr", "server.example.com"),
				),
			},
		},
	})
}
