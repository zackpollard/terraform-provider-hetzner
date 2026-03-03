// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func newTestSnapshotplanServer() *httptest.Server {
	plan := snapshotplanAPIData{
		Status:     "disabled",
		Minute:     0,
		Hour:       0,
		DayOfWeek:  0,
		DayOfMonth: 1,
		Month:      1,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/storagebox/1001/snapshotplan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(snapshotplanAPIResponse{Snapshotplan: plan})
		case http.MethodPost:
			r.ParseForm()
			if v := r.FormValue("status"); v != "" {
				plan.Status = v
			}
			if v := r.FormValue("minute"); v != "" {
				plan.Minute, _ = strconv.Atoi(v)
			}
			if v := r.FormValue("hour"); v != "" {
				plan.Hour, _ = strconv.Atoi(v)
			}
			if v := r.FormValue("day_of_week"); v != "" {
				plan.DayOfWeek, _ = strconv.Atoi(v)
			}
			if v := r.FormValue("day_of_month"); v != "" {
				plan.DayOfMonth, _ = strconv.Atoi(v)
			}
			if v := r.FormValue("month"); v != "" {
				plan.Month, _ = strconv.Atoi(v)
			}
			json.NewEncoder(w).Encode(snapshotplanAPIResponse{Snapshotplan: plan})
		}
	})

	return httptest.NewServer(mux)
}

func TestAccStorageboxSnapshotplanResource_Create(t *testing.T) {
	ts := newTestSnapshotplanServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox_snapshotplan" "test" {
  storagebox_id = 1001
  status        = "enabled"
  hour          = 3
  minute        = 30
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "storagebox_id", "1001"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "status", "enabled"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "hour", "3"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "minute", "30"),
				),
			},
		},
	})
}

func TestAccStorageboxSnapshotplanResource_Update(t *testing.T) {
	ts := newTestSnapshotplanServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox_snapshotplan" "test" {
  storagebox_id = 1001
  status        = "enabled"
  hour          = 3
  minute        = 30
}`,
			},
			{
				Config: `
resource "hetzner_storagebox_snapshotplan" "test" {
  storagebox_id = 1001
  status        = "enabled"
  hour          = 6
  minute        = 0
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "hour", "6"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "minute", "0"),
				),
			},
		},
	})
}

func TestAccStorageboxSnapshotplanResource_Import(t *testing.T) {
	ts := newTestSnapshotplanServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox_snapshotplan" "test" {
  storagebox_id = 1001
  status        = "enabled"
  hour          = 3
  minute        = 30
}`,
			},
			{
				ResourceName:      "hetzner_storagebox_snapshotplan.test",
				ImportState:       true,
				ImportStateId:     "1001",
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccStorageboxSnapshotplanDataSource(t *testing.T) {
	ts := newTestSnapshotplanServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_storagebox_snapshotplan" "test" {
  storagebox_id = 1001
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_storagebox_snapshotplan.test", "storagebox_id", "1001"),
					resource.TestCheckResourceAttr("data.hetzner_storagebox_snapshotplan.test", "status", "disabled"),
				),
			},
		},
	})
}
