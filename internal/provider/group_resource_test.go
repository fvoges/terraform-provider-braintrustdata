package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGroupResourceConfig("test-group", "Test Group Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_group.test", "name", "test-group"),
					resource.TestCheckResourceAttr("braintrustdata_group.test", "description", "Test Group Description"),
					resource.TestCheckResourceAttrSet("braintrustdata_group.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_group.test", "created"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccGroupResourceConfig("test-group-updated", "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_group.test", "name", "test-group-updated"),
					resource.TestCheckResourceAttr("braintrustdata_group.test", "description", "Updated Description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccGroupResource_WithMembers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a group with members (both user and group)
				Config: testAccGroupResourceConfigWithMembers(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_group.test", "name", "test-group-members"),
					resource.TestCheckResourceAttr("braintrustdata_group.test", "member_users.#", "1"),
					resource.TestCheckResourceAttr("braintrustdata_group.test", "member_groups.#", "1"),
					resource.TestCheckResourceAttr("braintrustdata_group.member", "name", "test-member-group"),
				),
			},
		},
	})
}

func testAccGroupResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "braintrustdata_group" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccGroupResourceConfigWithMembers() string {
	return `
resource "braintrustdata_group" "member" {
  name        = "test-member-group"
  description = "Group used as a member"
}

resource "braintrustdata_group" "test" {
  name         = "test-group-members"
  description  = "Group with both user and group members"
  # API uses separate fields for user and group members
  member_users = ["866a8a8a-fee9-4a5b-8278-12970de499c2"]  # Real user ID (TODO: replace with data source)
  member_groups = [braintrustdata_group.member.id]
}
`
}

func testAccPreCheck(_ *testing.T) {
	// This function is called before running acceptance tests
	// Check for required environment variables
	// Currently no-op since provider handles validation
}
