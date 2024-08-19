// Copyright (c) Predell Services
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserPermissionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testUserPermissionResourceConfig("database_name", "rw", "user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arangodb_user_permission.test", "database", "database_name"),
					resource.TestCheckResourceAttr("arangodb_user_permission.test", "permission", "rw"),
					resource.TestCheckResourceAttr("arangodb_user_permission.test", "user", "user"),
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
				Config: testUserPermissionResourceConfig("database_name", "ro", "user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arangodb_user_permission.test", "permission", "ro"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testUserPermissionResourceConfig(databaseName string, permission string, userName string) string {
	return providerConfig + fmt.Sprintf(`
resource "arangodb_user" "test" {
  active   = true
  user     = %[3]q
  password = "1234"
}

resource "arangodb_database" "test" {
  name = %[1]q
}

resource "arangodb_user_permission" "test" {
  database = arangodb_database.test.name
  permission = %[2]q
  user = arangodb_user.test.user
}
`, databaseName, permission, userName)
}
