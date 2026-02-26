package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentVariableDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEnvironmentVariableDataSourceConfigMissingLookupAttributes(),
				ExpectError: regexp.MustCompile(`'name' must be provided and non-empty when using lookup mode`),
			},
		},
	})
}

func testAccEnvironmentVariableDataSourceConfigMissingLookupAttributes() string {
	return `
data "braintrustdata_environment_variable" "test" {}
`
}

func TestAccEnvironmentVariableDataSource_ConflictingAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEnvironmentVariableDataSourceConfigConflictingAttributes(),
				ExpectError: regexp.MustCompile(`Cannot combine 'id' with lookup attributes`),
			},
		},
	})
}

func testAccEnvironmentVariableDataSourceConfigConflictingAttributes() string {
	return `
data "braintrustdata_environment_variable" "test" {
  id          = "env-var-123"
  name        = "OPENAI_API_KEY"
  object_type = "project"
  object_id   = "project-123"
}
`
}
