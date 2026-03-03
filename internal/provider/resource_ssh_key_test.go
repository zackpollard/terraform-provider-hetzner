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

func newTestSSHKeyServer() *httptest.Server {
	mux := http.NewServeMux()

	sshKey := sshKeyAPIModel{
		Name:        "my-key",
		Fingerprint: "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff",
		Type:        "ED25519",
		Size:        256,
		Data:        "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest test@example.com",
		CreatedAt:   "2024-01-01",
	}

	mux.HandleFunc("/key", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := []sshKeyAPIResponse{{Key: sshKey}}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case http.MethodPost:
			r.ParseForm()
			sshKey.Name = r.FormValue("name")
			if data := r.FormValue("data"); data != "" {
				sshKey.Data = data
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sshKeyAPIResponse{Key: sshKey})
		}
	})

	mux.HandleFunc("/key/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sshKeyAPIResponse{Key: sshKey})
		case http.MethodPost:
			r.ParseForm()
			sshKey.Name = r.FormValue("name")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sshKeyAPIResponse{Key: sshKey})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	})

	return httptest.NewServer(mux)
}

func TestAccSSHKeyResource_Create(t *testing.T) {
	server := newTestSSHKeyServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_ssh_key" "test" {
  name = "my-key"
  data = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest test@example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "name", "my-key"),
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "fingerprint", "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"),
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "type", "ED25519"),
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "size", "256"),
				),
			},
		},
	})
}

func TestAccSSHKeyResource_Update(t *testing.T) {
	server := newTestSSHKeyServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_ssh_key" "test" {
  name = "my-key"
  data = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest test@example.com"
}`,
			},
			{
				Config: `
resource "hetzner_ssh_key" "test" {
  name = "updated-key"
  data = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest test@example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_ssh_key.test", "name", "updated-key"),
				),
			},
		},
	})
}

func TestAccSSHKeyResource_Import(t *testing.T) {
	server := newTestSSHKeyServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_ssh_key" "test" {
  name = "my-key"
  data = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest test@example.com"
}`,
			},
			{
				ResourceName:                         "hetzner_ssh_key.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff",
				ImportStateVerifyIdentifierAttribute: "fingerprint",
			},
		},
	})
}

func TestAccSSHKeyDataSource(t *testing.T) {
	server := newTestSSHKeyServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_ssh_key" "test" {
  fingerprint = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_ssh_key.test", "name", "my-key"),
					resource.TestCheckResourceAttr("data.hetzner_ssh_key.test", "type", "ED25519"),
					resource.TestCheckResourceAttr("data.hetzner_ssh_key.test", "size", "256"),
				),
			},
		},
	})
}

func TestAccSSHKeysDataSource(t *testing.T) {
	server := newTestSSHKeyServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_ssh_keys" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_ssh_keys.test", "ssh_keys.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_ssh_keys.test", "ssh_keys.0.name", "my-key"),
					resource.TestCheckResourceAttr("data.hetzner_ssh_keys.test", "ssh_keys.0.fingerprint",
						"00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"),
				),
			},
		},
	})
}
