// Copyright (c) Predell Services
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	config := testUserResourceConfig("one")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arangodb_user.test", "active", "true"),
					resource.TestCheckResourceAttr("arangodb_user.test", "user", "one"),
					resource.TestCheckResourceAttr("arangodb_user.test", "password", "1234"),
				),
			},
			//// ImportState testing
			//{
			//	Config:            config,
			//	ResourceName:      "arangodb_user.test",
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
			// Update and Read testing
			{
				Config: testUserResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("arangodb_user.test", "user", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testUserResourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "arangodb_user" "test" {
  active = true
  user = %[1]q
  password = "1234"
}
`, name)
}
