package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentVariableResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentVariableResourceConfig("OPENAI_API_KEY", "initial-secret-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("braintrustdata_environment_variable.test", "id"),
					resource.TestCheckResourceAttr("braintrustdata_environment_variable.test", "name", "OPENAI_API_KEY"),
					resource.TestCheckResourceAttr("braintrustdata_environment_variable.test", "object_type", "project"),
					resource.TestCheckResourceAttr("braintrustdata_environment_variable.test", "value", "initial-secret-value"),
				),
			},
			{
				ResourceName:            "braintrustdata_environment_variable.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
			{
				Config: testAccEnvironmentVariableResourceConfig("OPENAI_API_KEY", "rotated-secret-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_environment_variable.test", "name", "OPENAI_API_KEY"),
					resource.TestCheckResourceAttr("braintrustdata_environment_variable.test", "value", "rotated-secret-value"),
				),
			},
		},
	})
}

func TestAccEnvironmentVariableResource_MissingValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEnvironmentVariableResourceConfigMissingValue(),
				ExpectError: regexp.MustCompile(`'value' must be provided and non-empty`),
			},
		},
	})
}

func testAccEnvironmentVariableResourceConfig(name, value string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name        = "test-env-var-project"
  description = "Project for environment variable resource testing"
}

resource "braintrustdata_environment_variable" "test" {
  object_type = "project"
  object_id   = braintrustdata_project.test.id
  name        = %[1]q
  value       = %[2]q
}
`, name, value)
}

func testAccEnvironmentVariableResourceConfigMissingValue() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-env-var-project-missing-value"
  description = "Project for environment variable resource validation testing"
}

resource "braintrustdata_environment_variable" "test" {
  object_type = "project"
  object_id   = braintrustdata_project.test.id
  name        = "OPENAI_API_KEY"
}
`
}
