// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccVSwitch_CRUD tests the full vSwitch lifecycle.
func TestAccVSwitch_CRUD(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccVSwitchConfig("acc-test-vswitch", 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "name", "acc-test-vswitch"),
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "vlan", "4000"),
					resource.TestCheckResourceAttrSet("hetzner_vswitch.test", "id"),
				),
			},
			// Import.
			{
				ResourceName:      "hetzner_vswitch.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update name.
			{
				Config: testAccVSwitchConfig("acc-test-vswitch-renamed", 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_vswitch.test", "name", "acc-test-vswitch-renamed"),
				),
			},
		},
	})
}

// TestAccVSwitch_DataSources tests vSwitch data sources.
func TestAccVSwitch_DataSources(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVSwitchConfig("acc-ds-test-vswitch", 4001) + "\n" + testAccVSwitchesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_vswitches.test", "vswitches.#"),
				),
			},
		},
	})
}

// TestAccVSwitch_DataSource reads a single vSwitch via singular data source.
func TestAccVSwitch_DataSource(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a vSwitch, then read it back via the singular data source.
				Config: testAccVSwitchConfig("acc-singular-ds-test", 4002) + `
data "hetzner_vswitch" "test" {
  id = hetzner_vswitch.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetzner_vswitch.test", "name", "acc-singular-ds-test"),
					resource.TestCheckResourceAttr("data.hetzner_vswitch.test", "vlan", "4002"),
				),
			},
		},
	})
}
