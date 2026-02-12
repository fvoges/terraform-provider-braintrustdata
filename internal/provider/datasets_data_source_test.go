package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasetsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_datasets.test", "datasets.#", "2"),
					resource.TestCheckResourceAttr("data.braintrustdata_datasets.test", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccDatasetsDataSourceConfig() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-datasets-list-project"
}

resource "braintrustdata_dataset" "test1" {
  name        = "test-datasets-list-1"
  description = "First test dataset"
  project_id  = braintrustdata_project.test.id
}

resource "braintrustdata_dataset" "test2" {
  name        = "test-datasets-list-2"
  description = "Second test dataset"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_datasets" "test" {
  project_id = braintrustdata_project.test.id
  depends_on = [
    braintrustdata_dataset.test1,
    braintrustdata_dataset.test2,
  ]
}
`
}

func TestAccDatasetsDataSource_WithFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetsDataSourceConfigWithFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_datasets.test", "datasets.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_datasets.test", "datasets.0.name", "filtered-dataset-1"),
					resource.TestCheckResourceAttr("data.braintrustdata_datasets.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccDatasetsDataSourceConfigWithFilter() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-datasets-filter-project"
}

resource "braintrustdata_dataset" "test1" {
  name        = "filtered-dataset-1"
  description = "First filtered dataset"
  project_id  = braintrustdata_project.test.id
}

resource "braintrustdata_dataset" "test2" {
  name        = "filtered-dataset-2"
  description = "Second filtered dataset"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_datasets" "test" {
  project_id = braintrustdata_project.test.id
  name       = "filtered-dataset-1"
  depends_on = [
    braintrustdata_dataset.test1,
    braintrustdata_dataset.test2,
  ]
}
`
}
