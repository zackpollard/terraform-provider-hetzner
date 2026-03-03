// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// --- Storage box resource tests ---

// TestAccStoragebox_Settings tests updating storage box settings.
func TestAccStoragebox_Settings(t *testing.T) {
	storageBoxID := testAccStorageBoxID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set initial config.
			{
				Config: testAccStorageBoxConfig(storageBoxID, "acc-test-box", false, false, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "ssh", "true"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "webdav", "false"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "samba", "false"),
					resource.TestCheckResourceAttrSet("hetzner_storagebox.test", "disk_quota"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_storagebox.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "storagebox_id",
			},
			// Update: enable webdav, change name.
			{
				Config: testAccStorageBoxConfig(storageBoxID, "acc-test-box-updated", true, false, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "storagebox_name", "acc-test-box-updated"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "webdav", "true"),
					resource.TestCheckResourceAttr("hetzner_storagebox.test", "ssh", "true"),
				),
			},
		},
	})
}

// TestAccStoragebox_DataSources tests data sources for storage boxes.
func TestAccStoragebox_DataSources(t *testing.T) {
	storageBoxID := testAccStorageBoxID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read single storage box.
			{
				Config: testAccStorageBoxDataSourceConfig(storageBoxID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_storagebox.test", "storagebox_name"),
					resource.TestCheckResourceAttrSet("data.hetzner_storagebox.test", "disk_quota"),
					resource.TestCheckResourceAttrSet("data.hetzner_storagebox.test", "status"),
				),
			},
			// List all storage boxes.
			{
				Config: testAccStorageBoxesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetzner_storageboxes.test", "storageboxes.#"),
				),
			},
		},
	})
}

// --- Snapshot plan tests ---

// TestAccStorageboxSnapshotplan_CRUD tests configuring and updating a snapshot plan.
func TestAccStorageboxSnapshotplan_CRUD(t *testing.T) {
	storageBoxID := testAccStorageBoxID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create snapshot plan.
			{
				Config: testAccStorageBoxSnapshotplanConfig(storageBoxID, "enabled", 2, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "status", "enabled"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "hour", "2"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "minute", "30"),
				),
			},
			// Import.
			{
				ResourceName:                         "hetzner_storagebox_snapshotplan.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "storagebox_id",
			},
			// Update: change time.
			{
				Config: testAccStorageBoxSnapshotplanConfig(storageBoxID, "enabled", 4, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "hour", "4"),
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "minute", "0"),
				),
			},
			// Disable.
			{
				Config: testAccStorageBoxSnapshotplanConfig(storageBoxID, "disabled", 4, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_snapshotplan.test", "status", "disabled"),
				),
			},
		},
	})
}

// --- Subaccount tests ---

// TestAccStorageboxSubaccount_CRUD tests creating, updating, and deleting a sub-account.
func TestAccStorageboxSubaccount_CRUD(t *testing.T) {
	storageBoxID := testAccStorageBoxID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create sub-account.
			{
				Config: testAccStorageBoxSubaccountConfig(storageBoxID, "/acc-test", "acceptance test subaccount", true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetzner_storagebox_subaccount.test", "username"),
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "homedirectory", "/acc-test"),
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "ssh", "true"),
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "readonly", "false"),
				),
			},
			// Update: toggle readonly.
			{
				Config: testAccStorageBoxSubaccountConfig(storageBoxID, "/acc-test", "updated comment", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "readonly", "true"),
					resource.TestCheckResourceAttr("hetzner_storagebox_subaccount.test", "comment", "updated comment"),
				),
			},
			// Destroy will delete the sub-account.
		},
	})
}
