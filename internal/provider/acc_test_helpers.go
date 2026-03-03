// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/zack/terraform-provider-hetzner/internal/client"
)

// --- Test API client ---

// testAccNewClient creates an API client using acceptance test credentials.
func testAccNewClient(t *testing.T) *client.Client {
	t.Helper()
	username := os.Getenv("HETZNER_ROBOT_USERNAME")
	password := os.Getenv("HETZNER_ROBOT_PASSWORD")
	if username == "" || password == "" {
		t.Fatal("HETZNER_ROBOT_USERNAME and HETZNER_ROBOT_PASSWORD must be set")
	}
	return client.NewClient(username, password)
}

// --- Server helpers ---

// testAccGetOrCreateServer returns a server number for testing. It uses
// HETZNER_TEST_SERVER_NUMBER if set, otherwise if HETZNER_TEST_SERVER_CREATE=1
// it orders the cheapest hourly server from the server market and registers
// cleanup to cancel it when the test finishes. If neither is set, the test is skipped.
func testAccGetOrCreateServer(t *testing.T) string {
	t.Helper()

	if v := os.Getenv("HETZNER_TEST_SERVER_NUMBER"); v != "" {
		return v
	}

	if os.Getenv("HETZNER_TEST_SERVER_CREATE") != "1" {
		t.Skip("HETZNER_TEST_SERVER_NUMBER or HETZNER_TEST_SERVER_CREATE=1 required; skipping server-dependent test")
	}

	c := testAccNewClient(t)

	// Find cheapest hourly market server.
	body, err := c.Get("/order/server_market/product")
	if err != nil {
		t.Fatalf("Error listing server market products: %s", err)
	}

	var products []struct {
		Product struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Price string `json:"price"`
		} `json:"product"`
	}
	if err := json.Unmarshal(body, &products); err != nil {
		t.Fatalf("Error parsing server market products: %s", err)
	}
	if len(products) == 0 {
		t.Fatal("No server market products available")
	}

	// Order the first (cheapest) product.
	productID := products[0].Product.ID
	t.Logf("Ordering server market product %d (%s)", productID, products[0].Product.Name)

	orderData := fmt.Sprintf("product_id=%d", productID)
	body, err = c.Post("/order/server_market/transaction", nil)
	if err != nil {
		// Fallback: try with form data
		t.Fatalf("Error ordering server (tried product_id=%d): %s\n%s", productID, err, orderData)
	}

	var orderResp struct {
		Transaction struct {
			ServerNumber int    `json:"server_number"`
			Status       string `json:"status"`
		} `json:"transaction"`
	}
	if err := json.Unmarshal(body, &orderResp); err != nil {
		t.Fatalf("Error parsing order response: %s", err)
	}

	serverNumber := fmt.Sprintf("%d", orderResp.Transaction.ServerNumber)
	t.Logf("Ordered server %s, waiting for it to be ready...", serverNumber)

	// Poll until server is ready (up to 30 minutes).
	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		body, err := c.Get("/server/" + serverNumber)
		if err != nil {
			t.Logf("Server %s not yet available: %s", serverNumber, err)
			time.Sleep(30 * time.Second)
			continue
		}

		var serverResp struct {
			Server struct {
				Status string `json:"status"`
			} `json:"server"`
		}
		if err := json.Unmarshal(body, &serverResp); err == nil && serverResp.Server.Status == "ready" {
			t.Logf("Server %s is ready", serverNumber)
			break
		}
		time.Sleep(30 * time.Second)
	}

	// Register cleanup to cancel the server.
	t.Cleanup(func() {
		t.Logf("Cancelling server %s", serverNumber)
		c := testAccNewClient(t)
		_, err := c.Post("/server/"+serverNumber+"/cancellation", nil)
		if err != nil {
			t.Logf("Warning: failed to cancel server %s: %s", serverNumber, err)
		}
	})

	return serverNumber
}

