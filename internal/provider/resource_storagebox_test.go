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

func newTestStorageboxServer() *httptest.Server {
	sb := storageboxAPIData{
		StorageboxID:         1001,
		StorageboxName:       "my-storage",
		DiskQuota:            1000,
		DiskUsage:            250,
		Status:               "enabled",
		PaidUntil:            "2025-12-31",
		Locked:               false,
		Webdav:               false,
		Samba:                false,
		SSH:                  true,
		ExternalReachability: true,
		ZFS:                  false,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/storagebox/1001", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(storageboxAPIResponse{Storagebox: sb})
		case http.MethodPost:
			r.ParseForm()
			if name := r.FormValue("storagebox_name"); name != "" {
				sb.StorageboxName = name
			}
			if v := r.FormValue("webdav"); v != "" {
				sb.Webdav = v == "true"
			}
			if v := r.FormValue("samba"); v != "" {
				sb.Samba = v == "true"
			}
			if v := r.FormValue("ssh"); v != "" {
				sb.SSH = v == "true"
			}
			if v := r.FormValue("external_reachability"); v != "" {
				sb.ExternalReachability = v == "true"
			}
			if v := r.FormValue("zfs"); v != "" {
				sb.ZFS = v == "true"
			}
			json.NewEncoder(w).Encode(storageboxAPIResponse{Storagebox: sb})
		}
	})

	mux.HandleFunc("/storagebox", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]storageboxListAPIResponse{
			{Storagebox: storageboxListAPIData{
				StorageboxID:   sb.StorageboxID,
				StorageboxName: sb.StorageboxName,
				DiskQuota:      sb.DiskQuota,
				DiskUsage:      sb.DiskUsage,
				Status:         sb.Status,
				PaidUntil:      sb.PaidUntil,
				Locked:         sb.Locked,
			}},
		})
	})

	return httptest.NewServer(mux)
}

func TestAccStorageboxResource_Create(t *testing.T) {
	ts := newTestStorageboxServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox" "test" {
  storagebox_id        = 1001
  storagebox_name      = "my-storage"
  ssh                  = true
  external_reachability = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "storagebox_id", "1001"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "storagebox_name", "my-storage"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "ssh", "true"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "external_reachability", "true"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "disk_quota", "1000"),
				),
			},
		},
	})
}

func TestAccStorageboxResource_Update(t *testing.T) {
	ts := newTestStorageboxServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox" "test" {
  storagebox_id   = 1001
  storagebox_name = "my-storage"
  ssh             = true
  external_reachability = true
}`,
			},
			{
				Config: `
resource "hetzner_storagebox" "test" {
  storagebox_id   = 1001
  storagebox_name = "updated-name"
  ssh             = true
  webdav          = true
  external_reachability = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "storagebox_name", "updated-name"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "webdav", "true"),
				),
			},
		},
	})
}

func TestAccStorageboxResource_Import(t *testing.T) {
	ts := newTestStorageboxServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox" "test" {
  storagebox_id   = 1001
  storagebox_name = "my-storage"
  ssh             = true
  external_reachability = true
}`,
			},
			{
				ResourceName:      "hetzner_storagebox.test",
				ImportState:       true,
				ImportStateId:     "1001",
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccStorageboxDataSource(t *testing.T) {
	ts := newTestStorageboxServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_storagebox" "test" {
  storagebox_id = 1001
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_storagebox.test", "storagebox_name", "my-storage"),
					resource.TestCheckResourceAttr("data.hetzner_storagebox.test", "disk_quota", "1000"),
					resource.TestCheckResourceAttr("data.hetzner_storagebox.test", "ssh", "true"),
				),
			},
		},
	})
}

func TestAccStorageboxesDataSource(t *testing.T) {
	ts := newTestStorageboxServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_storageboxes" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_storageboxes.test", "storageboxes.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_storageboxes.test", "storageboxes.0.storagebox_id", "1001"),
				),
			},
		},
	})
}
