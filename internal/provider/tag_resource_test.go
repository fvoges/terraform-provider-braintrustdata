package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfig("test-tag", "Initial tag description", "#ff0000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "name", "test-tag"),
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "description", "Initial tag description"),
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "color", "#ff0000"),
					resource.TestCheckResourceAttrSet("braintrustdata_tag.test", "id"),
					resource.TestCheckResourceAttrSet("braintrustdata_tag.test", "project_id"),
					resource.TestCheckResourceAttrSet("braintrustdata_tag.test", "created"),
				),
			},
			{
				Config: testAccTagResourceConfig("test-tag-updated", "Updated tag description", "#00ff00"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "name", "test-tag-updated"),
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "description", "Updated tag description"),
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "color", "#00ff00"),
				),
			},
			{
				ResourceName:      "braintrustdata_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTagResource_NullDescriptionPreservesState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfig("test-tag-clear-description", "Description set", "#112233"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "description", "Description set"),
				),
			},
			{
				Config: testAccTagResourceConfigClearDescription("test-tag-clear-description", "#112233"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "description", "Description set"),
					resource.TestCheckResourceAttr("braintrustdata_tag.test", "color", "#112233"),
				),
			},
		},
	})
}

func testAccTagResourceConfig(name, description, color string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-tag-resource"
}

resource "braintrustdata_tag" "test" {
  project_id  = braintrustdata_project.test.id
  name        = %[1]q
  description = %[2]q
  color       = %[3]q
}
`, name, description, color)
}

func testAccTagResourceConfigClearDescription(name, color string) string {
	return fmt.Sprintf(`
resource "braintrustdata_project" "test" {
  name = "test-project-for-tag-resource"
}

resource "braintrustdata_tag" "test" {
  project_id  = braintrustdata_project.test.id
  name        = %[1]q
  description = null
  color       = %[2]q
}
`, name, color)
}
