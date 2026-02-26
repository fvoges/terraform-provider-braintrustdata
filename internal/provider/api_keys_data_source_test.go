package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeysDataSource_WithAPIKeyNameFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeysDataSourceConfigWithAPIKeyNameFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_api_keys.test", "api_key_name", "test-api-keys-ds-filtered"),
					resource.TestCheckResourceAttr("data.braintrustdata_api_keys.test", "api_keys.#", "1"),
					resource.TestCheckResourceAttr("data.braintrustdata_api_keys.test", "api_keys.0.name", "test-api-keys-ds-filtered"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_keys.test", "api_keys.0.id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_keys.test", "api_keys.0.org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_keys.test", "api_keys.0.created"),
					resource.TestCheckResourceAttr("data.braintrustdata_api_keys.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccAPIKeysDataSourceConfigWithAPIKeyNameFilter() string {
	return `
resource "braintrustdata_api_key" "other" {
  name = "test-api-keys-ds-other"
}

resource "braintrustdata_api_key" "target" {
  name = "test-api-keys-ds-filtered"
}

data "braintrustdata_api_keys" "test" {
  api_key_name = braintrustdata_api_key.target.name
  depends_on = [
    braintrustdata_api_key.other,
    braintrustdata_api_key.target,
  ]
}
`
}

func TestAccAPIKeysDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAPIKeysDataSourceConfigInvalidPagination(),
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func testAccAPIKeysDataSourceConfigInvalidPagination() string {
	return `
data "braintrustdata_api_keys" "test" {
  starting_after = "00000000-0000-0000-0000-000000000001"
  ending_before  = "00000000-0000-0000-0000-000000000002"
}
`
}

func TestAccAPIKeysDataSource_InvalidLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAPIKeysDataSourceConfigInvalidLimit(),
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}

func testAccAPIKeysDataSourceConfigInvalidLimit() string {
	return `
data "braintrustdata_api_keys" "test" {
  limit = 0
}
`
}
