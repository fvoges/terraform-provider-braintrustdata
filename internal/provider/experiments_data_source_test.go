package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExperimentsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiments.test", "experiments.#"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiments.test", "ids.#"),
				),
			},
		},
	})
}

func testAccExperimentsDataSourceConfig() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-experiments-list-project"
}

resource "braintrustdata_experiment" "test1" {
  name        = "test-experiments-list-1"
  description = "First test experiment"
  project_id  = braintrustdata_project.test.id
}

resource "braintrustdata_experiment" "test2" {
  name        = "test-experiments-list-2"
  description = "Second test experiment"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_experiments" "test" {
  project_id = braintrustdata_project.test.id
  depends_on = [
    braintrustdata_experiment.test1,
    braintrustdata_experiment.test2,
  ]
}
`
}

func TestAccExperimentsDataSource_WithFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentsDataSourceConfigWithFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_experiments.test", "experiments.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_experiments.test", "experiments.0.name", "filtered-experiment-1"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiments.test", "ids.#"),
				),
			},
		},
	})
}

func testAccExperimentsDataSourceConfigWithFilter() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-experiments-filter-project"
}

resource "braintrustdata_experiment" "test1" {
  name        = "filtered-experiment-1"
  description = "First filtered experiment"
  project_id  = braintrustdata_project.test.id
}

resource "braintrustdata_experiment" "test2" {
  name        = "filtered-experiment-2"
  description = "Second filtered experiment"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_experiments" "test" {
  project_id = braintrustdata_project.test.id
  name       = "filtered-experiment-1"
  depends_on = [
    braintrustdata_experiment.test1,
    braintrustdata_experiment.test2,
  ]
}
`
}
