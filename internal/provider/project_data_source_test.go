package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_project.test", "name", "test-project-ds-by-id"),
					resource.TestCheckResourceAttr("data.braintrustdata_project.test", "description", "Project data source lookup by id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_project.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_project.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_project.test", "created"),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfigByID() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-project-ds-by-id"
  description = "Project data source lookup by id"
}

data "braintrustdata_project" "test" {
  id = braintrustdata_project.test.id
}
`
}

func TestAccProjectDataSource_ByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfigByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_project.test", "name", "test-project-ds-by-name"),
					resource.TestCheckResourceAttr("data.braintrustdata_project.test", "description", "Project data source lookup by name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_project.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_project.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_project.test", "created"),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfigByName() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-project-ds-by-name"
  description = "Project data source lookup by name"
}

data "braintrustdata_project" "test" {
  name = braintrustdata_project.test.name
}
`
}

func TestAccProjectDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectDataSourceConfigMissingLookupAttributes(),
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func testAccProjectDataSourceConfigMissingLookupAttributes() string {
	return `
data "braintrustdata_project" "test" {}
`
}

func TestAccProjectDataSource_NotFound(t *testing.T) {
	missingName := "missing-project-ds-name-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectDataSourceConfigNotFound(missingName),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No project found with name: %s", missingName)),
			},
		},
	})
}

func testAccProjectDataSourceConfigNotFound(missingName string) string {
	return fmt.Sprintf(`
data "braintrustdata_project" "test" {
  name = %q
}
`, missingName)
}
