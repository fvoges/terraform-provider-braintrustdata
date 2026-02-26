package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgDataSource_ByID(t *testing.T) {
	orgID := os.Getenv("BRAINTRUST_ORG_ID")
	if orgID == "" {
		t.Skip("BRAINTRUST_ORG_ID must be set for organization data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgDataSourceConfigByID(orgID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_org.test", "id", orgID),
					resource.TestCheckResourceAttr("data.braintrustdata_org.test", "org_id", orgID),
					resource.TestCheckResourceAttrSet("data.braintrustdata_org.test", "name"),
				),
			},
		},
	})
}

func testAccOrgDataSourceConfigByID(orgID string) string {
	return fmt.Sprintf(`
data "braintrustdata_org" "test" {
  id = %[1]q
}
`, orgID)
}

func TestAccOrgDataSource_ByName(t *testing.T) {
	orgID := os.Getenv("BRAINTRUST_ORG_ID")
	if orgID == "" {
		t.Skip("BRAINTRUST_ORG_ID must be set for organization data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgDataSourceConfigByName(orgID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_org.by_name", "id", "data.braintrustdata_org.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_org.by_name", "name", "data.braintrustdata_org.by_id", "name"),
				),
			},
		},
	})
}

func testAccOrgDataSourceConfigByName(orgID string) string {
	return fmt.Sprintf(`
data "braintrustdata_org" "by_id" {
  id = %[1]q
}

data "braintrustdata_org" "by_name" {
  name = data.braintrustdata_org.by_id.name
}
`, orgID)
}

func TestAccOrgDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data "braintrustdata_org" "test" {}`,
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func TestAccOrgDataSource_NotFound(t *testing.T) {
	missingName := "missing-org-ds-name-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "braintrustdata_org" "test" {
  name = %q
}
`, missingName),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No organization found with name: %s", missingName)),
			},
		},
	})
}
