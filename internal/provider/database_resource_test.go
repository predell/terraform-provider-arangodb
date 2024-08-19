// Copyright (c) Predell Services
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseResourceConfig("database_name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arangodb_database.test", "name", "database_name"),
				),
			},
			//// ImportState testing
			//{
			//	ResourceName:      "arangodb_database.test",
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
			// Update and Read testing
			{
				Config: testAccDatabaseResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arangodb_database.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDatabaseResourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "arangodb_database" "test" {
  name = %[1]q
}
`, name)
}
