package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupDataSource_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_group.test", "name", "test-datasource-group"),
					resource.TestCheckResourceAttr("data.braintrustdata_group.test", "description", "Data source test group"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_group.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_group.test", "created"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfigByID() string {
	return `
resource "braintrustdata_group" "test" {
  name        = "test-datasource-group"
  description = "Data source test group"
}

data "braintrustdata_group" "test" {
  id = braintrustdata_group.test.id
}
`
}

func TestAccGroupDataSource_ByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfigByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_group.test", "name", "test-datasource-byname"),
					resource.TestCheckResourceAttr("data.braintrustdata_group.test", "description", "Find by name test"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_group.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_group.test", "created"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfigByName() string {
	return `
resource "braintrustdata_group" "test" {
  name        = "test-datasource-byname"
  description = "Find by name test"
}

data "braintrustdata_group" "test" {
  name = braintrustdata_group.test.name
}
`
}

func TestAccGroupDataSource_AllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfigAllAttributes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_group.test", "name", "test-datasource-all-attrs"),
					resource.TestCheckResourceAttr("data.braintrustdata_group.test", "description", "Group with all attributes"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_group.test", "id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_group.test", "org_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_group.test", "created"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfigAllAttributes() string {
	return `
resource "braintrustdata_group" "test" {
  name        = "test-datasource-all-attrs"
  description = "Group with all attributes"
}

data "braintrustdata_group" "test" {
  name = braintrustdata_group.test.name
}
`
}
