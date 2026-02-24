package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectsDataSource_WithProjectNameFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectsDataSourceConfigWithProjectNameFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_projects.test", "project_name", "test-projects-ds-filtered"),
					resource.TestCheckResourceAttr("data.braintrustdata_projects.test", "projects.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_projects.test", "projects.0.name", "test-projects-ds-filtered"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_projects.test", "projects.0.id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_projects.test", "projects.0.org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_projects.test", "projects.0.created"),
					resource.TestCheckResourceAttr("data.braintrustdata_projects.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccProjectsDataSourceConfigWithProjectNameFilter() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-projects-ds-filtered"
  description = "Project for projects data source filter test"
}

data "braintrustdata_projects" "test" {
  project_name = braintrustdata_project.test.name
  depends_on = [
    braintrustdata_project.test,
  ]
}
`
}

func TestAccProjectsDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectsDataSourceConfigInvalidPagination(),
				ExpectError: regexp.MustCompile(`Cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func testAccProjectsDataSourceConfigInvalidPagination() string {
	return `
data "braintrustdata_projects" "test" {
  starting_after = "00000000-0000-0000-0000-000000000001"
  ending_before  = "00000000-0000-0000-0000-000000000002"
}
`
}

func TestAccProjectsDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectsDataSourceConfigInvalidLimit(),
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}

func testAccProjectsDataSourceConfigInvalidLimit() string {
	return `
data "braintrustdata_projects" "test" {
  limit = 0
}
`
}
