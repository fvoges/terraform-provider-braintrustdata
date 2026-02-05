package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExperimentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExperimentResourceConfig("test-experiment", "Test Experiment Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "test-experiment"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "description", "Test Experiment Description"),
					resource.TestCheckResourceAttrSet("braintrustdata_experiment.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_experiment.test", "project_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_experiment.test", "created"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "public", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_experiment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccExperimentResourceConfig("test-experiment-updated", "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "test-experiment-updated"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "description", "Updated Description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccExperimentResource_WithMetadataAndTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentResourceConfigWithMetadataAndTags(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "test-experiment-metadata"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "metadata.environment", "test"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "metadata.version", "1.0"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("braintrustdata_experiment.test", "tags.*", "ml"),
					resource.TestCheckTypeSetElemAttr("braintrustdata_experiment.test", "tags.*", "production"),
				),
			},
		},
	})
}

func TestAccExperimentResource_PublicToggle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentResourceConfigWithPublic("test-experiment-public", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "test-experiment-public"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "public", "true"),
				),
			},
			{
				Config: testAccExperimentResourceConfigWithPublic("test-experiment-public", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "test-experiment-public"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "public", "false"),
				),
			},
		},
	})
}

func TestAccExperimentResource_StatePersistence(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentResourceConfig("state-test", "State persistence test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name and description are persisted in state after create
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "state-test"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "description", "State persistence test"),
				),
			},
			{
				// Refresh state and verify no drift
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "state-test"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "description", "State persistence test"),
				),
			},
		},
	})
}

func TestAccExperimentResource_MetadataClearing(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create with metadata
				Config: testAccExperimentResourceConfigWithMetadataAndTags(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "metadata.environment", "test"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "metadata.version", "1.0"),
				),
			},
			{
				// Remove metadata to verify clearing works
				Config: testAccExperimentResourceConfigWithoutMetadata(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "name", "test-experiment-metadata"),
					resource.TestCheckResourceAttr("braintrustdata_experiment.test", "metadata.%", "0"),
				),
			},
		},
	})
}

func testAccExperimentResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-experiment"
}

resource "braintrustdata_experiment" "test" {
  project_id  = braintrustdata_project.test.id
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccExperimentResourceConfigWithMetadataAndTags() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-project-for-experiment"
}

resource "braintrustdata_experiment" "test" {
  project_id  = braintrustdata_project.test.id
  name        = "test-experiment-metadata"
  description = "Experiment with metadata and tags"

  metadata = {
    environment = "test"
    version     = "1.0"
  }

  tags = ["ml", "production"]
}
`
}

func testAccExperimentResourceConfigWithPublic(name string, public bool) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-experiment"
}

resource "braintrustdata_experiment" "test" {
  project_id = braintrustdata_project.test.id
  name       = %[1]q
  public     = %[2]t
}
`, name, public)
}

func testAccExperimentResourceConfigWithoutMetadata() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-project-for-experiment"
}

resource "braintrustdata_experiment" "test" {
  project_id  = braintrustdata_project.test.id
  name        = "test-experiment-metadata"
  description = "Experiment with metadata removed"
  tags        = ["ml", "production"]
}
`
}