// testAccServerIP queries the API to get a server's main IP address.
func testAccServerIP(t *testing.T, serverNumber string) string {
	t.Helper()
	c := testAccNewClient(t)

	body, err := c.Get("/server/" + serverNumber)
	if err != nil {
		t.Fatalf("Error reading server %s: %s", serverNumber, err)
	}

	var resp struct {
		Server struct {
			ServerIP string `json:"server_ip"`
		} `json:"server"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("Error parsing server response: %s", err)
	}
	if resp.Server.ServerIP == "" {
		t.Fatalf("Server %s has no IP address", serverNumber)
	}
	return resp.Server.ServerIP
}

// --- Environment variable gates ---

func testAccServerNumber(t *testing.T) string {
	t.Helper()
	v := os.Getenv("HETZNER_TEST_SERVER_NUMBER")
	if v == "" {
		t.Skip("HETZNER_TEST_SERVER_NUMBER not set; skipping")
	}
	return v
}

func testAccStorageBoxID(t *testing.T) string {
	t.Helper()
	v := os.Getenv("HETZNER_TEST_STORAGEBOX_ID")
	if v == "" {
		t.Skip("HETZNER_TEST_STORAGEBOX_ID not set; skipping")
	}
	return v
}

func testAccVSwitchCreateEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("HETZNER_TEST_VSWITCH_CREATE") != "1" {
		t.Skip("HETZNER_TEST_VSWITCH_CREATE not set to 1; skipping")
	}
}

func testAccFailoverIP(t *testing.T) string {
	t.Helper()
	v := os.Getenv("HETZNER_TEST_FAILOVER_IP")
	if v == "" {
		t.Skip("HETZNER_TEST_FAILOVER_IP not set; skipping")
	}
	return v
}

// --- Terraform config templates ---

func testAccSSHKeyConfig(name, publicKey string) string {
	return fmt.Sprintf(`
resource "hetzner_ssh_key" "test" {
  name = %q
  data = %q
}
`, name, publicKey)
}

func testAccSSHKeyDataSourceConfig(fingerprint string) string {
	return fmt.Sprintf(`
data "hetzner_ssh_key" "test" {
  fingerprint = %q
}
`, fingerprint)
}

func testAccSSHKeysDataSourceConfig() string {
	return `
data "hetzner_ssh_keys" "test" {
}
`
}

func testAccRDNSConfig(ip, ptr string) string {
	return fmt.Sprintf(`
resource "hetzner_rdns" "test" {
  ip  = %q
  ptr = %q
}
`, ip, ptr)
}

func testAccRDNSDataSourceConfig(ip string) string {
	return fmt.Sprintf(`
data "hetzner_rdns" "test" {
  ip = %q
}
`, ip)
}

func testAccFirewallConfig(serverNumber, status string) string {
	return fmt.Sprintf(`
resource "hetzner_firewall" "test" {
  server_number = %q
  status        = %q
  allowlist_hos = true
  filter_ipv6   = false

  input = [{
    name       = "Allow SSH"
    ip_version = "ipv4"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }]
}
`, serverNumber, status)
}

func testAccFirewallUpdatedConfig(serverNumber string) string {
	return fmt.Sprintf(`
resource "hetzner_firewall" "test" {
  server_number = %q
  status        = "active"
  allowlist_hos = true
  filter_ipv6   = false

  input = [{
    name       = "Allow SSH"
    ip_version = "ipv4"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }, {
    name       = "Allow HTTP"
    ip_version = "ipv4"
    dst_port   = "80"
    protocol   = "tcp"
    action     = "accept"
  }]
}
`, serverNumber)
}

func testAccFirewallDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_firewall" "test" {
  server_number = %q
}
`, serverNumber)
}

func testAccFirewallTemplateConfig(name string) string {
	return fmt.Sprintf(`
resource "hetzner_firewall_template" "test" {
  name          = %q
  filter_ipv6   = false
  allowlist_hos = true
  is_default    = false

  input = [{
    name       = "Allow SSH"
    ip_version = "ipv4"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }]
}
`, name)
}

func testAccFirewallTemplateUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "hetzner_firewall_template" "test" {
  name          = %q
  filter_ipv6   = true
  allowlist_hos = true
  is_default    = false

  input = [{
    name       = "Allow SSH"
    ip_version = "ipv4"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }, {
    name       = "Allow HTTPS"
    ip_version = "ipv4"
    dst_port   = "443"
    protocol   = "tcp"
    action     = "accept"
  }]
}
`, name)
}

func testAccFirewallTemplateDataSourceConfig(templateID string) string {
	return fmt.Sprintf(`
