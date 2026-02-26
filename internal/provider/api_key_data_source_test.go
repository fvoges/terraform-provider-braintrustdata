package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_api_key.test", "name", "test-api-key-ds-by-id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "preview_name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "created"),
				),
			},
		},
	})
}

func testAccAPIKeyDataSourceConfigByID() string {
	return `
resource "braintrustdata_api_key" "test" {
  name = "test-api-key-ds-by-id"
}

data "braintrustdata_api_key" "test" {
  id = braintrustdata_api_key.test.id
}
`
}

func TestAccAPIKeyDataSource_ByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyDataSourceConfigByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_api_key.test", "name", "test-api-key-ds-by-name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "preview_name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_api_key.test", "created"),
				),
			},
		},
	})
}

func testAccAPIKeyDataSourceConfigByName() string {
	return `
resource "braintrustdata_api_key" "test" {
  name = "test-api-key-ds-by-name"
}

data "braintrustdata_api_key" "test" {
  name = braintrustdata_api_key.test.name
}
`
}

func TestAccAPIKeyDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAPIKeyDataSourceConfigMissingLookupAttributes(),
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func testAccAPIKeyDataSourceConfigMissingLookupAttributes() string {
	return `
data "braintrustdata_api_key" "test" {}
`
}

func TestAccAPIKeyDataSource_NotFound(t *testing.T) {
	missingName := "missing-api-key-ds-name-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAPIKeyDataSourceConfigNotFound(missingName),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No API key found with name: %s", missingName)),
			},
		},
	})
}

func testAccAPIKeyDataSourceConfigNotFound(missingName string) string {
	return fmt.Sprintf(`
data "braintrustdata_api_key" "test" {
  name = %q
}
`, missingName)
}
