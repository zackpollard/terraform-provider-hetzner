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

func TestUnitIPCancellationResource_create(t *testing.T) {
	cancelled := false
	cancellationDate := ""

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/ip/1.2.3.4/cancellation":
			_ = r.ParseForm()
			cancellationDate = r.FormValue("cancellation_date")
			cancelled = true
			_ = json.NewEncoder(w).Encode(ipCancellationAPIResponse{
				Cancellation: ipCancellationAPI{
					IP:                       "1.2.3.4",
					ServerNumber:             "321",
					EarliestCancellationDate: "2025-01-01",
					Cancelled:                true,
					CancellationDate:         &cancellationDate,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4/cancellation":
			resp := ipCancellationAPIResponse{
				Cancellation: ipCancellationAPI{
					IP:                       "1.2.3.4",
					ServerNumber:             "321",
					EarliestCancellationDate: "2025-01-01",
					Cancelled:                cancelled,
				},
			}
			if cancelled {
				resp.Cancellation.CancellationDate = &cancellationDate
			}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodDelete && r.URL.Path == "/ip/1.2.3.4/cancellation":
			cancelled = false
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
				Config: `resource "hetzner_ip_cancellation" "test" {
					ip                = "1.2.3.4"
					cancellation_date = "2025-06-01"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ip_cancellation.test", "ip", "1.2.3.4"),
					resource.TestCheckResourceAttr("hetzner_ip_cancellation.test", "cancelled", "true"),
					resource.TestCheckResourceAttr("hetzner_ip_cancellation.test", "cancellation_date", "2025-06-01"),
					resource.TestCheckResourceAttr("hetzner_ip_cancellation.test", "server_number", "321"),
				),
			},
		},
	})
}

func TestUnitIPCancellationResource_import(t *testing.T) {
	cancellationDate := "2025-06-01"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/ip/1.2.3.4/cancellation":
			_ = json.NewEncoder(w).Encode(ipCancellationAPIResponse{
				Cancellation: ipCancellationAPI{
					IP:                       "1.2.3.4",
					ServerNumber:             "321",
					EarliestCancellationDate: "2025-01-01",
					Cancelled:                true,
					CancellationDate:         &cancellationDate,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/ip/1.2.3.4/cancellation":
			_ = json.NewEncoder(w).Encode(ipCancellationAPIResponse{
				Cancellation: ipCancellationAPI{
					IP:                       "1.2.3.4",
					ServerNumber:             "321",
					EarliestCancellationDate: "2025-01-01",
					Cancelled:                true,
					CancellationDate:         &cancellationDate,
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/ip/1.2.3.4/cancellation":
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
				Config: `resource "hetzner_ip_cancellation" "test" {
					ip                = "1.2.3.4"
					cancellation_date = "2025-06-01"
				}`,
			},
			{
				ResourceName:                         "hetzner_ip_cancellation.test",
				ImportState:                          true,
				ImportStateId:                        "1.2.3.4",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "ip",
			},
		},
	})
}