data "hetzner_firewall_template" "test" {
  id = %s
}
`, templateID)
}

func testAccFirewallTemplatesDataSourceConfig() string {
	return `
data "hetzner_firewall_templates" "test" {
}
`
}

func testAccServerConfig(serverNumber, name string) string {
	return fmt.Sprintf(`
resource "hetzner_server" "test" {
  server_number = %s
  server_name   = %q
}
`, serverNumber, name)
}

func testAccServerDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_server" "test" {
  server_number = %s
}
`, serverNumber)
}

func testAccServersDataSourceConfig() string {
	return `
data "hetzner_servers" "test" {
}
`
}

func testAccIPDataSourceConfig(ip string) string {
	return fmt.Sprintf(`
data "hetzner_ip" "test" {
  ip = %q
}
`, ip)
}

func testAccIPsDataSourceConfig() string {
	return `
data "hetzner_ips" "test" {
}
`
}

func testAccSubnetsDataSourceConfig() string {
	return `
data "hetzner_subnets" "test" {
}
`
}

func testAccBootRescueConfig(serverNumber, osType string) string {
	return fmt.Sprintf(`
resource "hetzner_boot_rescue" "test" {
  server_number = %s
  os            = %q
}
`, serverNumber, osType)
}

func testAccBootRescueDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_boot_rescue" "test" {
  server_number = %s
}
`, serverNumber)
}

func testAccResetDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_reset" "test" {
  server_number = %s
}
`, serverNumber)
}

func testAccWOLDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_wol" "test" {
  server_number = %s
}
`, serverNumber)
}

func testAccVSwitchConfig(name string, vlan int) string {
	return fmt.Sprintf(`
resource "hetzner_vswitch" "test" {
  name = %q
  vlan = %d
}
`, name, vlan)
}

func testAccVSwitchDataSourceConfig(vswitchID string) string {
	return fmt.Sprintf(`
data "hetzner_vswitch" "test" {
  id = %s
}
`, vswitchID)
}

func testAccVSwitchesDataSourceConfig() string {
	return `
data "hetzner_vswitches" "test" {
}
`
}

func testAccStorageBoxConfig(storageBoxID, name string, webdav, samba, ssh, externalReachability, zfs bool) string {
	return fmt.Sprintf(`
resource "hetzner_storagebox" "test" {
  storagebox_id         = %s
  storagebox_name       = %q
  webdav                = %t
  samba                 = %t
  ssh                   = %t
  external_reachability = %t
  zfs                   = %t
}
`, storageBoxID, name, webdav, samba, ssh, externalReachability, zfs)
}

func testAccStorageBoxDataSourceConfig(storageBoxID string) string {
	return fmt.Sprintf(`
data "hetzner_storagebox" "test" {
  storagebox_id = %s
}
`, storageBoxID)
}

func testAccStorageBoxesDataSourceConfig() string {
	return `
data "hetzner_storageboxes" "test" {
}
`
}

func testAccStorageBoxSnapshotplanConfig(storageBoxID string, status string, hour, minute int) string {
	return fmt.Sprintf(`
resource "hetzner_storagebox_snapshotplan" "test" {
  storagebox_id = %s
  status        = %q
  hour          = %d
  minute        = %d
  day_of_week   = 1
  day_of_month  = 1
  month         = 1
}
`, storageBoxID, status, hour, minute)
}

func testAccStorageBoxSubaccountConfig(storageBoxID, homedirectory, comment string, ssh, readonly bool) string {
	return fmt.Sprintf(`
resource "hetzner_storagebox_subaccount" "test" {
  storagebox_id         = %s
  homedirectory         = %q
  comment               = %q
  ssh                   = %t
  webdav                = false
  samba                 = false
  external_reachability = false
  readonly              = %t
  createdir             = true
}
`, storageBoxID, homedirectory, comment, ssh, readonly)
}

func testAccFailoverDataSourceConfig(failoverIP string) string {
	return fmt.Sprintf(`
data "hetzner_failover" "test" {
  ip = %q
}
`, failoverIP)
}

func testAccFailoversDataSourceConfig() string {
	return `
data "hetzner_failovers" "test" {
}
`
}
