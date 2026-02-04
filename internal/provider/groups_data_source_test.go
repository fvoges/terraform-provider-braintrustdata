package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Should have groups listed
					resource.TestCheckResourceAttrSet("data.braintrustdata_groups.test", "groups.#"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_groups.test", "ids.#"),
				),
			},
		},
	})
}

func testAccGroupsDataSourceConfig() string {
	return `
resource "braintrustdata_group" "test1" {
  name        = "test-groups-list-1"
  description = "First test group"
}

resource "braintrustdata_group" "test2" {
  name        = "test-groups-list-2"
  description = "Second test group"
}

data "braintrustdata_groups" "test" {
  depends_on = [
    braintrustdata_group.test1,
    braintrustdata_group.test2,
  ]
}
`
}
