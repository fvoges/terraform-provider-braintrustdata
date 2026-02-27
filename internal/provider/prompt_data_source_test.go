package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPromptDataSource_ByID(t *testing.T) {
	testAccPromptDataSourceRequiresAPIKey(t)

	promptID, _, ok := testAccPromptLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_PROMPT_ID and BRAINTRUST_PROMPT_PROJECT_ID must be set for prompt data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfigByID(promptID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_prompt.test", "id", promptID),
					resource.TestCheckResourceAttrSet("data.braintrustdata_prompt.test", "name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_prompt.test", "project_id"),
				),
			},
		},
	})
}

func testAccPromptDataSourceConfigByID(promptID string) string {
	return fmt.Sprintf(`
data "braintrustdata_prompt" "test" {
  id = %q
}
`, promptID)
}

func TestAccPromptDataSource_ByNameAndProject(t *testing.T) {
	testAccPromptDataSourceRequiresAPIKey(t)

	promptID, projectID, ok := testAccPromptLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_PROMPT_ID and BRAINTRUST_PROMPT_PROJECT_ID must be set for prompt data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfigByNameAndProject(promptID, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_prompt.by_name", "id", "data.braintrustdata_prompt.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_prompt.by_name", "name", "data.braintrustdata_prompt.by_id", "name"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_prompt.by_name", "project_id", "data.braintrustdata_prompt.by_id", "project_id"),
				),
			},
		},
	})
}

func testAccPromptDataSourceConfigByNameAndProject(promptID, projectID string) string {
	return fmt.Sprintf(`
data "braintrustdata_prompt" "by_id" {
  id = %q
}

data "braintrustdata_prompt" "by_name" {
  name       = data.braintrustdata_prompt.by_id.name
  project_id = %q
}
`, promptID, projectID)
}

func TestAccPromptDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_prompt" "test" {
  project_id = "project-1"
}
`,
				ExpectError: regexp.MustCompile(`Must specify either 'id' or both 'name' and 'project_id'`),
			},
		},
	})
}

func TestAccPromptDataSource_ConflictingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_prompt" "test" {
  id         = "prompt-1"
  name       = "support-agent"
  project_id = "project-1"
}
`,
				ExpectError: regexp.MustCompile(`Cannot combine 'id' with searchable attributes`),
			},
		},
	})
}

func testAccPromptLookupContext() (string, string, bool) {
	promptID := os.Getenv("BRAINTRUST_PROMPT_ID")
	projectID := os.Getenv("BRAINTRUST_PROMPT_PROJECT_ID")

	if promptID == "" || projectID == "" {
		return "", "", false
	}

	return promptID, projectID, true
}

func testAccPromptDataSourceRequiresAPIKey(t *testing.T) {
	t.Helper()

	if os.Getenv("BRAINTRUST_API_KEY") == "" {
		t.Skip("BRAINTRUST_API_KEY must be set for acceptance testing")
	}
}
