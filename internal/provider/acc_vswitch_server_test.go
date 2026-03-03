// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccCreateVSwitch creates a vSwitch via the API and returns its ID.
// It registers a cleanup function to delete the vSwitch when the test finishes.
func testAccCreateVSwitch(t *testing.T, name string, vlan int) string {
	t.Helper()
	c := testAccNewClient(t)

	form := url.Values{}
	form.Set("name", name)
	form.Set("vlan", fmt.Sprintf("%d", vlan))

	body, err := c.Post("/vswitch", form)
	if err != nil {
		t.Fatalf("Error creating vSwitch: %s", err)
	}

	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("Error parsing vSwitch response: %s\nBody: %s", err, string(body))
	}

	vswitchID := fmt.Sprintf("%d", resp.ID)
	t.Logf("Created vSwitch %s (name=%s, vlan=%d)", vswitchID, name, vlan)

	t.Cleanup(func() {
		c := testAccNewClient(t)

		// First, remove any attached servers.
		vsBody, err := c.Get("/vswitch/" + vswitchID)
		if err == nil {
			var detail struct {
				Server []struct {
					ServerNumber int `json:"server_number"`
				} `json:"server"`
			}
			if json.Unmarshal(vsBody, &detail) == nil {
				for _, s := range detail.Server {
					t.Logf("Removing server %d from vSwitch %s", s.ServerNumber, vswitchID)
					params := url.Values{}
					params.Set("server", fmt.Sprintf("%d", s.ServerNumber))
					_, _ = c.DeleteWithBody(fmt.Sprintf("/vswitch/%s/server", vswitchID), params)
				}
				if len(detail.Server) > 0 {
					// Wait for server removal to propagate.
					time.Sleep(5 * time.Second)
				}
			}
		}

		t.Logf("Cancelling vSwitch %s", vswitchID)
		cancelParams := url.Values{}
		cancelParams.Set("cancellation_date", "now")
		_, err = c.DeleteWithBody("/vswitch/"+vswitchID, cancelParams)
		if err != nil {
			t.Logf("Warning: failed to cancel vSwitch %s: %s", vswitchID, err)
		}
	})

	return vswitchID
}

// TestAccVSwitchServer_CRUD tests adding and removing a server from a vSwitch.
func TestAccVSwitchServer_CRUD(t *testing.T) {
	serverNumber := testAccGetOrCreateServer(t)
	vswitchID := testAccCreateVSwitch(t, "acc-test-vswitch-server", 4090)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Add server to vSwitch.
			{
				Config: testAccVSwitchServerConfig(vswitchID, serverNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_vswitch_server.test", "vswitch_id", vswitchID),
					resource.TestCheckResourceAttr("hetzner_vswitch_server.test", "server_number", serverNumber),
					resource.TestCheckResourceAttrSet("hetzner_vswitch_server.test", "status"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_vswitch_server.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vswitch_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["hetzner_vswitch_server.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					vsID := rs.Primary.Attributes["vswitch_id"]
					sn := rs.Primary.Attributes["server_number"]
					return fmt.Sprintf("%s/%s", vsID, sn), nil
				},
			},
			// Destroy removes server from vSwitch.
		},
	})
}
