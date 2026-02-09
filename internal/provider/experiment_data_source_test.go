package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExperimentDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "name", "test-datasource-experiment"),
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "description", "Data source test experiment"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "created"),
				),
			},
		},
	})
}

func testAccExperimentDataSourceConfigByID() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-datasource-exp-project"
}

resource "braintrustdata_experiment" "test" {
  name        = "test-datasource-experiment"
  description = "Data source test experiment"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_experiment" "test" {
  id = braintrustdata_experiment.test.id
}
`
}

func TestAccExperimentDataSource_ByNameAndProject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentDataSourceConfigByNameAndProject(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "name", "test-datasource-byname"),
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "description", "Find by name test"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "created"),
				),
			},
		},
	})
}

func testAccExperimentDataSourceConfigByNameAndProject() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-datasource-exp-byname-proj"
}

resource "braintrustdata_experiment" "test" {
  name        = "test-datasource-byname"
  description = "Find by name test"
  project_id  = braintrustdata_project.test.id
}

data "braintrustdata_experiment" "test" {
  name       = braintrustdata_experiment.test.name
  project_id = braintrustdata_project.test.id
}
`
}

func TestAccExperimentDataSource_BothIDAndName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccExperimentDataSourceConfigBothIDAndName(),
				ExpectError: regexp.MustCompile("Cannot specify both 'id' and 'name'"),
			},
		},
	})
}

func testAccExperimentDataSourceConfigBothIDAndName() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-validation-project"
}

resource "braintrustdata_experiment" "test" {
  name       = "test-validation-exp"
  project_id = braintrustdata_project.test.id
}

data "braintrustdata_experiment" "test" {
  id         = braintrustdata_experiment.test.id
  name       = braintrustdata_experiment.test.name
  project_id = braintrustdata_project.test.id
}
`
}

func TestAccExperimentDataSource_NeitherIDNorName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccExperimentDataSourceConfigNeitherIDNorName(),
				ExpectError: regexp.MustCompile("Must specify either 'id' or 'name'"),
			},
		},
	})
}

func testAccExperimentDataSourceConfigNeitherIDNorName() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-validation-project-2"
}

data "braintrustdata_experiment" "test" {
  project_id = braintrustdata_project.test.id
}
`
}

func TestAccExperimentDataSource_AllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentDataSourceConfigAllAttributes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "name", "test-datasource-all-attrs"),
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "description", "Experiment with all attributes"),
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "public", "true"),
					resource.TestCheckResourceAttr("data.braintrustdata_experiment.test", "tags.#", "2"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "created"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_experiment.test", "metadata.%"),
				),
			},
		},
	})
}

func testAccExperimentDataSourceConfigAllAttributes() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-datasource-exp-all-attrs-proj"
}

resource "braintrustdata_experiment" "test" {
  name        = "test-datasource-all-attrs"
  description = "Experiment with all attributes"
  project_id  = braintrustdata_project.test.id
  public      = true
  tags        = ["test", "datasource"]
  metadata = {
    key1 = "value1"
    key2 = "value2"
  }
}

data "braintrustdata_experiment" "test" {
  name       = braintrustdata_experiment.test.name
  project_id = braintrustdata_project.test.id
}
`
}
