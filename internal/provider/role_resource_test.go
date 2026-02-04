package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	t.Skip("TODO: This test fails in CI with 'Missing API Key Configuration' error during planning, " +
		"despite API key being set in environment. The error occurs after other tests pass successfully. " +
		"Needs investigation into test isolation or CI environment configuration.")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRoleResourceConfig("test-role", "Test Role Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_role.test", "name", "test-role"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "description", "Test Role Description"),
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
				Config: testAccRoleResourceConfig("test-role", "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_role.test", "name", "test-role"),
					resource.TestCheckResourceAttr("braintrustdata_role.test", "description", "Updated Description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRoleResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "braintrustdata_role" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}
