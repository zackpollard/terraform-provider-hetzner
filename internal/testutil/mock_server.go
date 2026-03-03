// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/zack/terraform-provider-hetzner/internal/client"
)

// MockRoute defines a mock response for a specific method+path combination.
type MockRoute struct {
	Method     string
	Path       string
	StatusCode int
	Response   interface{}
}

// MockServer wraps httptest.Server to simulate the Hetzner Robot API.
type MockServer struct {
	Server *httptest.Server

	mu     sync.RWMutex
	routes map[string]MockRoute
}

// routeKey returns a map key for a method+path combination.
func routeKey(method, path string) string {
	return method + " " + path
}

// NewMockServer creates a new mock Hetzner Robot API server.
// Register routes with RegisterRoute before making requests.
func NewMockServer() *MockServer {
	ms := &MockServer{
		routes: make(map[string]MockRoute),
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handler))
	return ms
}

// RegisterRoute adds or replaces a mock route.
func (ms *MockServer) RegisterRoute(route MockRoute) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.routes[routeKey(route.Method, route.Path)] = route
}

// RegisterRoutes adds multiple mock routes at once.
func (ms *MockServer) RegisterRoutes(routes []MockRoute) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, r := range routes {
		ms.routes[routeKey(r.Method, r.Path)] = r
	}
}

// ClearRoutes removes all registered routes.
func (ms *MockServer) ClearRoutes() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.routes = make(map[string]MockRoute)
}

// Close shuts down the mock server.
func (ms *MockServer) Close() {
	ms.Server.Close()
}

// Client returns a client.Client configured to talk to the mock server.
func (ms *MockServer) Client() *client.Client {
	return &client.Client{
		BaseURL:    ms.Server.URL,
		Username:   "test-user",
		Password:   "test-password",
		HTTPClient: ms.Server.Client(),
	}
}

// handler dispatches incoming requests to registered routes.
func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	ms.mu.RLock()
	route, ok := ms.routes[routeKey(r.Method, r.URL.Path)]
	ms.mu.RUnlock()

	if !ok {
		writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "no mock route registered for "+r.Method+" "+r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(route.StatusCode)

	if route.Response != nil {
		_ = json.NewEncoder(w).Encode(route.Response)
	}
}

// writeErrorResponse writes a Hetzner-style API error response.
func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"status":  statusCode,
			"code":    code,
			"message": message,
		},
	})
}

// --- Convenience helpers for common Hetzner API response shapes ---

// SSHKeyResponse returns a Hetzner API SSH key response envelope.
func SSHKeyResponse(name, fingerprint, keyType string, size int, data string) map[string]interface{} {
	return map[string]interface{}{
		"key": map[string]interface{}{
			"name":        name,
			"fingerprint": fingerprint,
			"type":        keyType,
			"size":        size,
			"data":        data,
			"created_at":  "2025-01-01 00:00:00",
		},
	}
}

// SSHKeyListResponse returns a list of SSH key response envelopes.
func SSHKeyListResponse(keys ...map[string]interface{}) []map[string]interface{} {
	return keys
}

// FirewallResponse returns a Hetzner API firewall response envelope.
func FirewallResponse(serverIP string, serverNumber int, status string, inputRules, outputRules []map[string]interface{}) map[string]interface{} {
	if inputRules == nil {
		inputRules = []map[string]interface{}{}
	}
	if outputRules == nil {
		outputRules = []map[string]interface{}{}
	}
	return map[string]interface{}{
		"firewall": map[string]interface{}{
			"server_ip":     serverIP,
			"server_number": serverNumber,
			"status":        status,
			"allowlist_hos": true,
			"filter_ipv6":   false,
			"port":          "main",
			"rules": map[string]interface{}{
				"input":  inputRules,
				"output": outputRules,
			},
		},
	}
}

