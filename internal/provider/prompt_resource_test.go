package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPromptResourceConfig("test-prompt", "Test Prompt Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "test-prompt"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "description", "Test Prompt Description"),
					resource.TestCheckResourceAttrSet("braintrustdata_prompt.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_prompt.test", "project_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_prompt.test", "created"),
					resource.TestCheckResourceAttrSet("braintrustdata_prompt.test", "slug"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "braintrustdata_prompt.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPromptResourceConfig("test-prompt-updated", "Updated Description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "test-prompt-updated"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "description", "Updated Description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPromptResource_WithTagsAndMetadata(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfigWithTagsAndMetadata(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "test-prompt-metadata"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "metadata.environment", "test"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "metadata.version", "1.0"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("braintrustdata_prompt.test", "tags.*", "ml"),
					resource.TestCheckTypeSetElemAttr("braintrustdata_prompt.test", "tags.*", "production"),
				),
			},
		},
	})
}

func TestAccPromptResource_WithPromptData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfigWithPromptData(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "test-prompt-data"),
					resource.TestCheckResourceAttrSet("braintrustdata_prompt.test", "prompt_data"),
				),
			},
			// Refresh and verify no perpetual diff
			{
				Config:   testAccPromptResourceConfigWithPromptData(),
				PlanOnly: true,
			},
		},
	})
}

func TestAccPromptResource_StatePersistence(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfig("state-test-prompt", "State persistence test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "state-test-prompt"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "description", "State persistence test"),
				),
			},
			{
				// Refresh state and verify no drift
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "state-test-prompt"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "description", "State persistence test"),
				),
			},
		},
	})
}

func testAccPromptResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-prompt"
}

resource "braintrustdata_prompt" "test" {
  project_id  = braintrustdata_project.test.id
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccPromptResourceConfigWithTagsAndMetadata() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-project-for-prompt"
}

resource "braintrustdata_prompt" "test" {
  project_id  = braintrustdata_project.test.id
  name        = "test-prompt-metadata"
  description = "Prompt with metadata and tags"

  metadata = {
    environment = "test"
    version     = "1.0"
  }

  tags = ["ml", "production"]
}
`
}

func testAccPromptResourceConfigWithPromptData() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-project-for-prompt"
}

resource "braintrustdata_prompt" "test" {
  project_id  = braintrustdata_project.test.id
  name        = "test-prompt-data"
  description = "Prompt with prompt_data"

  prompt_data = jsonencode({
    prompt = {
      type    = "completion"
      content = "You are a helpful assistant."
    }
    options = {
      model       = "gpt-4"
      temperature = 0.7
    }
  })
}
`
}
