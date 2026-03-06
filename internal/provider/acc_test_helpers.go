// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/zack/terraform-provider-hetzner/internal/client"
)

// cachedServerNumber stores the resolved server number across tests within a single test run.
var (
	cachedServerNumber string
	cachedServerOnce   sync.Once
)

// Default persistent test server and failover IP.
// These are long-lived resources kept on the account for acceptance tests.
const (
	defaultTestServerNumber = "2940646"
	defaultTestFailoverIP   = "88.99.239.234"
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

// serverMarketProduct represents a server from the Hetzner server auction.
type serverMarketProduct struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	CPU         string   `json:"cpu"`
	RAMSize     int      `json:"ram_size"`
	HDDSize     int      `json:"hdd_size"`
	Price       float64  `json:"price"`
	HourlyPrice float64  `json:"hourly_price"`
	Datacenter  string   `json:"datacenter"`
	FixedPrice  bool     `json:"fixed_price"`
	Specials    []string `json:"specials"`
}

// hasIPv4 checks if the specials list includes "IPv4".
func hasIPv4(specials []string) bool {
	for _, s := range specials {
		if s == "IPv4" {
			return true
		}
	}
	return false
}

// testAccFindCheapestServer queries the public Hetzner server auction endpoint
// and returns the product ID of the cheapest hourly-billed server.
func testAccFindCheapestServer(t *testing.T) int {
	t.Helper()

	resp, err := http.Get("https://www.hetzner.com/_resources/app/data/app/live_data_sb_EUR.json")
	if err != nil {
		t.Fatalf("Error fetching server auction data: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading auction response: %s", err)
	}

	// The auction JSON wraps servers under a "server" key.
	var wrapper struct {
		Server []serverMarketProduct `json:"server"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		t.Fatalf("Error parsing auction data: %s", err)
	}
	products := wrapper.Server

	// Filter to hourly-priced servers with IPv4 and sort by hourly price.
	var hourly []serverMarketProduct
	for _, p := range products {
		if p.HourlyPrice > 0 && hasIPv4(p.Specials) {
			hourly = append(hourly, p)
		}
	}
	if len(hourly) == 0 {
		t.Fatal("No hourly-billed servers with IPv4 available in the auction")
	}

	sort.Slice(hourly, func(i, j int) bool {
		return hourly[i].HourlyPrice < hourly[j].HourlyPrice
	})

	cheapest := hourly[0]
	t.Logf("Cheapest hourly server: ID=%d, %s, %dGB RAM, %dGB HDD, €%.4f/hr (€%.2f/mo), DC=%s",
		cheapest.ID, cheapest.CPU, cheapest.RAMSize, cheapest.HDDSize,
		cheapest.HourlyPrice, cheapest.Price, cheapest.Datacenter)

	return cheapest.ID
}

// testAccGetOrCreateServer returns the persistent test server number.
// Uses HETZNER_TEST_SERVER_NUMBER env var if set, otherwise checks
// the default server, then falls back to the first server on the account.
// Never orders a server — that is only done by server order tests.
// The result is cached for the duration of the test run.
func testAccGetOrCreateServer(t *testing.T) string {
	t.Helper()

	if v := os.Getenv("HETZNER_TEST_SERVER_NUMBER"); v != "" {
		return v
	}

	cachedServerOnce.Do(func() {
		c := testAccNewClient(t)

		// Check if the default server exists.
		if _, err := c.Get("/server/" + defaultTestServerNumber); err == nil {
			cachedServerNumber = defaultTestServerNumber
			return
		}

		// Fall back to the first server on the account.
		body, err := c.Get("/server")
		if err == nil {
			var servers []struct {
				Server struct {
					ServerNumber int    `json:"server_number"`
					Status       string `json:"status"`
				} `json:"server"`
			}
			if json.Unmarshal(body, &servers) == nil && len(servers) > 0 {
				cachedServerNumber = fmt.Sprintf("%d", servers[0].Server.ServerNumber)
				t.Logf("Using existing server %s (status: %s)", cachedServerNumber, servers[0].Server.Status)
				return
			}
		}
	})

	if cachedServerNumber == "" {
		t.Skip("No test server available; set HETZNER_TEST_SERVER_NUMBER or ensure a server exists on the account")
	}
	return cachedServerNumber
}

// testAccOrderServer orders the cheapest hourly server from the auction
// and waits for it to be ready. Returns the server number.
func testAccOrderServer(t *testing.T) (string, error) {
	t.Helper()

	// Determine which product to order.
	var productID string
	if v := os.Getenv("HETZNER_TEST_SERVER_MARKET_PRODUCT_ID"); v != "" {
		productID = v
		t.Logf("Using specified server market product: %s", productID)
	} else {
		cheapestID := testAccFindCheapestServer(t)
		productID = fmt.Sprintf("%d", cheapestID)
	}

	c := testAccNewClient(t)

	// Get the first available SSH key fingerprint for the order.
	keysBody, err := c.Get("/key")
	var keyFingerprint string
	if err == nil {
		var keys []struct {
			Key struct {
				Fingerprint string `json:"fingerprint"`
			} `json:"key"`
		}
		if json.Unmarshal(keysBody, &keys) == nil && len(keys) > 0 {
			keyFingerprint = keys[0].Key.Fingerprint
		}
	}
	if keyFingerprint == "" {
		return "", fmt.Errorf("no SSH keys found on the account; upload one before ordering a server")
	}

	// Order the server via the Robot API with IPv4 addon.
	t.Logf("Ordering server market product %s via Robot API (with IPv4 addon)", productID)
	form := url.Values{}
	form.Set("product_id", productID)
	form.Add("authorized_key[]", keyFingerprint)
	form.Add("addon[]", "primary_ipv4")

	body, err := c.Post("/order/server_market/transaction", form)
	if err != nil {
		return "", fmt.Errorf("error ordering server market product %s: %w", productID, err)
	}

	// Parse order response to get transaction ID and optionally the server number.
	var orderResp struct {
		Transaction struct {
			ID           string `json:"id"`
			ServerNumber *int   `json:"server_number"`
			Status       string `json:"status"`
		} `json:"transaction"`
	}
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return "", fmt.Errorf("error parsing order response: %w\nBody: %s", err, string(body))
	}

	txnID := orderResp.Transaction.ID
	t.Logf("Order transaction %s, status: %s", txnID, orderResp.Transaction.Status)

	// Poll the transaction until server_number is assigned (may take a few minutes).
	var serverNumber string
	if orderResp.Transaction.ServerNumber != nil && *orderResp.Transaction.ServerNumber != 0 {
		serverNumber = fmt.Sprintf("%d", *orderResp.Transaction.ServerNumber)
	} else {
		deadline := time.Now().Add(20 * time.Minute)
		for time.Now().Before(deadline) {
			time.Sleep(15 * time.Second)
			txnBody, err := c.Get("/order/server_market/transaction/" + txnID)
			if err != nil {
				t.Logf("Polling transaction %s: %s", txnID, err)
				continue
			}
			var txnResp struct {
				Transaction struct {
					ServerNumber *int   `json:"server_number"`
					Status       string `json:"status"`
				} `json:"transaction"`
			}
			if err := json.Unmarshal(txnBody, &txnResp); err != nil {
				t.Logf("Error parsing transaction response: %s", err)
				continue
			}
			t.Logf("Transaction %s status: %s, server_number: %v", txnID, txnResp.Transaction.Status, txnResp.Transaction.ServerNumber)
			if txnResp.Transaction.ServerNumber != nil && *txnResp.Transaction.ServerNumber != 0 {
				serverNumber = fmt.Sprintf("%d", *txnResp.Transaction.ServerNumber)
				break
			}
		}
		if serverNumber == "" {
			return "", fmt.Errorf("timed out waiting for server_number from transaction %s", txnID)
		}
	}

	t.Logf("Ordered server %s, waiting for it to be ready...", serverNumber)

	// Poll until server is ready (up to 30 minutes for dedicated server provisioning).
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
			return serverNumber, nil
		}
		t.Logf("Server %s status: %s, waiting...", serverNumber, serverResp.Server.Status)
		time.Sleep(30 * time.Second)
	}

	return serverNumber, nil
}

// testAccServerIP queries the API to get a server's main IP address.
// Skips the test if the server has no IPv4 address (e.g., IPv6-only servers).
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
		t.Skipf("Server %s has no IPv4 address (IPv6-only); skipping test that requires IPv4", serverNumber)
	}
	return resp.Server.ServerIP
}

// --- Environment variable gates ---

func testAccFailoverIP(t *testing.T) string {
	t.Helper()
	if v := os.Getenv("HETZNER_TEST_FAILOVER_IP"); v != "" {
		return v
	}

	// Verify the default failover IP actually exists on the account.
	c := testAccNewClient(t)
	_, err := c.Get("/failover/" + defaultTestFailoverIP)
	if err != nil {
		t.Skip("No failover IP available on account; skipping")
	}
	return defaultTestFailoverIP
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

  output = [{
    name   = "Allow all"
    action = "accept"
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

  output = [{
    name   = "Allow all"
    action = "accept"
  }]
}
`, name)
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

func testAccVSwitchesDataSourceConfig() string {
	return `
data "hetzner_vswitches" "test" {
}
`
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

func testAccBootLinuxConfig(serverNumber, dist, lang string) string {
	return fmt.Sprintf(`
resource "hetzner_boot_linux" "test" {
  server_number = %s
  dist          = %q
  lang          = %q
}
`, serverNumber, dist, lang)
}

func testAccBootLinuxDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_boot_linux" "test" {
  server_number = %s
}
`, serverNumber)
}

func testAccBootVNCDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_boot_vnc" "test" {
  server_number = %s
}
`, serverNumber)
}