// ServerResponse returns a Hetzner API server response envelope.
func ServerResponse(serverNumber int, serverName, serverIP, dc, status string) map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"server_ip":       serverIP,
			"server_ipv6_net": "2a01:4f8:111:4221::",
			"server_number":   serverNumber,
			"server_name":     serverName,
			"product":         "DS 3000",
			"dc":              dc,
			"traffic":         "5 TB",
			"status":          status,
			"cancelled":       false,
			"paid_until":      "2026-12-31",
			"ip":              []string{serverIP},
			"subnet":          []interface{}{},
		},
	}
}

// RDNSResponse returns a Hetzner API rDNS response envelope.
func RDNSResponse(ip, ptr string) map[string]interface{} {
	return map[string]interface{}{
		"rdns": map[string]interface{}{
			"ip":  ip,
			"ptr": ptr,
		},
	}
}

// ErrorResponse returns a Hetzner API error response body.
func ErrorResponse(statusCode int, code, message string) map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"status":  statusCode,
			"code":    code,
			"message": message,
		},
	}
}

// VSwitchResponse returns a Hetzner API vSwitch response envelope.
func VSwitchResponse(id int, name string, vlan int, cancelled bool) map[string]interface{} {
	return map[string]interface{}{
		"id":        id,
		"name":      name,
		"vlan":      vlan,
		"cancelled": cancelled,
	}
}

// StorageBoxResponse returns a Hetzner API storage box response envelope.
func StorageBoxResponse(id int, name string, quota, usage int, webdav, samba, ssh, externalReachability, zfs bool) map[string]interface{} {
	return map[string]interface{}{
		"storagebox": map[string]interface{}{
			"storagebox_id":         id,
			"storagebox_name":       name,
			"disk_quota":            quota,
			"disk_usage":            usage,
			"status":                "ready",
			"paid_until":            "2026-12-31",
			"locked":                false,
			"server":                nil,
			"webdav":                webdav,
			"samba":                 samba,
			"ssh":                   ssh,
			"external_reachability": externalReachability,
			"zfs":                   zfs,
		},
	}
}

// FailoverResponse returns a Hetzner API failover response envelope.
func FailoverResponse(ip, netmask, serverIP string, serverNumber int, activeServerIP string) map[string]interface{} {
	return map[string]interface{}{
		"failover": map[string]interface{}{
			"ip":               ip,
			"netmask":          netmask,
			"server_ip":        serverIP,
			"server_ipv6_net":  "2a01:4f8:111:4221::",
			"server_number":    serverNumber,
			"active_server_ip": activeServerIP,
		},
	}
}

// BootRescueResponse returns a Hetzner API boot rescue response envelope.
func BootRescueResponse(serverNumber int, serverIP string, active bool, os, password string) map[string]interface{} {
	resp := map[string]interface{}{
		"rescue": map[string]interface{}{
			"server_ip":       serverIP,
			"server_ipv6_net": "2a01:4f8:111:4221::",
			"server_number":   serverNumber,
			"os":              os,
			"arch":            64,
			"active":          active,
			"password":        password,
			"authorized_key":  []interface{}{},
			"host_key":        []interface{}{},
		},
	}
	return resp
}

// FirewallTemplateResponse returns a Hetzner API firewall template response envelope.
func FirewallTemplateResponse(id int, name string, filterIPv6, allowlistHOS, isDefault bool, inputRules, outputRules []map[string]interface{}) map[string]interface{} {
	if inputRules == nil {
		inputRules = []map[string]interface{}{}
	}
	if outputRules == nil {
		outputRules = []map[string]interface{}{}
	}
	return map[string]interface{}{
		"firewall_template": map[string]interface{}{
			"id":            id,
			"name":          name,
			"filter_ipv6":   filterIPv6,
			"allowlist_hos": allowlistHOS,
			"is_default":    isDefault,
			"rules": map[string]interface{}{
				"input":  inputRules,
				"output": outputRules,
			},
		},
	}
}
