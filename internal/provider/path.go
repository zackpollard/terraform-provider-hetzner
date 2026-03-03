// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import "github.com/hashicorp/terraform-plugin-framework/path"

// frameworkPath is a helper to create a path.Path from a string attribute name.
func frameworkPath(attr string) path.Path {
	return path.Root(attr)
}
