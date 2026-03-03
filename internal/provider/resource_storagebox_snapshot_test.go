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

func newTestSnapshotServer() *httptest.Server {
	snapshots := []snapshotAPIData{}

	mux := http.NewServeMux()

	// Handle sub-paths first (longer prefix wins in Go 1.22+ with patterns)
	mux.HandleFunc("/storagebox/1001/snapshot/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		trimmed := strings.TrimPrefix(path, "/storagebox/1001/snapshot/")
		parts := strings.Split(trimmed, "/")
		snapName := parts[0]

		if len(parts) == 2 && parts[1] == "comment" {
			r.ParseForm()
			for i, s := range snapshots {
				if s.Name == snapName {
					comment := r.FormValue("comment")
					snapshots[i].Comment = &comment
					json.NewEncoder(w).Encode(map[string]interface{}{"snapshot": snapshots[i]})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			for i, s := range snapshots {
				if s.Name == snapName {
					snapshots = append(snapshots[:i], snapshots[i+1:]...)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}
	})

	mux.HandleFunc("/storagebox/1001/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			resp := make([]struct {
				Snapshot snapshotAPIData `json:"snapshot"`
			}, len(snapshots))
			for i, s := range snapshots {
				resp[i].Snapshot = s
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodPost:
			r.ParseForm()
			snap := snapshotAPIData{
				Name:      "snap-2024-01-01",
				Timestamp: "2024-01-01T03:00:00+01:00",
				Size:      50,
			}
			if comment := r.FormValue("comment"); comment != "" {
				snap.Comment = &comment
			}
			snapshots = append(snapshots, snap)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(snapshotCreateAPIResponse{Snapshot: snap})
		}
	})

	return httptest.NewServer(mux)
}

func TestAccStorageboxSnapshotResource_Create(t *testing.T) {
	ts := newTestSnapshotServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_storagebox_snapshot" "test" {
  storagebox_id = 1001
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshot.test", "storagebox_id", "1001"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshot.test", "name", "snap-2024-01-01"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshot.test", "size", "50"),
				),
			},
		},
	})
}

func TestAccStorageboxSnapshotDataSource(t *testing.T) {
	ts := newTestSnapshotServer()
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch3ProviderFactories(ts),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_storagebox_snapshot" "test" {
  storagebox_id = 1001
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_storagebox_snapshot.test", "snapshots.#", "0"),
				),
			},
		},
	})
}
