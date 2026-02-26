package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_role.test", "name", "test-role-ds-by-id"),
					resource.TestCheckResourceAttr("data.braintrustdata_role.test", "description", "Role data source lookup by id"),
					resource.TestCheckResourceAttr("data.braintrustdata_role.test", "member_permissions.#", "2"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_role.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_role.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_role.test", "created"),
				),
			},
		},
	})
}

func testAccRoleDataSourceConfigByID() string {
	return `
resource "braintrustdata_role" "test" {
  name               = "test-role-ds-by-id"
  description        = "Role data source lookup by id"
  member_permissions = ["read", "update"]
}

data "braintrustdata_role" "test" {
  id = braintrustdata_role.test.id
}
`
}

func TestAccRoleDataSource_ByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDataSourceConfigByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_role.test", "name", "test-role-ds-by-name"),
					resource.TestCheckResourceAttr("data.braintrustdata_role.test", "description", "Role data source lookup by name"),
					resource.TestCheckResourceAttr("data.braintrustdata_role.test", "member_permissions.#", "1"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_role.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_role.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_role.test", "created"),
				),
			},
		},
	})
}

func testAccRoleDataSourceConfigByName() string {
	return `
resource "braintrustdata_role" "test" {
  name               = "test-role-ds-by-name"
  description        = "Role data source lookup by name"
  member_permissions = ["read"]
}

data "braintrustdata_role" "test" {
  name = braintrustdata_role.test.name
}
`
}

func TestAccRoleDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoleDataSourceConfigMissingLookupAttributes(),
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func testAccRoleDataSourceConfigMissingLookupAttributes() string {
	return `
data "braintrustdata_role" "test" {}
`
}

func TestAccRoleDataSource_NotFound(t *testing.T) {
	missingName := "missing-role-ds-name-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoleDataSourceConfigNotFound(missingName),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No role found with name: %s", missingName)),
			},
		},
	})
}

func testAccRoleDataSourceConfigNotFound(missingName string) string {
	return fmt.Sprintf(`
data "braintrustdata_role" "test" {
  name = %q
}
`, missingName)
}
