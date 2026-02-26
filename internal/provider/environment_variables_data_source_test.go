package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentVariablesDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEnvironmentVariablesDataSourceConfigInvalidPagination(),
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func testAccEnvironmentVariablesDataSourceConfigInvalidPagination() string {
	return `
data "braintrustdata_environment_variables" "test" {
  object_type    = "project"
  object_id      = "project-123"
  starting_after = "env-var-1"
  ending_before  = "env-var-2"
}
`
}

func TestAccEnvironmentVariablesDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEnvironmentVariablesDataSourceConfigInvalidLimit(),
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}

func testAccEnvironmentVariablesDataSourceConfigInvalidLimit() string {
	return `
data "braintrustdata_environment_variables" "test" {
  object_type = "project"
  object_id   = "project-123"
  limit       = 0
}
`
}
