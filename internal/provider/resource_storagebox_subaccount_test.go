// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func newTestSubaccountServer() *httptest.Server {
	subaccounts := []subaccountAPIData{}

	mux := http.NewServeMux()

	mux.HandleFunc("/storagebox/1001/subaccount", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			resp := make([]subaccountAPIResponse, len(subaccounts))
			for i, sa := range subaccounts {
				resp[i] = subaccountAPIResponse{Subaccount: sa}
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodPost:
			r.ParseForm()
			sa := subaccountAPIData{
				Username:             r.FormValue("username"),
				Homedirectory:        r.FormValue("homedirectory"),
				Samba:                r.FormValue("samba") == "true",
				Webdav:               r.FormValue("webdav") == "true",
				SSH:                  r.FormValue("ssh") == "true",
				ExternalReachability: r.FormValue("external_reachability") == "true",
				Readonly:             r.FormValue("readonly") == "true",
				Createdir:            r.FormValue("createdir") == "true",
			}
			if comment := r.FormValue("comment"); comment != "" {
				sa.Comment = &comment
			}
			subaccounts = append(subaccounts, sa)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(subaccountAPIResponse{Subaccount: sa})
		}
	})

	mux.HandleFunc("/storagebox/1001/subaccount/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		username := parts[len(parts)-1]
		// Handle password reset path
		if username == "password" && len(parts) >= 5 {
			username = parts[len(parts)-2]
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPut:
			r.ParseForm()
			for i, sa := range subaccounts {
				if sa.Username == username {
					if v := r.FormValue("homedirectory"); v != "" {
						subaccounts[i].Homedirectory = v
					}
					if v := r.FormValue("samba"); v != "" {
						subaccounts[i].Samba = v == "true"
					}
					if v := r.FormValue("ssh"); v != "" {
						subaccounts[i].SSH = v == "true"
					}
					if v := r.FormValue("webdav"); v != "" {
						subaccounts[i].Webdav = v == "true"
					}
					if v := r.FormValue("external_reachability"); v != "" {
						subaccounts[i].ExternalReachability = v == "true"
					}
					if v := r.FormValue("readonly"); v != "" {
						subaccounts[i].Readonly = v == "true"
					}
					if v := r.FormValue("createdir"); v != "" {
						subaccounts[i].Createdir = v == "true"
					}
					json.NewEncoder(w).Encode(subaccountAPIResponse{Subaccount: subaccounts[i]})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodDelete:
			for i, sa := range subaccounts {
				if sa.Username == username {
					subaccounts = append(subaccounts[:i], subaccounts[i+1:]...)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return httptest.NewServer(mux)
}

func TestAccStorageboxSubaccountResource_Create(t *testing.T) {
	ts := newTestSubaccountServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox_subaccount" "test" {
  storagebox_id = 1001
  username      = "sub1"
  homedirectory = "/home/sub1"
  ssh           = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "storagebox_id", "1001"),
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "username", "sub1"),
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "homedirectory", "/home/sub1"),
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "ssh", "true"),
				),
			},
		},
	})
}

func TestAccStorageboxSubaccountResource_Update(t *testing.T) {
	ts := newTestSubaccountServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox_subaccount" "test" {
  storagebox_id = 1001
  username      = "sub1"
  homedirectory = "/home/sub1"
  ssh           = true
}`,
			},
			{
				Config: `
resource "hetzner_storagebox_subaccount" "test" {
  storagebox_id = 1001
  username      = "sub1"
  homedirectory = "/home/sub1"
  ssh           = true
  webdav        = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "webdav", "true"),
				),
			},
		},
	})
}

func TestAccStorageboxSubaccountDataSource(t *testing.T) {
	ts := newTestSubaccountServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_storagebox_subaccount" "test" {
  storagebox_id = 1001
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_storagebox_subaccount.test", "subaccounts.#", "0"),
				),
			},
		},
	})
}
