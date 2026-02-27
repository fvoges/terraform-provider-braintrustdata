package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccViewsDataSource_WithFilterIDsAndViewName(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	viewID, objectID, objectType, ok := testAccViewLookupContext()
	if !ok {
		t.Skip("BRAINTRUST_VIEW_ID, BRAINTRUST_VIEW_OBJECT_ID, and BRAINTRUST_VIEW_OBJECT_TYPE must be set for views data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccViewsDataSourceConfigWithFilterIDsAndViewName(viewID, objectID, objectType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_views.test", "views.#", "1"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_views.test", "views.0.id", "data.braintrustdata_view.seed", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_views.test", "views.0.name", "data.braintrustdata_view.seed", "name"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_views.test", "views.0.object_id", "data.braintrustdata_view.seed", "object_id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_views.test", "views.0.object_type", "data.braintrustdata_view.seed", "object_type"),
					resource.TestCheckResourceAttr("data.braintrustdata_views.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccViewsDataSourceConfigWithFilterIDsAndViewName(viewID, objectID, objectType string) string {
	return `
data "braintrustdata_view" "seed" {
  id          = "` + viewID + `"
  object_id   = "` + objectID + `"
  object_type = "` + objectType + `"
}

data "braintrustdata_views" "test" {
  object_id   = data.braintrustdata_view.seed.object_id
  object_type = data.braintrustdata_view.seed.object_type
  filter_ids  = [data.braintrustdata_view.seed.id]
  view_name   = data.braintrustdata_view.seed.name
  depends_on  = [data.braintrustdata_view.seed]
}
`
}

func TestAccViewsDataSource_InvalidPagination(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_views" "test" {
  object_id      = "project-1"
  object_type    = "project"
  starting_after = "view-1"
  ending_before  = "view-2"
}
`,
				ExpectError: regexp.MustCompile(`Cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func TestAccViewsDataSource_InvalidLimit(t *testing.T) {
	testAccViewDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_views" "test" {
  object_id   = "project-1"
  object_type = "project"
  limit       = 0
}
`,
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}
