package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFunctionResource(t *testing.T) {
	testAccFunctionResourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionResourceConfig("test-function", "Initial function description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_function.test", "name", "test-function"),
					resource.TestCheckResourceAttr("braintrustdata_function.test", "description", "Initial function description"),
					resource.TestCheckResourceAttr("braintrustdata_function.test", "slug", "test-function"),
					resource.TestCheckResourceAttr("braintrustdata_function.test", "metadata.owner", "terraform"),
					resource.TestCheckResourceAttr("braintrustdata_function.test", "tags.#", "2"),
					resource.TestCheckResourceAttrSet("braintrustdata_function.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_function.test", "project_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_function.test", "created"),
				),
			},
			{
				ResourceName:      "braintrustdata_function.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionResourceConfig("test-function-updated", "Updated function description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_function.test", "name", "test-function-updated"),
					resource.TestCheckResourceAttr("braintrustdata_function.test", "description", "Updated function description"),
					resource.TestCheckResourceAttr("braintrustdata_function.test", "metadata.owner", "platform"),
					resource.TestCheckResourceAttr("braintrustdata_function.test", "tags.#", "2"),
				),
			},
		},
	})
}

func testAccFunctionResourceConfig(name, description string) string {
	metadataOwner := "terraform"
	model := "gpt-4o-mini"
	tags := `["acceptance", "function"]`

	if name == "test-function-updated" {
		metadataOwner = "platform"
		model = "gpt-4.1-mini"
		tags = `["acceptance", "updated"]`
	}

	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-function-resource"
}

resource "braintrustdata_function" "test" {
  project_id    = braintrustdata_project.test.id
  name          = %[1]q
  slug          = "test-function"
  description   = %[2]q
  function_data = jsonencode({
    type = "prompt"
  })
  prompt_data = jsonencode({
    prompt = {
      type = "chat"
      messages = [
        {
          role    = "system"
          content = "You are a helpful assistant."
        }
      ]
    }
    options = {
      model = %[3]q
    }
  })
  metadata = {
    owner = %[4]q
  }
  tags = %[5]s
}
`, name, description, model, metadataOwner, tags)
}

func testAccFunctionResourceRequiresAPIKey(t *testing.T) {
	t.Helper()

	if os.Getenv("BRAINTRUST_API_KEY") == "" {
		t.Skip("BRAINTRUST_API_KEY must be set for acceptance testing")
	}
}
