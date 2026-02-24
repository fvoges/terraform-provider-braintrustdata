package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.braintrustdata_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_user.test", "email"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfigByID() string {
	return `
data "braintrustdata_users" "seed" {
  limit = 1
}

data "braintrustdata_user" "test" {
  id = data.braintrustdata_users.seed.ids[0]
}
`
}

func TestAccUserDataSource_MissingLookupAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfigMissingLookupAttributes(),
				ExpectError: regexp.MustCompile(`Must specify either 'id' or at least one searchable attribute \('email', 'given_name', 'family_name'\)`),
			},
		},
	})
}

func testAccUserDataSourceConfigMissingLookupAttributes() string {
	return `
data "braintrustdata_user" "test" {}
`
}

func TestAccUserDataSource_NotFound(t *testing.T) {
	missingID := "00000000-0000-0000-0000-000000000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfigNotFound(missingID),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Could not read user ID %s", missingID)),
			},
		},
	})
}

func testAccUserDataSourceConfigNotFound(missingID string) string {
	return fmt.Sprintf(`
data "braintrustdata_user" "test" {
  id = %q
}
`, missingID)
}
