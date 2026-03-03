// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitVSwitchResource_create(t *testing.T) {
	created := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vswitch":
			created = true
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(vSwitchAPIResponse{
				ID: 123, Name: "test-vswitch", Vlan: 4000, Cancelled: false,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/vswitch/123":
			json.NewEncoder(w).Encode(vSwitchAPIResponse{
				ID: 123, Name: "test-vswitch", Vlan: 4000, Cancelled: false,
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/vswitch/123":
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
				Config: `resource "hetzner_vswitch" "test" {
					name = "test-vswitch"
					vlan = 4000
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "id", "123"),
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "name", "test-vswitch"),
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "vlan", "4000"),
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "cancelled", "false"),
				),
			},
		},
	})
	if !created {
		t.Fatal("expected vSwitch to be created")
	}
}

func TestUnitVSwitchResource_update(t *testing.T) {
	currentName := "test-vswitch"
	currentVlan := 4000
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vswitch":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(vSwitchAPIResponse{
				ID: 123, Name: "test-vswitch", Vlan: 4000, Cancelled: false,
			})
		case r.Method == http.MethodPost && r.URL.Path == "/vswitch/123":
			r.ParseForm()
			if name := r.FormValue("name"); name != "" {
				currentName = name
			}
			if vlan := r.FormValue("vlan"); vlan != "" {
				fmt.Sscanf(vlan, "%d", &currentVlan)
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/vswitch/123":
			json.NewEncoder(w).Encode(vSwitchAPIResponse{
				ID: 123, Name: currentName, Vlan: currentVlan, Cancelled: false,
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/vswitch/123":
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
				Config: `resource "hetzner_vswitch" "test" {
					name = "test-vswitch"
					vlan = 4000
				}`,
				Check: resource.TestCheckResourceAttr("hetzner_vswitch.test", "name", "test-vswitch"),
			},
			{
				Config: `resource "hetzner_vswitch" "test" {
					name = "updated-vswitch"
					vlan = 4001
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "name", "updated-vswitch"),
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "vlan", "4001"),
				),
			},
		},
	})
}

func TestUnitVSwitchResource_import(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/vswitch":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(vSwitchAPIResponse{
				ID: 123, Name: "test-vswitch", Vlan: 4000, Cancelled: false,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/vswitch/123":
			json.NewEncoder(w).Encode(vSwitchAPIResponse{
				ID: 123, Name: "test-vswitch", Vlan: 4000, Cancelled: false,
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/vswitch/123":
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
				Config: `resource "hetzner_vswitch" "test" {
					name = "test-vswitch"
					vlan = 4000
				}`,
			},
			{
				ResourceName:      "hetzner_vswitch.test",
				ImportState:       true,
				ImportStateId:     "123",
				ImportStateVerify: true,
			},
		},
	})
}

func TestUnitVSwitchDataSource_read(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/vswitch/123" {
			json.NewEncoder(w).Encode(vSwitchAPIResponse{
				ID: 123, Name: "test-vswitch", Vlan: 4000, Cancelled: false,
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
				Config: `data "hetzner_vswitch" "test" {
					id = 123
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_vswitch.test", "name", "test-vswitch"),
					resource.TestCheckResourceAttr("data.hetzner_vswitch.test", "vlan", "4000"),
					resource.TestCheckResourceAttr("data.hetzner_vswitch.test", "cancelled", "false"),
				),
			},
		},
	})
}

func TestUnitVSwitchesDataSource_list(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/vswitch" {
			json.NewEncoder(w).Encode([]vSwitchAPIResponse{
				{ID: 1, Name: "vs1", Vlan: 4000, Cancelled: false},
				{ID: 2, Name: "vs2", Vlan: 4001, Cancelled: false},
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
				Config: `data "hetzner_vswitches" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_vswitches.test", "vswitches.#", "2"),
					resource.TestCheckResourceAttr("data.hetzner_vswitches.test", "vswitches.0.name", "vs1"),
					resource.TestCheckResourceAttr("data.hetzner_vswitches.test", "vswitches.1.name", "vs2"),
				),
			},
		},
	})
}
