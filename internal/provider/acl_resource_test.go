package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccACLResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccACLResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "object_id"),
					resource.TestCheckResourceAttr("braintrustdata_acl.test", "object_type", "project"),
					resource.TestCheckResourceAttr("braintrustdata_acl.test", "permission", "read"),
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "user_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "created"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_acl.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccACLResource_WithGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccACLResourceConfigWithGroup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "group_id"),
					resource.TestCheckResourceAttr("braintrustdata_acl.test", "object_type", "project"),
					resource.TestCheckResourceAttr("braintrustdata_acl.test", "permission", "update"),
				),
			},
		},
	})
}

func TestAccACLResource_WithRole(t *testing.T) {
	t.Skip("TODO: Role-based ACLs require roles to have member_permissions defined. " +
		"The role resource doesn't currently support setting permissions, so this test " +
		"fails with API 500 error. Implement role permissions support before enabling this test.")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccACLResourceConfigWithRole(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_acl.test", "role_id"),
					resource.TestCheckResourceAttr("braintrustdata_acl.test", "object_type", "project"),
				),
			},
		},
	})
}

func testAccACLResourceConfig() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-acl-project"
  description = "Project for ACL testing"
}

resource "braintrustdata_acl" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  user_id     = "866a8a8a-fee9-4a5b-8278-12970de499c2"  # Real user ID
  permission  = "read"
}
`
}

func testAccACLResourceConfigWithGroup() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-acl-project-group"
  description = "Project for ACL group testing"
}

resource "braintrustdata_group" "test" {
  name        = "test-acl-group"
  description = "Group for ACL testing"
}

resource "braintrustdata_acl" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  group_id    = braintrustdata_group.test.id
  permission  = "update"
}
`
}

func testAccACLResourceConfigWithRole() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-acl-project-role"
  description = "Project for ACL role testing"
}

resource "braintrustdata_role" "test" {
  name        = "test-acl-role"
  description = "Role for ACL testing"
}

resource "braintrustdata_acl" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  role_id     = braintrustdata_role.test.id
}
`
}
