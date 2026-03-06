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

func TestUnitSubnetMACResource_create(t *testing.T) {
	macValue := "00:50:56:00:D7:B2"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/subnet/10.0.0.0/mac":
			_ = r.ParseForm()
			macValue = r.FormValue("mac")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(subnetMACAPIResponse{
				MAC: subnetMACAPIData{IP: "10.0.0.0", Mask: "24", MAC: macValue},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/subnet/10.0.0.0/mac":
			_ = json.NewEncoder(w).Encode(subnetMACAPIResponse{
				MAC: subnetMACAPIData{IP: "10.0.0.0", Mask: "24", MAC: macValue},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/subnet/10.0.0.0/mac":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_subnet_mac" "test" {
					ip  = "10.0.0.0"
					mac = "00:50:56:00:D7:B2"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_subnet_mac.test", "ip", "10.0.0.0"),
					resource.TestCheckResourceAttr("hetzner_subnet_mac.test", "mac", "00:50:56:00:D7:B2"),
				),
			},
		},
	})
}

func TestUnitSubnetMACResource_import(t *testing.T) {
	macValue := "00:50:56:00:D7:B2"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/subnet/10.0.0.0/mac":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(subnetMACAPIResponse{
				MAC: subnetMACAPIData{IP: "10.0.0.0", Mask: "24", MAC: macValue},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/subnet/10.0.0.0/mac":
			_ = json.NewEncoder(w).Encode(subnetMACAPIResponse{
				MAC: subnetMACAPIData{IP: "10.0.0.0", Mask: "24", MAC: macValue},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/subnet/10.0.0.0/mac":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch2ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `resource "hetzner_subnet_mac" "test" {
					ip  = "10.0.0.0"
					mac = "00:50:56:00:D7:B2"
				}`,
			},
			{
				ResourceName:                         "hetzner_subnet_mac.test",
				ImportState:                          true,
				ImportStateId:                        "10.0.0.0",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}
