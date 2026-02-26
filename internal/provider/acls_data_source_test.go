package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccACLsDataSource_WithObjectFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccACLsDataSourceConfigWithObjectFilters(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_acls.test", "object_id", "braintrustdata_project.test", "id"),
					resource.TestCheckResourceAttr("data.braintrustdata_acls.test", "object_type", "project"),
					resource.TestCheckResourceAttr("data.braintrustdata_acls.test", "limit", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_acls.test", "acls.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_acls.test", "ids.#", "1"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_acls.test", "acls.0.object_id", "braintrustdata_project.test", "id"),
					resource.TestCheckResourceAttr("data.braintrustdata_acls.test", "acls.0.object_type", "project"),
				),
			},
		},
	})
}

func testAccACLsDataSourceConfigWithObjectFilters() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-acls-ds-project"
  description = "Project for ACLs data source testing"
}

resource "braintrustdata_acl" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  user_id     = "866a8a8a-fee9-4a5b-8278-12970de499c2"
  permission  = "read"
}

data "braintrustdata_acls" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  limit       = 1
  depends_on = [
    braintrustdata_acl.test,
  ]
}
`
}

func TestAccACLsDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccACLsDataSourceConfigInvalidPagination(),
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func testAccACLsDataSourceConfigInvalidPagination() string {
	return `
data "braintrustdata_acls" "test" {
  object_id      = "project-1"
  object_type    = "project"
  starting_after = "acl-1"
  ending_before  = "acl-2"
}
`
}

func TestAccACLsDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccACLsDataSourceConfigInvalidLimit(),
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}

func testAccACLsDataSourceConfigInvalidLimit() string {
	return `
data "braintrustdata_acls" "test" {
  object_id   = "project-1"
  object_type = "project"
  limit       = 0
}
`
}
