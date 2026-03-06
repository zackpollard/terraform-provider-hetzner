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

func TestUnitSubnetCancellationResource_create(t *testing.T) {
	cancelled := false
	cancellationDate := ""

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/subnet/10.0.0.0/cancellation":
			_ = r.ParseForm()
			cancellationDate = r.FormValue("cancellation_date")
			cancelled = true
			_ = json.NewEncoder(w).Encode(subnetCancellationAPIResponse{
				Cancellation: subnetCancellationAPI{
					IP:                       "10.0.0.0",
					ServerNumber:             "321",
					EarliestCancellationDate: "2025-01-01",
					Cancelled:                true,
					CancellationDate:         &cancellationDate,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/subnet/10.0.0.0/cancellation":
			resp := subnetCancellationAPIResponse{
				Cancellation: subnetCancellationAPI{
					IP:                       "10.0.0.0",
					ServerNumber:             "321",
					EarliestCancellationDate: "2025-01-01",
					Cancelled:                cancelled,
				},
			}
			if cancelled {
				resp.Cancellation.CancellationDate = &cancellationDate
			}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodDelete && r.URL.Path == "/subnet/10.0.0.0/cancellation":
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
				Config: `resource "hetzner_subnet_cancellation" "test" {
					ip                = "10.0.0.0"
					cancellation_date = "2025-06-01"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_subnet_cancellation.test", "ip", "10.0.0.0"),
					resource.TestCheckResourceAttr("hetzner_subnet_cancellation.test", "cancelled", "true"),
					resource.TestCheckResourceAttr("hetzner_subnet_cancellation.test", "cancellation_date", "2025-06-01"),
					resource.TestCheckResourceAttr("hetzner_subnet_cancellation.test", "server_number", "321"),
				),
			},
		},
	})
}
