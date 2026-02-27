package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAISecretsDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAISecretsDataSourceConfigInvalidPagination(),
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func testAccAISecretsDataSourceConfigInvalidPagination() string {
	return `
data "braintrustdata_ai_secrets" "test" {
  starting_after = "ai-secret-1"
  ending_before  = "ai-secret-2"
}
`
}

func TestAccAISecretsDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAISecretsDataSourceConfigInvalidLimit(),
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}

func testAccAISecretsDataSourceConfigInvalidLimit() string {
	return `
data "braintrustdata_ai_secrets" "test" {
  limit = 0
}
`
}
