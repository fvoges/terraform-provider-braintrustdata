package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAISecretDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAISecretDataSourceConfigMissingLookupAttributes(),
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func testAccAISecretDataSourceConfigMissingLookupAttributes() string {
	return `
data "braintrustdata_ai_secret" "test" {}
`
}

func TestAccAISecretDataSource_ConflictingAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAISecretDataSourceConfigConflictingAttributes(),
				ExpectError: regexp.MustCompile(`Cannot combine 'id' with searchable attributes`),
			},
		},
	})
}

func testAccAISecretDataSourceConfigConflictingAttributes() string {
	return `
data "braintrustdata_ai_secret" "test" {
  id             = "ai-secret-123"
  name           = "PROVIDER_OPENAI_CREDENTIAL"
  org_name       = "example-org"
  ai_secret_type = "openai"
}
`
}
