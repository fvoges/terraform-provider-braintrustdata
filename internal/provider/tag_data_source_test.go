package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagDataSource_ByID(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	tagID := os.Getenv("BRAINTRUST_TAG_ID")
	if tagID == "" {
		t.Skip("BRAINTRUST_TAG_ID must be set for tag data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfigByID(tagID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.braintrustdata_tag.test", "id", tagID),
					resource.TestCheckResourceAttrSet("data.braintrustdata_tag.test", "name"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_tag.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.braintrustdata_tag.test", "user_id"),
				),
			},
		},
	})
}

func testAccTagDataSourceConfigByID(tagID string) string {
	return fmt.Sprintf(`
data "braintrustdata_tag" "test" {
  id = %[1]q
}
`, tagID)
}

func TestAccTagDataSource_ByName(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	tagID := os.Getenv("BRAINTRUST_TAG_ID")
	if tagID == "" {
		t.Skip("BRAINTRUST_TAG_ID must be set for tag data source acceptance testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfigByName(tagID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.braintrustdata_tag.by_name", "id", "data.braintrustdata_tag.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_tag.by_name", "name", "data.braintrustdata_tag.by_id", "name"),
					resource.TestCheckResourceAttrPair("data.braintrustdata_tag.by_name", "project_id", "data.braintrustdata_tag.by_id", "project_id"),
				),
			},
		},
	})
}

func testAccTagDataSourceConfigByName(tagID string) string {
	return fmt.Sprintf(`
data "braintrustdata_tag" "by_id" {
  id = %[1]q
}

data "braintrustdata_tag" "by_name" {
  name       = data.braintrustdata_tag.by_id.name
  project_id = data.braintrustdata_tag.by_id.project_id
}
`, tagID)
}

func TestAccTagDataSource_MissingLookupAttributes(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data "braintrustdata_tag" "test" {}`,
				ExpectError: regexp.MustCompile(`Must specify either 'id' or 'name'`),
			},
		},
	})
}

func TestAccTagDataSource_ConflictingLookupAttributes(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "braintrustdata_tag" "test" {
  id         = "tag-1"
  name       = "production"
  project_id = "proj-1"
}
`,
				ExpectError: regexp.MustCompile(`Cannot combine 'id' with searchable attributes`),
			},
		},
	})
}

func TestAccTagDataSource_NotFound(t *testing.T) {
	testAccTagDataSourceRequiresAPIKey(t)

	missingName := "missing-tag-ds-name-00000000"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "braintrustdata_tag" "test" {
  name       = %q
  project_id = %q
}
`, missingName, "00000000-0000-0000-0000-000000000000"),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No tag found with name: %s", missingName)),
			},
		},
	})
}

func testAccTagDataSourceRequiresAPIKey(t *testing.T) {
	t.Helper()

	if os.Getenv("BRAINTRUST_API_KEY") == "" {
		t.Skip("BRAINTRUST_API_KEY must be set for acceptance testing")
	}
}
