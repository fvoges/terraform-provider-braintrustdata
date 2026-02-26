package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRolesDataSource_WithRoleNameFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolesDataSourceConfigWithRoleNameFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_roles.test", "role_name", "test-roles-ds-filtered"),
					resource.TestCheckResourceAttr("data.braintrustdata_roles.test", "roles.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_roles.test", "roles.0.name", "test-roles-ds-filtered"),
					resource.TestCheckResourceAttr("data.braintrustdata_roles.test", "roles.0.member_permissions.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_roles.test", "roles.0.member_permissions.0", "read"),
					resource.TestCheckResourceAttr("data.braintrustdata_roles.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccRolesDataSourceConfigWithRoleNameFilter() string {
	return `
resource "braintrustdata_role" "other" {
  name               = "test-roles-ds-other"
  description        = "Role for roles data source filter test"
  member_permissions = ["update"]
}

resource "braintrustdata_role" "target" {
  name               = "test-roles-ds-filtered"
  description        = "Target role for roles data source filter test"
  member_permissions = ["read"]
}

data "braintrustdata_roles" "test" {
  role_name = braintrustdata_role.target.name
  depends_on = [
    braintrustdata_role.other,
    braintrustdata_role.target,
  ]
}
`
}

func TestAccRolesDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRolesDataSourceConfigInvalidPagination(),
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func testAccRolesDataSourceConfigInvalidPagination() string {
	return `
data "braintrustdata_roles" "test" {
  starting_after = "00000000-0000-0000-0000-000000000001"
  ending_before  = "00000000-0000-0000-0000-000000000002"
}
`
}

func TestAccRolesDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRolesDataSourceConfigInvalidLimit(),
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}

func testAccRolesDataSourceConfigInvalidLimit() string {
	return `
data "braintrustdata_roles" "test" {
  limit = 0
}
`
}