func testAccVSwitchServerConfig(vswitchID, serverNumber string) string {
	return fmt.Sprintf(`
resource "hetzner_vswitch_server" "test" {
  vswitch_id    = %s
  server_number = %s
}
`, vswitchID, serverNumber)
}

func testAccIPConfig(ip string, trafficWarnings bool, hourly, daily, monthly int) string {
	return fmt.Sprintf(`
resource "hetzner_ip" "test" {
  ip               = %q
  traffic_warnings = %t
  traffic_hourly   = %d
  traffic_daily    = %d
  traffic_monthly  = %d
}
`, ip, trafficWarnings, hourly, daily, monthly)
}

func testAccSubnetConfig(ip string, trafficWarnings bool, hourly, daily, monthly int) string {
	return fmt.Sprintf(`
resource "hetzner_subnet" "test" {
  ip               = %q
  traffic_warnings = %t
  traffic_hourly   = %d
  traffic_daily    = %d
  traffic_monthly  = %d
}
`, ip, trafficWarnings, hourly, daily, monthly)
}

func testAccSubnetDataSourceConfig(ip string) string {
	return fmt.Sprintf(`
data "hetzner_subnet" "test" {
  ip = %q
}
`, ip)
}

func testAccFailoverConfig(ip, activeServerIP string) string {
	return fmt.Sprintf(`
resource "hetzner_failover" "test" {
  ip               = %q
  active_server_ip = %q
}
`, ip, activeServerIP)
}

