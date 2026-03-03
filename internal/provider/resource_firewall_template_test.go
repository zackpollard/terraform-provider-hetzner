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

func newTestFirewallTemplateServer() *httptest.Server {
	mux := http.NewServeMux()

	tmpl := firewallTemplateAPIModel{
		ID:           42,
		Name:         "my-template",
		FilterIPv6:   false,
		AllowlistHOS: true,
		IsDefault:    false,
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

	mux.HandleFunc("/firewall/template", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := []firewallTemplateAPIResponse{{FirewallTemplate: tmpl}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodPost:
			_ = r.ParseForm()
			tmpl.Name = r.FormValue("name")
			if r.FormValue("filter_ipv6") == "true" {
				tmpl.FilterIPv6 = true
			} else {
				tmpl.FilterIPv6 = false
			}
			if r.FormValue("whitelist_hos") == "true" {
				tmpl.AllowlistHOS = true
			} else {
				tmpl.AllowlistHOS = false
			}
			if r.FormValue("is_default") == "true" {
				tmpl.IsDefault = true
			} else {
				tmpl.IsDefault = false
			}
			// Rebuild rules from form
			tmpl.Rules.Input = nil
			tmpl.Rules.Output = nil
			for i := 0; i < 10; i++ {
				action := r.FormValue(fmt.Sprintf("rules[input][%d][action]", i))
				if action == "" {
					break
				}
				tmpl.Rules.Input = append(tmpl.Rules.Input, firewallAPIRule{
					IPVersion: r.FormValue(fmt.Sprintf("rules[input][%d][ip_version]", i)),
					Name:      r.FormValue(fmt.Sprintf("rules[input][%d][name]", i)),
					DstPort:   r.FormValue(fmt.Sprintf("rules[input][%d][dst_port]", i)),
					Protocol:  r.FormValue(fmt.Sprintf("rules[input][%d][protocol]", i)),
					Action:    action,
				})
			}
			if tmpl.Rules.Input == nil {
				tmpl.Rules.Input = []firewallAPIRule{}
			}
			if tmpl.Rules.Output == nil {
				tmpl.Rules.Output = []firewallAPIRule{}
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(firewallTemplateAPIResponse{FirewallTemplate: tmpl})
		}
	})

	mux.HandleFunc("/firewall/template/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(firewallTemplateAPIResponse{FirewallTemplate: tmpl})
		case http.MethodPost:
			_ = r.ParseForm()
			tmpl.Name = r.FormValue("name")
			if r.FormValue("filter_ipv6") == "true" {
				tmpl.FilterIPv6 = true
			} else {
				tmpl.FilterIPv6 = false
			}
			if r.FormValue("whitelist_hos") == "true" {
				tmpl.AllowlistHOS = true
			} else {
				tmpl.AllowlistHOS = false
			}
			// Rebuild rules from form
			tmpl.Rules.Input = nil
			tmpl.Rules.Output = nil
			for i := 0; i < 10; i++ {
				action := r.FormValue(fmt.Sprintf("rules[input][%d][action]", i))
				if action == "" {
					break
				}
				tmpl.Rules.Input = append(tmpl.Rules.Input, firewallAPIRule{
					IPVersion: r.FormValue(fmt.Sprintf("rules[input][%d][ip_version]", i)),
					Name:      r.FormValue(fmt.Sprintf("rules[input][%d][name]", i)),
					DstPort:   r.FormValue(fmt.Sprintf("rules[input][%d][dst_port]", i)),
					Protocol:  r.FormValue(fmt.Sprintf("rules[input][%d][protocol]", i)),
					Action:    action,
				})
			}
			if tmpl.Rules.Input == nil {
				tmpl.Rules.Input = []firewallAPIRule{}
			}
			if tmpl.Rules.Output == nil {
				tmpl.Rules.Output = []firewallAPIRule{}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(firewallTemplateAPIResponse{FirewallTemplate: tmpl})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	})

	return httptest.NewServer(mux)
}

func TestUnitFirewallTemplateResource_Create(t *testing.T) {
	server := newTestFirewallTemplateServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_firewall_template" "test" {
  name         = "my-template"
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
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "id", "42"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "name", "my-template"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "allowlist_hos", "true"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "input.#", "1"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "input.0.action", "accept"),
				),
			},
		},
	})
}

func TestUnitFirewallTemplateResource_Update(t *testing.T) {
	server := newTestFirewallTemplateServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_firewall_template" "test" {
  name         = "my-template"
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
resource "hetzner_firewall_template" "test" {
  name         = "updated-template"
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
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "name", "updated-template"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "allowlist_hos", "false"),
					resource.TestCheckResourceAttr("hetzner_firewall_template.test", "input.0.name", "Allow HTTPS"),
				),
			},
		},
	})
}

func TestUnitFirewallTemplateResource_Import(t *testing.T) {
	server := newTestFirewallTemplateServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hetzner_firewall_template" "test" {
  name         = "my-template"
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
				ResourceName:      "hetzner_firewall_template.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "42",
			},
		},
	})
}

func TestUnitFirewallTemplateDataSource(t *testing.T) {
	server := newTestFirewallTemplateServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `
data "hetzner_firewall_template" "test" {
  id = "42"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_firewall_template.test", "name", "my-template"),
					resource.TestCheckResourceAttr("data.hetzner_firewall_template.test", "allowlist_hos", "true"),
					resource.TestCheckResourceAttr("data.hetzner_firewall_template.test", "input.#", "1"),
				),
			},
		},
	})
}

func TestUnitFirewallTemplatesDataSource(t *testing.T) {
	server := newTestFirewallTemplateServer()
	defer server.Close()

	c := client.NewClient("test", "test")
	c.BaseURL = server.URL

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: batch1ProviderFactories(c),
		Steps: []resource.TestStep{
			{
				Config: `data "hetzner_firewall_templates" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_firewall_templates.test", "templates.#", "1"),
					resource.TestCheckResourceAttr("data.hetzner_firewall_templates.test", "templates.0.name", "my-template"),
				),
			},
		},
	})
}
