package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

// TestAccPromptResource_EmptyTagsNoDiff verifies that explicitly setting
// `tags = []` does not produce a perpetual plan diff on subsequent refreshes.
func TestAccPromptResource_EmptyTagsNoDiff(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfigWithEmptyTags(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "test-prompt-empty-tags"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "tags.#", "0"),
				),
			},
			// Refresh and verify no perpetual diff when tags = [].
			{
				Config:   testAccPromptResourceConfigWithEmptyTags(),
				PlanOnly: true,
			},
		},
	})
}

// TestAccPromptResource_SlugDerivedFromName verifies that when no slug is
// provided, the provider derives one from the name and the resource is created
// successfully with a non-empty slug in state.
func TestAccPromptResource_SlugDerivedFromName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfigNoSlug("My Test Prompt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "My Test Prompt"),
					resource.TestCheckResourceAttrSet("braintrustdata_prompt.test", "slug"),
				),
			},
		},
	})
}

// TestAccPromptResource_WithExplicitSlug verifies that when an explicit slug is
// provided, that exact slug is used and stored in state.
func TestAccPromptResource_WithExplicitSlug(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptResourceConfigWithSlug("My Test Prompt", "my-custom-slug"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "My Test Prompt"),
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "slug", "my-custom-slug"),
				),
			},
		},
	})
}

// TestAccPromptResource_RequiresReplaceOnProjectIDChange verifies that
// changing project_id destroys and recreates the prompt (RequiresReplace
// plan modifier) rather than attempting an in-place update.
func TestAccPromptResource_RequiresReplaceOnProjectIDChange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create the prompt under project A.
			{
				Config: testAccPromptResourceConfigOnProjectA("test-prompt-replace"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_prompt.test", "name", "test-prompt-replace"),
					resource.TestCheckResourceAttrSet("braintrustdata_prompt.test", "project_id"),
				),
			},
			// Step 2: switch prompt to project B — project_id changes → replacement required.
			{
				Config: testAccPromptResourceConfigOnProjectB("test-prompt-replace"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"braintrustdata_prompt.test",
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
			},
		},
	})
}

func testAccPromptResourceConfigOnProjectA(promptName string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "proj_a" {
  name = "test-project-a-for-prompt-replace"
}

resource "braintrustdata_project" "proj_b" {
  name = "test-project-b-for-prompt-replace"
}

resource "braintrustdata_prompt" "test" {
  project_id = braintrustdata_project.proj_a.id
  name       = %[1]q
}
`, promptName)
}

func testAccPromptResourceConfigOnProjectB(promptName string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "proj_a" {
  name = "test-project-a-for-prompt-replace"
}

resource "braintrustdata_project" "proj_b" {
  name = "test-project-b-for-prompt-replace"
}

resource "braintrustdata_prompt" "test" {
  project_id = braintrustdata_project.proj_b.id
  name       = %[1]q
}
`, promptName)
}

func testAccPromptResourceConfigWithEmptyTags() string {
	return `
resource "braintrustdata_project" "test" {
  name = "test-project-for-prompt"
}

resource "braintrustdata_prompt" "test" {
  project_id  = braintrustdata_project.test.id
  name        = "test-prompt-empty-tags"
  description = "Prompt with explicitly empty tags"

  tags = []
}
`
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

func testAccPromptResourceConfigNoSlug(name string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-prompt"
}

resource "braintrustdata_prompt" "test" {
  project_id = braintrustdata_project.test.id
  name       = %[1]q
}
`, name)
}

func testAccPromptResourceConfigWithSlug(name, slug string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-prompt"
}

resource "braintrustdata_prompt" "test" {
  project_id = braintrustdata_project.test.id
  name       = %[1]q
  slug       = %[2]q
}
`, name, slug)
}
