package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatasetResourceConfig("test-dataset", "Test Dataset Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "test-dataset"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "description", "Test Dataset Description"),
					resource.TestCheckResourceAttrSet("braintrustdata_dataset.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_dataset.test", "project_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_dataset.test", "created"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "public", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_dataset.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDatasetResourceConfig("test-dataset-updated", "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "test-dataset-updated"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "description", "Updated Description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDatasetResource_WithMetadataAndTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetResourceConfigWithMetadataAndTags(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "test-dataset-metadata"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "metadata.environment", "test"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "metadata.version", "1.0"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("braintrustdata_dataset.test", "tags.*", "ml"),
					resource.TestCheckTypeSetElemAttr("braintrustdata_dataset.test", "tags.*", "production"),
				),
			},
		},
	})
}

func TestAccDatasetResource_PublicToggle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetResourceConfigWithPublic("test-dataset-public", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "test-dataset-public"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "public", "true"),
				),
			},
			{
				Config: testAccDatasetResourceConfigWithPublic("test-dataset-public", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "test-dataset-public"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "public", "false"),
				),
			},
		},
	})
}

func TestAccDatasetResource_StatePersistence(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetResourceConfig("state-test", "State persistence test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name and description are persisted in state after create
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "state-test"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "description", "State persistence test"),
				),
			},
			{
				// Refresh state and verify no drift
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "state-test"),
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "description", "State persistence test"),
				),
			},
		},
	})
}

func TestAccDatasetResource_ProjectIDChange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetResourceConfigForProject("test-dataset", "project1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "test-dataset"),
					resource.TestCheckResourceAttrPair("braintrustdata_dataset.test", "project_id", "braintrustdata_project.test1", "id"),
				),
			},
			{
				Config: testAccDatasetResourceConfigForProject("test-dataset", "project2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_dataset.test", "name", "test-dataset"),
					resource.TestCheckResourceAttrPair("braintrustdata_dataset.test", "project_id", "braintrustdata_project.test2", "id"),
				),
			},
		},
	})
}

func testAccDatasetResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-dataset"
}

resource "braintrustdata_dataset" "test" {
  project_id  = braintrustdata_project.test.id
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccDatasetResourceConfigWithMetadataAndTags() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-project-for-dataset"
}

resource "braintrustdata_dataset" "test" {
  project_id  = braintrustdata_project.test.id
  name        = "test-dataset-metadata"
  description = "Dataset with metadata and tags"

  metadata = {
    environment = "test"
    version     = "1.0"
  }

  tags = ["ml", "production"]
}
`
}

func testAccDatasetResourceConfigWithPublic(name string, public bool) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-dataset"
}

resource "braintrustdata_dataset" "test" {
  project_id = braintrustdata_project.test.id
  name       = %[1]q
  public     = %[2]t
}
`, name, public)
}

func testAccDatasetResourceConfigForProject(name, projectName string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test1" {
  name = "test-project-1"
}

resource "braintrustdata_project" "test2" {
  name = "test-project-2"
}

resource "braintrustdata_dataset" "test" {
  project_id  = braintrustdata_project.%[2]s.id
  name        = %[1]q
  description = "Dataset for project change test"
}
`, name, projectName)
}
