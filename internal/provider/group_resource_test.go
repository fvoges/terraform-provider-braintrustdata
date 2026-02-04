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
				Config: testAccGroupResourceConfigWithMembers(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_group.test", "name", "test-group-members"),
					resource.TestCheckResourceAttr("braintrustdata_group.test", "member_ids.#", "2"),
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
resource "braintrustdata_group" "test" {
  name        = "test-group-members"
  description = "Group with members"
  member_ids  = ["user-1", "user-2"]
}
`
}

func testAccPreCheck(_ *testing.T) {
	// This function is called before running acceptance tests
	// Check for required environment variables
	// Currently no-op since provider handles validation
}
