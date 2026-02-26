package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgResource(t *testing.T) {
	orgID := os.Getenv("BRAINTRUST_ORG_ID")
	if orgID == "" {
		t.Skip("BRAINTRUST_ORG_ID must be set for org resource acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgResourceConfig(orgID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_org.test", "org_id", orgID),
					resource.TestCheckResourceAttr("braintrustdata_org.test", "id", orgID),
					resource.TestCheckResourceAttrSet("braintrustdata_org.test", "name"),
				),
			},
			{
				ResourceName:      "braintrustdata_org.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccOrgResourceConfig(orgID string) string {
	return fmt.Sprintf(`
resource "braintrustdata_org" "test" {
  org_id = %[1]q
}
`, orgID)
}
