package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgsDataSource_List(t *testing.T) {
	orgID := os.Getenv("BRAINTRUST_ORG_ID")
	if orgID == "" {
		t.Skip("BRAINTRUST_ORG_ID must be set for organizations data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgsDataSourceConfigList(orgID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_orgs.test", "limit", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_orgs.test", "orgs.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_orgs.test", "ids.#", "1"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_orgs.test", "orgs.0.id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_orgs.test", "orgs.0.name"),
				),
			},
		},
	})
}

func testAccOrgsDataSourceConfigList(orgID string) string {
	return fmt.Sprintf(`
data "braintrustdata_org" "current" {
  id = %[1]q
}

data "braintrustdata_orgs" "test" {
  org_name = data.braintrustdata_org.current.name
  limit    = 1
}
`, orgID)
}

func TestAccOrgsDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_orgs" "test" {
  starting_after = "org-1"
  ending_before  = "org-2"
}
`,
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func TestAccOrgsDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_orgs" "test" {
  limit = 0
}
`,
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}
