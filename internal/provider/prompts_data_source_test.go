package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptsDataSource_WithNameFilter(t *testing.T) {
	testAccPromptDataSourceRequiresAPIKey(t)

	promptID, projectID, ok := testAccPromptLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_PROMPT_ID and BRAINTRUST_PROMPT_PROJECT_ID must be set for prompts data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptsDataSourceConfigWithNameFilter(promptID, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.braintrustdata_prompts.test", "prompts.#"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_prompts.test", "ids.#"),
				),
			},
		},
	})
}

func testAccPromptsDataSourceConfigWithNameFilter(promptID, projectID string) string {
	return fmt.Sprintf(`
data "braintrustdata_prompt" "seed" {
  id = %q
}

data "braintrustdata_prompts" "test" {
  project_id = %q
  name       = data.braintrustdata_prompt.seed.name
  depends_on = [data.braintrustdata_prompt.seed]
}
`, promptID, projectID)
}

func TestAccPromptsDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_prompts" "test" {
  project_id     = "project-1"
  starting_after = "prompt-1"
  ending_before  = "prompt-2"
}
`,
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func TestAccPromptsDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_prompts" "test" {
  project_id = "project-1"
  limit      = 0
}
`,
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}
