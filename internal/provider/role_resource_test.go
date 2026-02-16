package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRoleResourceConfig(
					"test-role",
					"Test Role Description",
					[]string{"read", "update"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_role.test", "name", "test-role"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "description", "Test Role Description"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "member_permissions.#", "2"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "member_roles.#", "0"),
					resource.TestCheckResourceAttrSet("braintrustdata_role.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_role.test", "org_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_role.test", "created"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccRoleResourceConfigWithMemberRole(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_role.test", "name", "test-role"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "description", "Updated Description"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "member_permissions.#", "1"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "member_roles.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRoleResourceConfig(name, description string, memberPermissions []string) string {
	quotedPermissions := make([]string, 0, len(memberPermissions))
	for _, permission := range memberPermissions {
		quotedPermissions = append(quotedPermissions, fmt.Sprintf("%q", permission))
	}

	return fmt.Sprintf(`
resource "braintrustdata_role" "test" {
  name        = %[1]q
  description = %[2]q
  member_permissions = [%[3]s]
}
`, name, description, strings.Join(quotedPermissions, ", "))
}

func testAccRoleResourceConfigWithMemberRole() string {
	return `
resource "braintrustdata_role" "member" {
  name        = "test-member-role"
  description = "Role used as member role"
  member_permissions = ["read"]
}

resource "braintrustdata_role" "test" {
  name        = "test-role"
  description = "Updated Description"
  member_permissions = ["delete"]
  member_roles = [braintrustdata_role.member.id]
}
`
}
