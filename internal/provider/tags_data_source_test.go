package provider

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagsDataSource_WithTagNameAndProjectIDFilter(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	tagID := os.Getenv("BRAINTRUST_TAG_ID")
	if tagID == "" {
		t.Skip("BRAINTRUST_TAG_ID must be set for tags data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagsDataSourceConfigWithTagNameAndProjectIDFilter(tagID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_tags.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_tags.test", "tags.0.id", "data.braintrustdata_tag.seed", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_tags.test", "tags.0.name", "data.braintrustdata_tag.seed", "name"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_tags.test", "tags.0.project_id", "data.braintrustdata_tag.seed", "project_id"),
					resource.TestCheckResourceAttr("data.braintrustdata_tags.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccTagsDataSourceConfigWithTagNameAndProjectIDFilter(tagID string) string {
	return `
data "braintrustdata_tag" "seed" {
  id = "` + tagID + `"
}

data "braintrustdata_tags" "test" {
  project_id = data.braintrustdata_tag.seed.project_id
  tag_name   = data.braintrustdata_tag.seed.name
  depends_on = [data.braintrustdata_tag.seed]
}
`
}

func TestAccTagsDataSource_InvalidPagination(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_tags" "test" {
  starting_after = "tag-1"
  ending_before  = "tag-2"
}
`,
				ExpectError: regexp.MustCompile(`cannot specify both 'starting_after' and 'ending_before'`),
			},
		},
	})
}

func TestAccTagsDataSource_InvalidLimit(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_tags" "test" {
  limit = 0
}
`,
				ExpectError: regexp.MustCompile(`'limit' must be greater than or equal to 1`),
			},
		},
	})
}