func testAccBootVNCConfig(serverNumber, dist, lang string) string {
	return fmt.Sprintf(`
resource "hetzner_boot_vnc" "test" {
  server_number = %s
  dist          = %q
  lang          = %q
}
`, serverNumber, dist, lang)
}

func testAccBootWindowsConfig(serverNumber, dist, lang string) string {
	return fmt.Sprintf(`
resource "hetzner_boot_windows" "test" {
  server_number = %s
  dist          = %q
  lang          = %q
}
`, serverNumber, dist, lang)
}

func testAccBootWindowsDataSourceConfig(serverNumber string) string {
	return fmt.Sprintf(`
data "hetzner_boot_windows" "test" {
  server_number = %s
}
`, serverNumber)
}

// testAccRequireBootOption skips the test if the given boot option (vnc, windows, etc.)
// is not available on the server. Checks GET /boot/{server_number}/{option}.
func testAccRequireBootOption(t *testing.T, serverNumber, option string) {
	t.Helper()
	c := testAccNewClient(t)
	_, err := c.Get(fmt.Sprintf("/boot/%s/%s", serverNumber, option))
	if err != nil {
		t.Skipf("Boot option %q not available on server %s: %s", option, serverNumber, err)
	}
}

// testAccFirstBootDist queries the boot API for the given option and returns the first
// available distribution. Skips if no distributions are available.
func testAccFirstBootDist(t *testing.T, serverNumber, option string) string {
	t.Helper()
	c := testAccNewClient(t)
	body, err := c.Get(fmt.Sprintf("/boot/%s/%s", serverNumber, option))
	if err != nil {
		t.Skipf("Boot option %q not available on server %s: %s", option, serverNumber, err)
	}

	// The response shape is {"<option>": {"dist": [...], ...}}.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("Error parsing boot %s response: %s", option, err)
	}
	optionData, ok := raw[option]
	if !ok {
		t.Skipf("No %q key in boot response", option)
	}
	var details struct {
		Dist interface{} `json:"dist"`
	}
	if err := json.Unmarshal(optionData, &details); err != nil {
		t.Fatalf("Error parsing boot %s details: %s", option, err)
	}

	// dist can be a list of strings or null.
	switch v := details.Dist.(type) {
	case []interface{}:
		if len(v) == 0 {
			t.Skipf("No distributions available for boot %s on server %s", option, serverNumber)
		}
		return fmt.Sprintf("%v", v[0])
	default:
		t.Skipf("No distributions available for boot %s on server %s", option, serverNumber)
	}
	return ""
}

// testAccSubnetIP queries the API for the first available subnet IP.
// Skips the test if no subnets exist on the account.
func testAccSubnetIP(t *testing.T) string {
	t.Helper()
	c := testAccNewClient(t)

	body, err := c.Get("/subnet")
	if err != nil {
		t.Skipf("No subnets available (GET /subnet failed: %s); skipping", err)
	}

	var subnets []struct {
		Subnet struct {
			IP string `json:"ip"`
		} `json:"subnet"`
	}
	if err := json.Unmarshal(body, &subnets); err != nil || len(subnets) == 0 {
		t.Skip("No subnets available on the account; skipping")
	}
	return subnets[0].Subnet.IP
}
