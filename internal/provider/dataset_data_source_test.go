package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasetDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_dataset.test", "name", "test-datasource-dataset"),
					resource.TestCheckResourceAttr("data.braintrustdata_dataset.test", "description", "Data source test dataset"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_dataset.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_dataset.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_dataset.test", "created"),
				),
			},
		},
	})
}

func testAccDatasetDataSourceConfigByID() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-datasource-dataset-project"
}

resource "braintrustdata_dataset" "test" {
  name        = "test-datasource-dataset"
  description = "Data source test dataset"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_dataset" "test" {
  id = braintrustdata_dataset.test.id
}
`
}

func TestAccDatasetDataSource_ByNameAndProject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetDataSourceConfigByNameAndProject(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_dataset.test", "name", "test-datasource-byname"),
					resource.TestCheckResourceAttr("data.braintrustdata_dataset.test", "description", "Find by name test"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_dataset.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_dataset.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_dataset.test", "created"),
				),
			},
		},
	})
}

func testAccDatasetDataSourceConfigByNameAndProject() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-datasource-dataset-byname-proj"
}

resource "braintrustdata_dataset" "test" {
  name        = "test-datasource-byname"
  description = "Find by name test"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_dataset" "test" {
  name       = braintrustdata_dataset.test.name
  project_id = braintrustdata_project.test.id
}
`
}

func TestAccDatasetDataSource_NeitherIDNorName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasetDataSourceConfigNeitherIDNorName(),
				ExpectError: regexp.MustCompile(`Must specify either 'id' or both 'name' and 'project_id' to look up the\s+dataset\.`),
			},
		},
	})
}

func testAccDatasetDataSourceConfigNeitherIDNorName() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-validation-project-dataset"
}

data "braintrustdata_dataset" "test" {
  project_id = braintrustdata_project.test.id
}
`
}

func TestAccDatasetDataSource_NotFound(t *testing.T) {
	projectName := "test-notfound-project-dataset"
	datasetName := "non-existent-dataset"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasetDataSourceConfigNotFound(projectName, datasetName),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No dataset found with name: %s in project:", datasetName)),
			},
		},
	})
}

func testAccDatasetDataSourceConfigNotFound(projectName, datasetName string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "%s"
}

data "braintrustdata_dataset" "test" {
  name       = "%s"
  project_id = braintrustdata_project.test.id
}
`, projectName, datasetName)
}
