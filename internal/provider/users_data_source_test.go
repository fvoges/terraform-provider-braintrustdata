package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.braintrustdata_users.test", "users.#"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_users.test", "ids.#"),
				),
			},
		},
	})
}

func testAccUsersDataSourceConfig() string {
	return `
data "braintrustdata_users" "test" {
  limit = 10
}
`
}

func TestAccUsersDataSource_WithEmailFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfigWithEmailFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_users.filtered", "users.#", "0"),
					resource.TestCheckResourceAttr("data.braintrustdata_users.filtered", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccUsersDataSourceConfigWithEmailFilter() string {
	return `
data "braintrustdata_users" "filtered" {
  email = "not-a-real-user+tf-provider-test@example.invalid"
}
`
}

func TestAccUsersDataSource_InvalidPagination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUsersDataSourceConfigInvalidPagination(),
				ExpectError: regexp.MustCompile(`Cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func testAccUsersDataSourceConfigInvalidPagination() string {
	return `
data "braintrustdata_users" "test" {
  starting_after = "00000000-0000-0000-0000-000000000001"
  ending_before  = "00000000-0000-0000-0000-000000000002"
}
`
}
