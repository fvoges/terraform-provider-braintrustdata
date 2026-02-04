package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig("test-project", "Test Project Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("braintrustdata_project.test", "description", "Test Project Description"),
					resource.TestCheckResourceAttrSet("braintrustdata_project.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_project.test", "org_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_project.test", "created"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccProjectResourceConfig("test-project", "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("braintrustdata_project.test", "description", "Updated Description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProjectResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}
