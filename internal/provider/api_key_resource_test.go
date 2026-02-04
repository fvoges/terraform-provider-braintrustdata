package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAPIKeyResourceConfig("test-api-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_api_key.test", "name", "test-api-key"),
					resource.TestCheckResourceAttrSet("braintrustdata_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_api_key.test", "org_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_api_key.test", "preview_name"),
					resource.TestCheckResourceAttrSet("braintrustdata_api_key.test", "created"),
					resource.TestCheckResourceAttrSet("braintrustdata_api_key.test", "key"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Key is only available at creation time, not on import
				ImportStateVerifyIgnore: []string{"key"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccAPIKeyResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "braintrustdata_api_key" "test" {
  name = %[1]q
}
`, name)
}
