package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccACLDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccACLDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.braintrustdata_acl.test", "id"),
					resource.TestCheckResourceAttr("data.braintrustdata_acl.test", "object_type", "project"),
					resource.TestCheckResourceAttr("data.braintrustdata_acl.test", "permission", "read"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_acl.test", "object_id", "braintrustdata_project.test", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_acl.test", "id", "braintrustdata_acl.test", "id"),
				),
			},
		},
	})
}

func testAccACLDataSourceConfigByID() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-acl-ds-project"
  description = "Project for ACL data source testing"
}

resource "braintrustdata_acl" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  user_id     = "866a8a8a-fee9-4a5b-8278-12970de499c2"
  permission  = "read"
}

data "braintrustdata_acl" "test" {
  id = braintrustdata_acl.test.id
}
`
}

func TestAccACLDataSource_NotFound(t *testing.T) {
	missingID := "missing-acl-ds-id-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccACLDataSourceConfigNotFound(missingID),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No ACL found with ID: %s", missingID)),
			},
		},
	})
}

func testAccACLDataSourceConfigNotFound(missingID string) string {
	return fmt.Sprintf(`
data "braintrustdata_acl" "test" {
  id = %q
}
`, missingID)
}
