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

func TestUnitIPMACResource_create(t *testing.T) {
	macValue := "00:50:56:00:D7:A1"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/ip/1.2.3.4/mac":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(ipMACAPIResponse{
				MAC: ipMACAPIData{IP: "1.2.3.4", MAC: macValue},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4/mac":
			_ = json.NewEncoder(w).Encode(ipMACAPIResponse{
				MAC: ipMACAPIData{IP: "1.2.3.4", MAC: macValue},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/ip/1.2.3.4/mac":
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
				Config: `resource "hetzner_ip_mac" "test" {
					ip = "1.2.3.4"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ip_mac.test", "ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("hetzner_ip_mac.test", "mac", macValue),
				),
			},
		},
	})
}

func TestUnitIPMACResource_import(t *testing.T) {
	macValue := "00:50:56:00:D7:A1"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/ip/1.2.3.4/mac":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(ipMACAPIResponse{
				MAC: ipMACAPIData{IP: "1.2.3.4", MAC: macValue},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4/mac":
			_ = json.NewEncoder(w).Encode(ipMACAPIResponse{
				MAC: ipMACAPIData{IP: "1.2.3.4", MAC: macValue},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/ip/1.2.3.4/mac":
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
				Config: `resource "hetzner_ip_mac" "test" {
					ip = "1.2.3.4"
				}`,
			},
			{
				ResourceName:                         "hetzner_ip_mac.test",
				ImportState:                          true,
				ImportStateId:                        "1.2.3.4",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}
