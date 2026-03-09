package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccACLDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccACLDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.braintrustdata_acl.test", "id"),
					resource.TestCheckResourceAttr("data.braintrustdata_acl.test", "object_type", "project"),
					resource.TestCheckResourceAttr("data.braintrustdata_acl.test", "permission", "read"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_acl.test", "object_id", "braintrustdata_project.test", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_acl.test", "id", "braintrustdata_acl.test", "id"),
				),
			},
		},
	})
}

func testAccACLDataSourceConfigByID() string {
	return `
resource "braintrustdata_project" "test" {
  name        = "test-acl-ds-project"
  description = "Project for ACL data source testing"
}

resource "braintrustdata_acl" "test" {
  object_id   = braintrustdata_project.test.id
  object_type = "project"
  user_id     = "866a8a8a-fee9-4a5b-8278-12970de499c2"
  permission  = "read"
}

data "braintrustdata_acl" "test" {
  id = braintrustdata_acl.test.id
}
`
}

func TestAccACLDataSource_NotFound(t *testing.T) {
	missingID := "00000000-0000-0000-0000-000000000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccACLDataSourceConfigNotFound(missingID),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`(?is)(no\s+acl\s+found.*%[1]s|error\s+reading\s+acl.*(%[1]s.*(missing\s+read\s+access|does\s+not\s+exist)|(missing\s+read\s+access|does\s+not\s+exist).*(%[1]s)?|status\s*400|400\s+bad\s+request))`, regexp.QuoteMeta(missingID))),
			},
		},
	})
}

func TestAccACLDataSource_InvalidUUID(t *testing.T) {
	invalidID := "missing-acl-ds-id-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccACLDataSourceConfigNotFound(invalidID),
				ExpectError: regexp.MustCompile(`(?is)error\s+reading\s+acl.*(invalid\s+uuid|uuid.*invalid|status\s*400|400\s+bad\s+request)`),
			},
		},
	})
}

func testAccACLDataSourceConfigNotFound(missingID string) string {
	return fmt.Sprintf(`
data "braintrustdata_acl" "test" {
  id = %q
}
`, missingID)
}
