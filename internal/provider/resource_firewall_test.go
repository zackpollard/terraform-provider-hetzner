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
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

func newTestFirewallServer() *httptest.Server {
	mux := http.NewServeMux()

	fw := firewallAPIModel{
		ServerIP:     "10.0.0.1",
		ServerNumber: 12345,
		Status:       "active",
		AllowlistHOS: true,
		FilterIPv6:   false,
		Port:         "main",
		Rules: firewallAPIRules{
			Input: []firewallAPIRule{
				{
					IPVersion: "ipv4",
					Name:      "Allow SSH",
					DstPort:   "22",
					Protocol:  "tcp",
					Action:    "accept",
				},
			},
			Output: []firewallAPIRule{},
		},
	}

	mux.HandleFunc("/firewall/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(firewallAPIResponse{Firewall: fw})
		case http.MethodPost:
			_ = r.ParseForm()
			if s := r.FormValue("status"); s != "" {
				fw.Status = s
			}
			if a := r.FormValue("whitelist_hos"); a == "true" {
				fw.AllowlistHOS = true
			} else if a == "false" {
				fw.AllowlistHOS = false
			}
			if f := r.FormValue("filter_ipv6"); f == "true" {
				fw.FilterIPv6 = true
			} else if f == "false" {
				fw.FilterIPv6 = false
			}
			// Rebuild rules from form
			fw.Rules.Input = nil
			fw.Rules.Output = nil
			for i := 0; i < 10; i++ {
				action := r.FormValue(fmt.Sprintf("rules[input][%d][action]", i))
				if action == "" {
					break
				}
				fw.Rules.Input = append(fw.Rules.Input, firewallAPIRule{
					IPVersion: r.FormValue(fmt.Sprintf("rules[input][%d][ip_version]", i)),
					Name:      r.FormValue(fmt.Sprintf("rules[input][%d][name]", i)),
					DstIP:     r.FormValue(fmt.Sprintf("rules[input][%d][dst_ip]", i)),
					SrcIP:     r.FormValue(fmt.Sprintf("rules[input][%d][src_ip]", i)),
					DstPort:   r.FormValue(fmt.Sprintf("rules[input][%d][dst_port]", i)),
					SrcPort:   r.FormValue(fmt.Sprintf("rules[input][%d][src_port]", i)),
					Protocol:  r.FormValue(fmt.Sprintf("rules[input][%d][protocol]", i)),
					TCPFlags:  r.FormValue(fmt.Sprintf("rules[input][%d][tcp_flags]", i)),
					Action:    action,
				})
			}
			for i := 0; i < 10; i++ {
				action := r.FormValue(fmt.Sprintf("rules[output][%d][action]", i))
				if action == "" {
					break
				}
				fw.Rules.Output = append(fw.Rules.Output, firewallAPIRule{
					IPVersion: r.FormValue(fmt.Sprintf("rules[output][%d][ip_version]", i)),
					Name:      r.FormValue(fmt.Sprintf("rules[output][%d][name]", i)),
					DstIP:     r.FormValue(fmt.Sprintf("rules[output][%d][dst_ip]", i)),
					SrcIP:     r.FormValue(fmt.Sprintf("rules[output][%d][src_ip]", i)),
					DstPort:   r.FormValue(fmt.Sprintf("rules[output][%d][dst_port]", i)),
					SrcPort:   r.FormValue(fmt.Sprintf("rules[output][%d][src_port]", i)),
					Protocol:  r.FormValue(fmt.Sprintf("rules[output][%d][protocol]", i)),
					TCPFlags:  r.FormValue(fmt.Sprintf("rules[output][%d][tcp_flags]", i)),
					Action:    action,
				})
			}
			if fw.Rules.Input == nil {
				fw.Rules.Input = []firewallAPIRule{}
			}
			if fw.Rules.Output == nil {
				fw.Rules.Output = []firewallAPIRule{}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(firewallAPIResponse{Firewall: fw})
		case http.MethodDelete:
			fw.Rules.Input = []firewallAPIRule{}
			fw.Rules.Output = []firewallAPIRule{}
			fw.Status = "disabled"
			w.WriteHeader(http.StatusOK)
		}
	})

	return httptest.NewServer(mux)
}

func TestUnitFirewallResource_Create(t *testing.T) {
	server := newTestFirewallServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_firewall" "test" {
  server_number = "12345"
  status        = "active"
  allowlist_hos = true

  input = [{
    ip_version = "ipv4"
    name       = "Allow SSH"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall.test", "server_number", "12345"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "status", "active"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "allowlist_hos", "true"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "server_ip", "10.0.0.1"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "input.#", "1"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "input.0.action", "accept"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "input.0.dst_port", "22"),
				),
			},
		},
	})
}

func TestUnitFirewallResource_Update(t *testing.T) {
	server := newTestFirewallServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_firewall" "test" {
  server_number = "12345"
  status        = "active"
  allowlist_hos = true

  input = [{
    ip_version = "ipv4"
    name       = "Allow SSH"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }]
}`,
			},
			{
				Config: `
resource "hetzner_firewall" "test" {
  server_number = "12345"
  status        = "active"
  allowlist_hos = false

  input = [{
    ip_version = "ipv4"
    name       = "Allow HTTPS"
    dst_port   = "443"
    protocol   = "tcp"
    action     = "accept"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_firewall.test", "allowlist_hos", "false"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "input.0.name", "Allow HTTPS"),
					resource.TestCheckResourceAttr("hetzner_firewall.test", "input.0.dst_port", "443"),
				),
			},
		},
	})
}

func TestUnitFirewallResource_Import(t *testing.T) {
	server := newTestFirewallServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_firewall" "test" {
  server_number = "12345"
  status        = "active"
  allowlist_hos = true

  input = [{
    ip_version = "ipv4"
    name       = "Allow SSH"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }]
}`,
			},
			{
				ResourceName:                         "hetzner_firewall.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "12345",
				ImportStateVerifyIdentifierAttribute: "server_number",
			},
		},
	})
}

func TestUnitFirewallDataSource(t *testing.T) {
	server := newTestFirewallServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_firewall" "test" {
  server_number = "12345"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_firewall.test", "server_number", "12345"),
					resource.TestCheckResourceAttr("data.hetzner_firewall.test", "status", "active"),
					resource.TestCheckResourceAttr("data.hetzner_firewall.test", "server_ip", "10.0.0.1"),
					resource.TestCheckResourceAttr("data.hetzner_firewall.test", "input.#", "1"),
				),
			},
		},
	})
}
